package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"mgds/src/pkg/node"
	"mgds/src/pkg/object"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/quad"

	// Import LevelDB backend
	_ "github.com/cayleygraph/cayley/graph/kv/leveldb"
)

// CayleyGraph implements the Interface using Cayley graph database
type CayleyGraph struct {
	store  *cayley.Handle
	config *Config
}

// NewCayleyGraph creates a new Cayley graph instance
func NewCayleyGraph(config *Config) Interface {
	return &CayleyGraph{
		config: config,
	}
}

// Connect initializes connection to Cayley database
func (c *CayleyGraph) Connect(ctx context.Context) error {
	var store *cayley.Handle
	var err error

	switch c.config.DatabasePath {
	case "", ":memory:":
		store, err = cayley.NewMemoryGraph()
	default:
		// Try different backend names that might be available
		store, err = cayley.NewGraph("kv", c.config.DatabasePath, nil)
		if err != nil {
			// Fallback to leveldb
			store, err = cayley.NewGraph("leveldb", c.config.DatabasePath, nil)
		}
		if err != nil {
			// Last fallback to bolt
			store, err = cayley.NewGraph("bolt", c.config.DatabasePath, nil)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to create Cayley store: %w", err)
	}

	c.store = store
	return nil
}

// Close closes the database connection
func (c *CayleyGraph) Close() error {
	if c.store != nil {
		return c.store.Close()
	}
	return nil
}

// Ping checks database health
func (c *CayleyGraph) Ping(ctx context.Context) error {
	if c.store == nil {
		return fmt.Errorf("database not connected")
	}
	return nil
}

// AddNode adds a node to the graph
func (c *CayleyGraph) AddNode(ctx context.Context, n node.Interface) error {
	if c.store == nil {
		return fmt.Errorf("database not connected")
	}

	nodeData, err := c.nodeToJSON(n)
	if err != nil {
		return fmt.Errorf("failed to serialize node: %w", err)
	}

	quads := []quad.Quad{
		quad.Make(n.GetMgdsId(), "mgds:id", n.GetMgdsId(), nil),
		quad.Make(n.GetMgdsId(), "mgds:displayName", n.GetDisplayName(), nil),
		quad.Make(n.GetMgdsId(), "mgds:description", n.GetDescription(), nil),
		quad.Make(n.GetMgdsId(), "mgds:type", n.GetType(), nil),
		quad.Make(n.GetMgdsId(), "mgds:data", nodeData, nil),
	}

	return c.store.AddQuadSet(quads)
}

// GetNode retrieves a node by its mgdsId
func (c *CayleyGraph) GetNode(ctx context.Context, mgdsId string) (node.Interface, error) {
	if c.store == nil {
		return nil, fmt.Errorf("database not connected")
	}

	path := cayley.StartPath(c.store, quad.String(mgdsId)).Out(quad.String("mgds:data"))

	var nodeData string
	err := path.Iterate(context.TODO()).EachValue(nil, func(value quad.Value) {
		nodeData = quad.NativeOf(value).(string)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query node: %w", err)
	}

	if nodeData == "" {
		return nil, fmt.Errorf("node not found: %s", mgdsId)
	}

	return c.jsonToNode(nodeData)
}

// UpdateNode updates an existing node
func (c *CayleyGraph) UpdateNode(ctx context.Context, n node.Interface) error {
	if c.store == nil {
		return fmt.Errorf("database not connected")
	}

	// Remove existing node data
	err := c.DeleteNode(ctx, n.GetMgdsId())
	if err != nil {
		return fmt.Errorf("failed to delete existing node: %w", err)
	}

	// Add updated node
	return c.AddNode(ctx, n)
}

// DeleteNode removes a node and its relations
func (c *CayleyGraph) DeleteNode(ctx context.Context, mgdsId string) error {
	if c.store == nil {
		return fmt.Errorf("database not connected")
	}

	// Remove all quads where this node is the subject
	quadsToDelete := []quad.Quad{
		quad.Make(mgdsId, "mgds:id", mgdsId, nil),
		quad.Make(mgdsId, "mgds:displayName", "", nil),
		quad.Make(mgdsId, "mgds:description", "", nil),
		quad.Make(mgdsId, "mgds:type", "", nil),
		quad.Make(mgdsId, "mgds:data", "", nil),
	}

	// Try to remove core quads - some might not exist, that's ok
	for _, q := range quadsToDelete {
		c.store.RemoveQuad(q) // Ignore errors as some quads might not exist
	}

	return nil
}

// AddRelation adds a relation between two nodes
func (c *CayleyGraph) AddRelation(ctx context.Context, relation *Relation) error {
	if c.store == nil {
		return fmt.Errorf("database not connected")
	}

	relQuad := quad.Make(
		relation.FromNodeId,
		"rel:"+relation.Label,
		relation.ToNodeId,
		nil,
	)

	quads := []quad.Quad{relQuad}

	// Add relation properties if any
	if len(relation.Properties) > 0 {
		relId := fmt.Sprintf("rel:%s:%s:%s", relation.FromNodeId, relation.Label, relation.ToNodeId)
		for key, value := range relation.Properties {
			propQuad := quad.Make(relId, "prop:"+key, value, nil)
			quads = append(quads, propQuad)
		}
	}

	return c.store.AddQuadSet(quads)
}

// GetRelations gets all outgoing relations for a node
func (c *CayleyGraph) GetRelations(ctx context.Context, fromNodeId string) ([]*Relation, error) {
	if c.store == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var relations []*Relation

	// Get all outgoing predicates for this node
	predicatesPath := cayley.StartPath(c.store, quad.String(fromNodeId)).OutPredicates()

	err := predicatesPath.Iterate(context.TODO()).EachValue(nil, func(predVal quad.Value) {
		predStr := quad.NativeOf(predVal).(string)

		// Only process relation predicates
		if len(predStr) > 4 && predStr[:4] == "rel:" {
			label := predStr[4:] // Remove "rel:" prefix

			// Get all objects for this predicate
			objectsPath := cayley.StartPath(c.store, quad.String(fromNodeId)).Out(predVal)
			objectsPath.Iterate(context.TODO()).EachValue(nil, func(objVal quad.Value) {
				toNodeId := quad.NativeOf(objVal).(string)

				relation := &Relation{
					FromNodeId: fromNodeId,
					ToNodeId:   toNodeId,
					Label:      label,
					Properties: make(map[string]any),
				}
				relations = append(relations, relation)
			})
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get relations: %w", err)
	}

	return relations, nil
}

// GetIncomingRelations gets all incoming relations for a node
func (c *CayleyGraph) GetIncomingRelations(ctx context.Context, toNodeId string) ([]*Relation, error) {
	if c.store == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var relations []*Relation

	// Get all incoming predicates for this node
	predicatesPath := cayley.StartPath(c.store, quad.String(toNodeId)).InPredicates()

	err := predicatesPath.Iterate(context.TODO()).EachValue(nil, func(predVal quad.Value) {
		predStr := quad.NativeOf(predVal).(string)

		// Only process relation predicates
		if len(predStr) > 4 && predStr[:4] == "rel:" {
			label := predStr[4:] // Remove "rel:" prefix

			// Get all subjects for this predicate pointing to our toNodeId
			subjectsPath := cayley.StartPath(c.store, quad.String(toNodeId)).In(predVal)
			subjectsPath.Iterate(context.TODO()).EachValue(nil, func(subjVal quad.Value) {
				fromNodeId := quad.NativeOf(subjVal).(string)

				relation := &Relation{
					FromNodeId: fromNodeId,
					ToNodeId:   toNodeId,
					Label:      label,
					Properties: make(map[string]any),
				}
				relations = append(relations, relation)
			})
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get incoming relations: %w", err)
	}

	return relations, nil
}

// DeleteRelation removes a specific relation
func (c *CayleyGraph) DeleteRelation(ctx context.Context, fromNodeId, toNodeId, label string) error {
	if c.store == nil {
		return fmt.Errorf("database not connected")
	}

	relQuad := quad.Make(fromNodeId, "rel:"+label, toNodeId, nil)
	return c.store.RemoveQuad(relQuad)
}

// GetConnectedNodes gets all nodes connected to a given node
func (c *CayleyGraph) GetConnectedNodes(ctx context.Context, mgdsId string) ([]node.Interface, error) {
	if c.store == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var connectedIds []string

	// Get outgoing connections
	outRelations, err := c.GetRelations(ctx, mgdsId)
	if err != nil {
		return nil, err
	}
	for _, rel := range outRelations {
		connectedIds = append(connectedIds, rel.ToNodeId)
	}

	// Get incoming connections
	inRelations, err := c.GetIncomingRelations(ctx, mgdsId)
	if err != nil {
		return nil, err
	}
	for _, rel := range inRelations {
		connectedIds = append(connectedIds, rel.FromNodeId)
	}

	// Remove duplicates and get nodes
	seen := make(map[string]bool)
	var nodes []node.Interface

	for _, id := range connectedIds {
		if !seen[id] && id != mgdsId {
			seen[id] = true
			n, err := c.GetNode(ctx, id)
			if err == nil {
				nodes = append(nodes, n)
			}
		}
	}

	return nodes, nil
}

// FindPath finds a path between two nodes (simple implementation)
func (c *CayleyGraph) FindPath(ctx context.Context, fromNodeId, toNodeId string) ([]*Relation, error) {
	if c.store == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Simple implementation: check direct connection first
	relations, err := c.GetRelations(ctx, fromNodeId)
	if err != nil {
		return nil, err
	}

	for _, rel := range relations {
		if rel.ToNodeId == toNodeId {
			return []*Relation{rel}, nil
		}
	}

	// For now, return empty if no direct path found
	// TODO: Implement proper path finding algorithm
	return []*Relation{}, nil
}

// FindNodesByType finds all nodes of a specific type
func (c *CayleyGraph) FindNodesByType(ctx context.Context, nodeType string) ([]node.Interface, error) {
	if c.store == nil {
		return nil, fmt.Errorf("database not connected")
	}

	path := cayley.StartPath(c.store).Has(quad.String("mgds:type"), quad.String(nodeType))

	var nodes []node.Interface
	err := path.Iterate(context.TODO()).EachValue(nil, func(value quad.Value) {
		mgdsId := quad.NativeOf(value).(string)
		n, err := c.GetNode(ctx, mgdsId)
		if err == nil {
			nodes = append(nodes, n)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find nodes by type: %w", err)
	}

	return nodes, nil
}

// FindNodesByProperty finds nodes with specific property value
func (c *CayleyGraph) FindNodesByProperty(ctx context.Context, key string, value any) ([]node.Interface, error) {
	if c.store == nil {
		return nil, fmt.Errorf("database not connected")
	}

	path := cayley.StartPath(c.store).Has(quad.String("prop:"+key), quad.String(fmt.Sprintf("%v", value)))

	var nodes []node.Interface
	err := path.Iterate(context.TODO()).EachValue(nil, func(val quad.Value) {
		mgdsId := quad.NativeOf(val).(string)
		n, err := c.GetNode(ctx, mgdsId)
		if err == nil {
			nodes = append(nodes, n)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find nodes by property: %w", err)
	}

	return nodes, nil
}

// NodeExists checks if a node exists
func (c *CayleyGraph) NodeExists(ctx context.Context, mgdsId string) (bool, error) {
	if c.store == nil {
		return false, fmt.Errorf("database not connected")
	}

	path := cayley.StartPath(c.store, quad.String(mgdsId)).Out(quad.String("mgds:id"))

	var exists bool
	err := path.Iterate(context.TODO()).EachValue(nil, func(value quad.Value) {
		exists = true
	})

	return exists, err
}

// GetAllNodes returns all nodes in the graph
func (c *CayleyGraph) GetAllNodes(ctx context.Context) ([]node.Interface, error) {
	if c.store == nil {
		return nil, fmt.Errorf("database not connected")
	}

	path := cayley.StartPath(c.store).Has(quad.String("mgds:type"))

	var nodes []node.Interface
	err := path.Iterate(context.TODO()).EachValue(nil, func(value quad.Value) {
		mgdsId := quad.NativeOf(value).(string)
		n, err := c.GetNode(ctx, mgdsId)
		if err == nil {
			nodes = append(nodes, n)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all nodes: %w", err)
	}

	return nodes, nil
}

// Helper methods for serialization
func (c *CayleyGraph) nodeToJSON(n node.Interface) (string, error) {
	data, err := n.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize node: %w", err)
	}

	bytes, err := json.Marshal(data)
	return string(bytes), err
}

func (c *CayleyGraph) jsonToNode(data string) (node.Interface, error) {
	var nodeData map[string]any
	err := json.Unmarshal([]byte(data), &nodeData)
	if err != nil {
		return nil, err
	}

	nodeType, ok := nodeData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node type")
	}

	switch nodeType {
	case object.URLNodeType:
		// Create empty URLNode and deserialize into it
		n := &object.URLNode{}
		err := n.Deserialize(nodeData)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize URLNode: %w", err)
		}
		return n, nil
	default:
		return nil, fmt.Errorf("unsupported node type: %s", nodeType)
	}
}
