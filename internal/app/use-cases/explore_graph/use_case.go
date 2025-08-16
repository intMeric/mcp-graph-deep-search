package explore_graph

import (
	"context"
	"fmt"

	"mgds/internal/app/services/graph_explorer"
)

type exploreGraphUseCase struct {
	graphExplorerService graph_explorer.GraphExplorerService
}

func NewExploreGraphUseCase(
	graphExplorerService graph_explorer.GraphExplorerService,
) ExploreGraphUseCase {
	return &exploreGraphUseCase{
		graphExplorerService: graphExplorerService,
	}
}

func (uc *exploreGraphUseCase) GetGraphOverview(ctx context.Context) (*ExploreGraphResponse, error) {
	overview, err := uc.graphExplorerService.GetGraphOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph overview: %w", err)
	}

	return &ExploreGraphResponse{
		Overview: overview,
	}, nil
}

func (uc *exploreGraphUseCase) GetNodesByType(ctx context.Context, request *GetNodesByTypeRequest) (*GetNodesByTypeResponse, error) {
	if err := uc.validateGetNodesByTypeRequest(request); err != nil {
		return nil, err
	}

	result, err := uc.graphExplorerService.GetNodesByType(ctx, request.NodeType, request.Offset, request.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes by type: %w", err)
	}

	return &GetNodesByTypeResponse{
		Result: result,
	}, nil
}

func (uc *exploreGraphUseCase) GetNodeRelations(ctx context.Context, request *GetNodeRelationsRequest) (*GetNodeRelationsResponse, error) {
	if err := uc.validateGetNodeRelationsRequest(request); err != nil {
		return nil, err
	}

	relations, err := uc.graphExplorerService.GetNodeRelations(ctx, request.NodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node relations: %w", err)
	}

	return &GetNodeRelationsResponse{
		NodeID:    request.NodeID,
		Relations: relations,
	}, nil
}

func (uc *exploreGraphUseCase) GetConnectedNodes(ctx context.Context, request *GetConnectedNodesRequest) (*GetConnectedNodesResponse, error) {
	if err := uc.validateGetConnectedNodesRequest(request); err != nil {
		return nil, err
	}

	connections, err := uc.graphExplorerService.GetConnectedNodes(ctx, request.NodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connected nodes: %w", err)
	}

	return &GetConnectedNodesResponse{
		NodeID:      request.NodeID,
		Connections: connections,
	}, nil
}

func (uc *exploreGraphUseCase) validateGetNodesByTypeRequest(request *GetNodesByTypeRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.NodeType == "" {
		return fmt.Errorf("nodeType is required")
	}

	if request.Offset < 0 {
		return fmt.Errorf("offset must be >= 0")
	}

	if request.Limit <= 0 {
		return fmt.Errorf("limit must be > 0")
	}

	if request.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}

	return nil
}

func (uc *exploreGraphUseCase) validateGetNodeRelationsRequest(request *GetNodeRelationsRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.NodeID == "" {
		return fmt.Errorf("nodeId is required")
	}

	return nil
}

func (uc *exploreGraphUseCase) validateGetConnectedNodesRequest(request *GetConnectedNodesRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.NodeID == "" {
		return fmt.Errorf("nodeId is required")
	}

	return nil
}
