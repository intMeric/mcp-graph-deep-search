package explore_graph

import (
	"context"

	"mgds/internal/app/services/graph_explorer"
	"mgds/internal/pkg/graph"
)

type ExploreGraphUseCase interface {
	GetGraphOverview(ctx context.Context) (*ExploreGraphResponse, error)
	GetNodesByType(ctx context.Context, request *GetNodesByTypeRequest) (*GetNodesByTypeResponse, error)
	GetNodeRelations(ctx context.Context, request *GetNodeRelationsRequest) (*GetNodeRelationsResponse, error)
	GetConnectedNodes(ctx context.Context, request *GetConnectedNodesRequest) (*GetConnectedNodesResponse, error)
}

type ExploreGraphResponse struct {
	Overview *graph_explorer.GraphOverview `json:"overview"`
}

type GetNodesByTypeRequest struct {
	NodeType string `json:"nodeType"`
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
}

type GetNodesByTypeResponse struct {
	Result *graph.PaginatedNodesResult `json:"result"`
}

type GetNodeRelationsRequest struct {
	NodeID string `json:"nodeId"`
}

type GetNodeRelationsResponse struct {
	NodeID    string            `json:"nodeId"`
	Relations []*graph.Relation `json:"relations"`
}

type GetConnectedNodesRequest struct {
	NodeID string `json:"nodeId"`
}

type GetConnectedNodesResponse struct {
	NodeID      string                  `json:"nodeId"`
	Connections []*graph.NodeConnection `json:"connections"`
}
