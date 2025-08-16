package graph_explorer

import (
	"context"
	"fmt"

	"mgds/internal/pkg/graph"
)

type graphExplorerService struct {
	graph graph.Graph
}

func NewGraphExplorerService(g graph.Graph) GraphExplorerService {
	return &graphExplorerService{
		graph: g,
	}
}

func (s *graphExplorerService) GetGraphOverview(ctx context.Context) (*GraphOverview, error) {
	stats, err := s.graph.GetGraphStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph stats: %w", err)
	}

	// Extract available types from the stats
	var availableTypes []string
	for nodeType := range stats.NodesByType {
		availableTypes = append(availableTypes, nodeType)
	}

	overview := &GraphOverview{
		Stats:          stats,
		AvailableTypes: availableTypes,
		TotalNodes:     stats.TotalNodes,
		TotalRelations: stats.TotalRelations,
	}

	return overview, nil
}

func (s *graphExplorerService) GetNodesByType(ctx context.Context, nodeType string, offset, limit int) (*graph.PaginatedNodesResult, error) {
	if nodeType == "" {
		return nil, fmt.Errorf("nodeType cannot be empty")
	}

	if offset < 0 {
		offset = 0
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	result, err := s.graph.GetNodesByType(ctx, nodeType, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes by type: %w", err)
	}

	return result, nil
}

func (s *graphExplorerService) GetNodeRelations(ctx context.Context, nodeID string) ([]*graph.Relation, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}

	relations, err := s.graph.GetNodeRelations(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node relations: %w", err)
	}

	return relations, nil
}

func (s *graphExplorerService) GetConnectedNodes(ctx context.Context, nodeID string) ([]*graph.NodeConnection, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("nodeID cannot be empty")
	}

	connections, err := s.graph.GetConnectedNodes(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connected nodes: %w", err)
	}

	return connections, nil
}

func (s *graphExplorerService) Close() error {
	if s.graph != nil {
		return s.graph.Close(context.Background())
	}
	return nil
}
