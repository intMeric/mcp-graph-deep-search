package graph_explorer

import (
	"context"

	"mgds/internal/pkg/graph"
)

type GraphExplorerService interface {
	GetGraphOverview(ctx context.Context) (*GraphOverview, error)
	GetNodesByType(ctx context.Context, nodeType string, offset, limit int) (*graph.PaginatedNodesResult, error)
	GetNodeRelations(ctx context.Context, nodeID string) ([]*graph.Relation, error)
	GetConnectedNodes(ctx context.Context, nodeID string) ([]*graph.NodeConnection, error)
	Close() error
}

type GraphOverview struct {
	Stats          *graph.GraphStats `json:"stats"`
	AvailableTypes []string          `json:"availableTypes"`
	TotalNodes     int               `json:"totalNodes"`
	TotalRelations int               `json:"totalRelations"`
}
