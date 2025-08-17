package graph_pruner

import (
	"context"
	"fmt"
	"strings"

	"mgds/internal/pkg/database"
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

type graphPrunerService struct {
	graph    graph.Graph
	database database.Database
}

func NewGraphPrunerService(graph graph.Graph, database database.Database) GraphPrunerService {
	return &graphPrunerService{
		graph:    graph,
		database: database,
	}
}

func (s *graphPrunerService) DeleteNode(ctx context.Context, nodeID string) (*graph.DeletionResult, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}

	// Check if node exists and get its details
	targetNode, err := s.graph.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	result := &graph.DeletionResult{
		DeletedNodeIDs: []string{},
		Errors:         []string{},
	}

	// Delete associated document from MongoDB if it has a location
	if targetNode.Location != "" {
		err := s.database.Delete(ctx, targetNode.Location, nodeID)
		if err != nil && err != database.ErrNotFound {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete document for node %s: %v", nodeID, err))
		}
	}

	// Delete the node from the graph (this also removes all relationships)
	err = s.graph.DeleteNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete node from graph: %w", err)
	}

	result.DeletedNodes = 1
	result.DeletedNodeIDs = append(result.DeletedNodeIDs, nodeID)

	// Note: DeletedRelations count is not accurately tracked in this implementation
	// as Neo4j's DETACH DELETE doesn't return the count of deleted relationships
	// This could be improved by counting relationships before deletion

	return result, nil
}

func (s *graphPrunerService) DeleteNodeCascade(ctx context.Context, nodeID string, maxDepth int) (*graph.DeletionResult, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}

	if maxDepth <= 0 {
		maxDepth = 10 // Default safe limit
	}

	// Get all descendant nodes
	descendants, err := s.graph.GetDescendantNodes(ctx, nodeID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendant nodes: %w", err)
	}

	result := &graph.DeletionResult{
		DeletedNodeIDs: []string{},
		Errors:         []string{},
	}

	// Delete descendants first (bottom-up approach)
	for _, descendant := range descendants {
		descendantResult, err := s.DeleteNode(ctx, descendant.ID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete descendant %s: %v", descendant.ID, err))
		} else {
			result.DeletedNodes += descendantResult.DeletedNodes
			result.DeletedNodeIDs = append(result.DeletedNodeIDs, descendantResult.DeletedNodeIDs...)
		}
	}

	// Finally delete the root node
	rootResult, err := s.DeleteNode(ctx, nodeID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete root node %s: %v", nodeID, err))
	} else {
		result.DeletedNodes += rootResult.DeletedNodes
		result.DeletedNodeIDs = append(result.DeletedNodeIDs, rootResult.DeletedNodeIDs...)
	}

	return result, nil
}

func (s *graphPrunerService) DeleteRelation(ctx context.Context, sourceID, targetID, relationType string) error {
	if strings.TrimSpace(sourceID) == "" {
		return fmt.Errorf("sourceID cannot be empty")
	}
	if strings.TrimSpace(targetID) == "" {
		return fmt.Errorf("targetID cannot be empty")
	}
	if strings.TrimSpace(relationType) == "" {
		return fmt.Errorf("relationType cannot be empty")
	}

	return s.graph.DeleteRelation(ctx, sourceID, targetID, relationType)
}

func (s *graphPrunerService) PreviewDeletion(ctx context.Context, nodeID string, cascade bool, maxDepth int) (*DeletionPreview, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}

	// Get the target node
	targetNode, err := s.graph.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target node: %w", err)
	}

	preview := &DeletionPreview{
		TargetNode:     targetNode,
		AffectedNodes:  []*node.Node{},
		TotalNodes:     1,
		TotalRelations: 0,
		HasDocuments:   targetNode.Location != "",
	}

	// Get relations count for the target node
	relations, err := s.graph.GetNodeRelations(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node relations: %w", err)
	}
	preview.TotalRelations = len(relations)

	if cascade {
		if maxDepth <= 0 {
			maxDepth = 10
		}

		// Get all descendants
		descendants, err := s.graph.GetDescendantNodes(ctx, nodeID, maxDepth)
		if err != nil {
			return nil, fmt.Errorf("failed to get descendant nodes: %w", err)
		}

		preview.AffectedNodes = descendants
		preview.TotalNodes += len(descendants)

		// Count relations for descendants
		for _, descendant := range descendants {
			descendantRelations, err := s.graph.GetNodeRelations(ctx, descendant.ID)
			if err == nil {
				preview.TotalRelations += len(descendantRelations)
			}

			if descendant.Location != "" {
				preview.HasDocuments = true
			}
		}

		// Add warning for large cascades
		if len(descendants) > 50 {
			preview.Warning = fmt.Sprintf("Warning: This will delete %d nodes. Consider using a smaller maxDepth or reviewing the deletion scope.", preview.TotalNodes)
		}
	}

	return preview, nil
}

func (s *graphPrunerService) Close() error {
	if s.database != nil {
		return s.database.Close(context.Background())
	}
	return nil
}