package prune_graph

import (
	"context"
	"fmt"
	"strings"

	"mgds/internal/app/services/graph_pruner"
)

type pruneGraphUseCase struct {
	prunerService graph_pruner.GraphPrunerService
}

func NewPruneGraphUseCase(prunerService graph_pruner.GraphPrunerService) PruneGraphUseCase {
	return &pruneGraphUseCase{
		prunerService: prunerService,
	}
}

func (uc *pruneGraphUseCase) DeleteNode(ctx context.Context, request *DeleteNodeRequest) (*DeleteNodeResponse, error) {
	if err := uc.validateDeleteNodeRequest(request); err != nil {
		return nil, err
	}

	result, err := uc.prunerService.DeleteNode(ctx, request.NodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete node: %w", err)
	}

	return &DeleteNodeResponse{
		NodeID: request.NodeID,
		Result: result,
	}, nil
}

func (uc *pruneGraphUseCase) DeleteNodeCascade(ctx context.Context, request *DeleteNodeCascadeRequest) (*DeleteNodeCascadeResponse, error) {
	if err := uc.validateDeleteNodeCascadeRequest(request); err != nil {
		return nil, err
	}

	maxDepth := request.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 5 // Safe default for cascade operations
	}

	result, err := uc.prunerService.DeleteNodeCascade(ctx, request.NodeID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to delete node cascade: %w", err)
	}

	return &DeleteNodeCascadeResponse{
		NodeID:   request.NodeID,
		MaxDepth: maxDepth,
		Result:   result,
	}, nil
}

func (uc *pruneGraphUseCase) DeleteRelation(ctx context.Context, request *DeleteRelationRequest) (*DeleteRelationResponse, error) {
	if err := uc.validateDeleteRelationRequest(request); err != nil {
		return nil, err
	}

	err := uc.prunerService.DeleteRelation(ctx, request.SourceID, request.TargetID, request.RelationType)
	success := err == nil

	response := &DeleteRelationResponse{
		SourceID:     request.SourceID,
		TargetID:     request.TargetID,
		RelationType: request.RelationType,
		Success:      success,
	}

	if err != nil {
		return response, fmt.Errorf("failed to delete relation: %w", err)
	}

	return response, nil
}

func (uc *pruneGraphUseCase) PreviewDeletion(ctx context.Context, request *PreviewDeletionRequest) (*PreviewDeletionResponse, error) {
	if err := uc.validatePreviewDeletionRequest(request); err != nil {
		return nil, err
	}

	maxDepth := request.MaxDepth
	if request.Cascade && maxDepth <= 0 {
		maxDepth = 5 // Safe default for cascade previews
	}

	preview, err := uc.prunerService.PreviewDeletion(ctx, request.NodeID, request.Cascade, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to preview deletion: %w", err)
	}

	return &PreviewDeletionResponse{
		NodeID:  request.NodeID,
		Cascade: request.Cascade,
		Preview: preview,
	}, nil
}

func (uc *pruneGraphUseCase) validateDeleteNodeRequest(request *DeleteNodeRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(request.NodeID) == "" {
		return fmt.Errorf("nodeID cannot be empty")
	}
	return nil
}

func (uc *pruneGraphUseCase) validateDeleteNodeCascadeRequest(request *DeleteNodeCascadeRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(request.NodeID) == "" {
		return fmt.Errorf("nodeID cannot be empty")
	}
	if request.MaxDepth < 0 {
		return fmt.Errorf("maxDepth cannot be negative")
	}
	if request.MaxDepth > 20 {
		return fmt.Errorf("maxDepth cannot exceed 20 for safety reasons")
	}
	return nil
}

func (uc *pruneGraphUseCase) validateDeleteRelationRequest(request *DeleteRelationRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(request.SourceID) == "" {
		return fmt.Errorf("sourceID cannot be empty")
	}
	if strings.TrimSpace(request.TargetID) == "" {
		return fmt.Errorf("targetID cannot be empty")
	}
	if strings.TrimSpace(request.RelationType) == "" {
		return fmt.Errorf("relationType cannot be empty")
	}
	if request.SourceID == request.TargetID {
		return fmt.Errorf("sourceID and targetID cannot be the same")
	}
	return nil
}

func (uc *pruneGraphUseCase) validatePreviewDeletionRequest(request *PreviewDeletionRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(request.NodeID) == "" {
		return fmt.Errorf("nodeID cannot be empty")
	}
	if request.MaxDepth < 0 {
		return fmt.Errorf("maxDepth cannot be negative")
	}
	if request.MaxDepth > 20 {
		return fmt.Errorf("maxDepth cannot exceed 20 for safety reasons")
	}
	return nil
}