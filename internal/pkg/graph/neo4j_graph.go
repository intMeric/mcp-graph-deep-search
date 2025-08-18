package graph

import (
	"context"
	"fmt"
	"mgds/internal/pkg/node"
	"strings"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type DriverWithContext interface {
	VerifyConnectivity(ctx context.Context) error
	NewSession(ctx context.Context, config neo4j.SessionConfig) neo4j.SessionWithContext
	Close(ctx context.Context) error
}

type Neo4jGraph struct {
	driver DriverWithContext
	mu     sync.RWMutex
	config *Neo4jConfig
}

type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

func NewNeo4jGraph(config *Neo4jConfig) (Graph, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	auth := neo4j.BasicAuth(config.Username, config.Password, "")

	driver, err := neo4j.NewDriverWithContext(
		config.URI,
		auth,
		func(config *neo4j.Config) {
			config.MaxConnectionLifetime = 5 * time.Minute
			config.MaxConnectionPoolSize = 50
			config.ConnectionAcquisitionTimeout = 2 * time.Minute
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	graph := &Neo4jGraph{
		driver: driver,
		config: config,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := graph.verifyConnectivity(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to verify connectivity: %w", err)
	}

	return graph, nil
}

func (g *Neo4jGraph) verifyConnectivity(ctx context.Context) error {
	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	return driver.VerifyConnectivity(ctx)
}

func (g *Neo4jGraph) ensureConnection(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if err := g.driver.VerifyConnectivity(ctx); err != nil {
		g.driver.Close(ctx)

		auth := neo4j.BasicAuth(g.config.Username, g.config.Password, "")

		driver, newErr := neo4j.NewDriverWithContext(
			g.config.URI,
			auth,
			func(config *neo4j.Config) {
				config.MaxConnectionLifetime = 5 * time.Minute
				config.MaxConnectionPoolSize = 50
				config.ConnectionAcquisitionTimeout = 2 * time.Minute
			},
		)
		if newErr != nil {
			return fmt.Errorf("failed to reconnect to Neo4j: %w", newErr)
		}

		g.driver = driver

		if verifyErr := g.driver.VerifyConnectivity(ctx); verifyErr != nil {
			return fmt.Errorf("failed to verify reconnection: %w", verifyErr)
		}
	}

	return nil
}

func (g *Neo4jGraph) CreateNode(ctx context.Context, node *node.Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	if err := node.Validate(); err != nil {
		return fmt.Errorf("invalid node: %w", err)
	}

	if err := g.ensureConnection(ctx); err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := fmt.Sprintf(`
		MERGE (n:%s {id: $id})
		ON CREATE SET n.displayName = $displayName,
		              n.location = $location,
		              n.created_at = datetime()
		ON MATCH SET n.displayName = $displayName,
		             n.location = $location,
		             n.updated_at = datetime()
		RETURN n
	`, node.Type)

	parameters := map[string]any{
		"id":          node.ID,
		"displayName": node.DisplayName,
		"isDocumentAvailable":    node.IsDocumentAvailable,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	return nil
}

func (g *Neo4jGraph) CreateRelation(ctx context.Context, relation *Relation) error {
	if relation == nil {
		return fmt.Errorf("relation cannot be nil")
	}

	if err := relation.Validate(); err != nil {
		return fmt.Errorf("invalid relation: %w", err)
	}

	if err := g.ensureConnection(ctx); err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := fmt.Sprintf(`
		MATCH (source {id: $sourceId})
		MATCH (target {id: $targetId})
		MERGE (source)-[r:%s]->(target)
		ON CREATE SET r.created_at = datetime()
		ON MATCH SET r.updated_at = datetime()
		RETURN r
	`, relation.Type)

	parameters := map[string]any{
		"sourceId": relation.SourceID,
		"targetId": relation.TargetID,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}

		return result.Consume(ctx)
	})

	if err != nil {
		return fmt.Errorf("failed to create relation: %w", err)
	}

	return nil
}

func (g *Neo4jGraph) CreateLink(ctx context.Context, sourceNode *node.Node, targetNode *node.Node, relationType string) (*Relation, error) {
	if sourceNode == nil {
		return nil, fmt.Errorf("sourceNode cannot be nil")
	}
	if targetNode == nil {
		return nil, fmt.Errorf("targetNode cannot be nil")
	}
	if err := sourceNode.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sourceNode: %w", err)
	}
	if err := targetNode.Validate(); err != nil {
		return nil, fmt.Errorf("invalid targetNode: %w", err)
	}
	if strings.TrimSpace(relationType) == "" {
		return nil, fmt.Errorf("relation type cannot be empty")
	}

	// Create source node if it doesn't exist
	if err := g.CreateNode(ctx, sourceNode); err != nil {
		return nil, fmt.Errorf("failed to create source node: %w", err)
	}

	// Create target node if it doesn't exist
	if err := g.CreateNode(ctx, targetNode); err != nil {
		return nil, fmt.Errorf("failed to create target node: %w", err)
	}

	// Create the relation
	relation := &Relation{
		Type:     relationType,
		SourceID: sourceNode.ID,
		TargetID: targetNode.ID,
	}

	if err := g.CreateRelation(ctx, relation); err != nil {
		return nil, fmt.Errorf("failed to create relation: %w", err)
	}

	return relation, nil
}

func (g *Neo4jGraph) GetNode(ctx context.Context, id string) (*node.Node, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	if err := g.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH (n {id: $id})
		RETURN n.id as id, n.displayName as displayName, n.location as location, labels(n) as labels
	`

	parameters := map[string]any{
		"id": id,
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		return record, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	record := result.(*neo4j.Record)

	nodeID, _ := record.Get("id")
	displayName, _ := record.Get("displayName")
	location, _ := record.Get("location")
	labels, _ := record.Get("labels")

	labelsList := labels.([]any)
	if len(labelsList) == 0 {
		return nil, fmt.Errorf("node has no labels")
	}

	nodeType := labelsList[0].(string)

	node := &node.Node{
		Type:                nodeType,
		DisplayName:         displayName.(string),
		ID:                  nodeID.(string),
		IsDocumentAvailable: location.(string) != "",
	}

	return node, nil
}

func (g *Neo4jGraph) NodeExists(ctx context.Context, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, fmt.Errorf("id cannot be empty")
	}

	if err := g.ensureConnection(ctx); err != nil {
		return false, fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH (n {id: $id})
		RETURN count(n) > 0 as exists
	`

	parameters := map[string]any{
		"id": id,
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		exists, _ := record.Get("exists")
		return exists.(bool), nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to check node existence: %w", err)
	}

	return result.(bool), nil
}

func (g *Neo4jGraph) Close(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.driver != nil {
		return g.driver.Close(ctx)
	}
	return nil
}

func (g *Neo4jGraph) GetGraphStats(ctx context.Context) (*GraphStats, error) {
	if err := g.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := `
		CALL {
			MATCH (n) 
			WITH labels(n)[0] as nodeType, count(n) as nodeCount
			RETURN collect({type: nodeType, count: nodeCount}) as nodeStats, sum(nodeCount) as totalNodes
		}
		CALL {
			MATCH ()-[r]->() 
			WITH type(r) as relationType, count(r) as relationCount
			RETURN collect({type: relationType, count: relationCount}) as relationStats, sum(relationCount) as totalRelations
		}
		RETURN nodeStats, totalNodes, relationStats, totalRelations
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		return record, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get graph stats: %w", err)
	}

	record := result.(*neo4j.Record)

	totalNodes, _ := record.Get("totalNodes")
	nodeStats, _ := record.Get("nodeStats")
	totalRelations, _ := record.Get("totalRelations")
	relationStats, _ := record.Get("relationStats")

	stats := &GraphStats{
		TotalNodes:      int(totalNodes.(int64)),
		NodesByType:     make(map[string]int),
		TotalRelations:  int(totalRelations.(int64)),
		RelationsByType: make(map[string]int),
	}

	if nodeStatsList, ok := nodeStats.([]any); ok {
		for _, stat := range nodeStatsList {
			if statMap, ok := stat.(map[string]any); ok {
				nodeType := statMap["type"].(string)
				count := int(statMap["count"].(int64))
				stats.NodesByType[nodeType] = count
			}
		}
	}

	if relationStatsList, ok := relationStats.([]any); ok {
		for _, stat := range relationStatsList {
			if statMap, ok := stat.(map[string]any); ok {
				relationType := statMap["type"].(string)
				count := int(statMap["count"].(int64))
				stats.RelationsByType[relationType] = count
			}
		}
	}

	return stats, nil
}

func (g *Neo4jGraph) GetNodesByType(ctx context.Context, nodeType string, offset, limit int) (*PaginatedNodesResult, error) {
	if strings.TrimSpace(nodeType) == "" {
		return nil, fmt.Errorf("nodeType cannot be empty")
	}

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}

	if err := g.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	countQuery := fmt.Sprintf(`MATCH (n:%s) RETURN count(n) as total`, nodeType)
	nodesQuery := fmt.Sprintf(`
		MATCH (n:%s) 
		RETURN n.id as id, n.displayName as displayName, n.location as location
		ORDER BY n.id
		SKIP $offset LIMIT $limit
	`, nodeType)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Get total count
		countResult, err := tx.Run(ctx, countQuery, nil)
		if err != nil {
			return nil, err
		}

		countRecord, err := countResult.Single(ctx)
		if err != nil {
			return nil, err
		}

		total, _ := countRecord.Get("total")

		// Get nodes
		nodesResult, err := tx.Run(ctx, nodesQuery, map[string]any{
			"offset": offset,
			"limit":  limit,
		})
		if err != nil {
			return nil, err
		}

		var nodes []*node.Node
		for nodesResult.Next(ctx) {
			record := nodesResult.Record()
			nodeID, _ := record.Get("id")
			displayName, _ := record.Get("displayName")
			location, _ := record.Get("location")

			n := &node.Node{
				Type:                nodeType,
				ID:                  nodeID.(string),
				DisplayName:         displayName.(string),
				IsDocumentAvailable: location.(string) != "",
			}
			nodes = append(nodes, n)
		}

		return map[string]any{
			"total": total,
			"nodes": nodes,
		}, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get nodes by type: %w", err)
	}

	resultMap := result.(map[string]any)
	total := int(resultMap["total"].(int64))
	nodes := resultMap["nodes"].([]*node.Node)

	return &PaginatedNodesResult{
		Nodes:  nodes,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}

func (g *Neo4jGraph) GetNodeRelations(ctx context.Context, nodeID string) ([]*Relation, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}

	if err := g.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH (n {id: $nodeId})
		OPTIONAL MATCH (n)-[r]-(m)
		RETURN type(r) as relationType, startNode(r).id as sourceId, endNode(r).id as targetId
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any{
			"nodeId": nodeID,
		})
		if err != nil {
			return nil, err
		}

		var relations []*Relation
		for result.Next(ctx) {
			record := result.Record()
			relationType, hasType := record.Get("relationType")
			sourceID, hasSource := record.Get("sourceId")
			targetID, hasTarget := record.Get("targetId")

			if hasType && hasSource && hasTarget {
				relation := &Relation{
					Type:     relationType.(string),
					SourceID: sourceID.(string),
					TargetID: targetID.(string),
				}
				relations = append(relations, relation)
			}
		}

		return relations, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get node relations: %w", err)
	}

	return result.([]*Relation), nil
}

func (g *Neo4jGraph) GetConnectedNodes(ctx context.Context, nodeID string) ([]*NodeConnection, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}

	if err := g.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH (source {id: $nodeId})-[r]-(target)
		RETURN 
			target.id as targetId, 
			target.displayName as targetDisplayName, 
			target.location as targetLocation,
			labels(target) as targetLabels,
			type(r) as relationType,
			startNode(r).id = $nodeId as isSource,
			startNode(r).id as sourceId,
			endNode(r).id as targetRelId
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any{
			"nodeId": nodeID,
		})
		if err != nil {
			return nil, err
		}

		var connections []*NodeConnection
		for result.Next(ctx) {
			record := result.Record()

			targetID, _ := record.Get("targetId")
			targetDisplayName, _ := record.Get("targetDisplayName")
			targetLocation, _ := record.Get("targetLocation")
			targetLabels, _ := record.Get("targetLabels")
			relationType, _ := record.Get("relationType")
			isSource, _ := record.Get("isSource")
			sourceID, _ := record.Get("sourceId")
			targetRelID, _ := record.Get("targetRelId")

			// Get the first label as the node type
			labelsList := targetLabels.([]any)
			var nodeType string
			if len(labelsList) > 0 {
				nodeType = labelsList[0].(string)
			}

			targetNode := &node.Node{
				Type:                nodeType,
				ID:                  targetID.(string),
				DisplayName:         targetDisplayName.(string),
				IsDocumentAvailable: targetLocation.(string) != "",
			}

			relation := &Relation{
				Type:     relationType.(string),
				SourceID: sourceID.(string),
				TargetID: targetRelID.(string),
			}

			connection := &NodeConnection{
				Node:     targetNode,
				Relation: relation,
				IsSource: isSource.(bool),
			}

			connections = append(connections, connection)
		}

		return connections, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get connected nodes: %w", err)
	}

	return result.([]*NodeConnection), nil
}

// DeleteNode removes a node and all its relationships from the graph
func (g *Neo4jGraph) DeleteNode(ctx context.Context, nodeID string) error {
	if strings.TrimSpace(nodeID) == "" {
		return fmt.Errorf("nodeID cannot be empty")
	}

	if err := g.ensureConnection(ctx); err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH (n {id: $nodeId})
		DETACH DELETE n
		RETURN count(n) as deletedCount
	`

	parameters := map[string]any{
		"nodeId": nodeID,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			deletedCount := result.Record().Values[0].(int64)
			if deletedCount == 0 {
				return nil, fmt.Errorf("node with ID '%s' not found", nodeID)
			}
		}

		return nil, result.Err()
	})

	if err != nil {
		return fmt.Errorf("failed to delete node %s: %w", nodeID, err)
	}

	return nil
}

// DeleteRelation removes a specific relationship between two nodes
func (g *Neo4jGraph) DeleteRelation(ctx context.Context, sourceID, targetID, relationType string) error {
	if strings.TrimSpace(sourceID) == "" {
		return fmt.Errorf("sourceID cannot be empty")
	}
	if strings.TrimSpace(targetID) == "" {
		return fmt.Errorf("targetID cannot be empty")
	}
	if strings.TrimSpace(relationType) == "" {
		return fmt.Errorf("relationType cannot be empty")
	}

	if err := g.ensureConnection(ctx); err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := fmt.Sprintf(`
		MATCH (source {id: $sourceId})-[r:%s]->(target {id: $targetId})
		DELETE r
		RETURN count(r) as deletedCount
	`, relationType)

	parameters := map[string]any{
		"sourceId": sourceID,
		"targetId": targetID,
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			deletedCount := result.Record().Values[0].(int64)
			if deletedCount == 0 {
				return nil, fmt.Errorf("relation '%s' between '%s' and '%s' not found", relationType, sourceID, targetID)
			}
		}

		return nil, result.Err()
	})

	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}

	return nil
}

// GetDescendantNodes returns all descendant nodes up to maxDepth levels
func (g *Neo4jGraph) GetDescendantNodes(ctx context.Context, nodeID string, maxDepth int) ([]*node.Node, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}
	if maxDepth <= 0 {
		maxDepth = 10 // Default max depth to prevent infinite traversal
	}

	if err := g.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	g.mu.RLock()
	driver := g.driver
	g.mu.RUnlock()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: g.config.Database,
	})
	defer session.Close(ctx)

	query := `
		MATCH path = (start {id: $nodeId})-[*1..%d]->(descendant)
		RETURN DISTINCT descendant.id as id, 
		       labels(descendant)[0] as type,
		       descendant.displayName as displayName,
		       descendant.location as location
		ORDER BY length(path), descendant.id
	`
	query = fmt.Sprintf(query, maxDepth)

	parameters := map[string]any{
		"nodeId": nodeID,
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}

		var descendants []*node.Node
		for result.Next(ctx) {
			record := result.Record()
			id := record.Values[0].(string)
			nodeType := record.Values[1].(string)
			displayName := ""
			location := ""

			if record.Values[2] != nil {
				displayName = record.Values[2].(string)
			}
			if record.Values[3] != nil {
				location = record.Values[3].(string)
			}

			descendant := &node.Node{
				ID:                  id,
				Type:                nodeType,
				DisplayName:         displayName,
				IsDocumentAvailable: location != "",
			}

			descendants = append(descendants, descendant)
		}

		return descendants, result.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get descendant nodes: %w", err)
	}

	return result.([]*node.Node), nil
}
