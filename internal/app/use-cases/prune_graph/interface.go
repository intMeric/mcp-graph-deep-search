package prune_graph

import (
	"context"

	"mgds/internal/app/services/graph_pruner"
	"mgds/internal/pkg/graph"
)

type PruneGraphUseCase interface {
	DeleteNode(ctx context.Context, request *DeleteNodeRequest) (*DeleteNodeResponse, error)
	DeleteNodeCascade(ctx context.Context, request *DeleteNodeCascadeRequest) (*DeleteNodeCascadeResponse, error)
	DeleteRelation(ctx context.Context, request *DeleteRelationRequest) (*DeleteRelationResponse, error)
	PreviewDeletion(ctx context.Context, request *PreviewDeletionRequest) (*PreviewDeletionResponse, error)
}

type DeleteNodeRequest struct {
	NodeID string `json:"nodeId"`
}

type DeleteNodeResponse struct {
	NodeID string                 `json:"nodeId"`
	Result *graph.DeletionResult  `json:"result"`
}

type DeleteNodeCascadeRequest struct {
	NodeID   string `json:"nodeId"`
	MaxDepth int    `json:"maxDepth,omitempty"`
}

type DeleteNodeCascadeResponse struct {
	NodeID   string                 `json:"nodeId"`
	MaxDepth int                    `json:"maxDepth"`
	Result   *graph.DeletionResult  `json:"result"`
}

type DeleteRelationRequest struct {
	SourceID     string `json:"sourceId"`
	TargetID     string `json:"targetId"`
	RelationType string `json:"relationType"`
}

type DeleteRelationResponse struct {
	SourceID     string `json:"sourceId"`
	TargetID     string `json:"targetId"`
	RelationType string `json:"relationType"`
	Success      bool   `json:"success"`
}

type PreviewDeletionRequest struct {
	NodeID   string `json:"nodeId"`
	Cascade  bool   `json:"cascade,omitempty"`
	MaxDepth int    `json:"maxDepth,omitempty"`
}

type PreviewDeletionResponse struct {
	NodeID  string                              `json:"nodeId"`
	Cascade bool                                `json:"cascade"`
	Preview *graph_pruner.DeletionPreview       `json:"preview"`
}