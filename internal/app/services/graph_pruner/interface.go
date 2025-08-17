package graph_pruner

import (
	"context"

	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

type GraphPrunerService interface {
	DeleteNode(ctx context.Context, nodeID string) (*graph.DeletionResult, error)
	DeleteNodeCascade(ctx context.Context, nodeID string, maxDepth int) (*graph.DeletionResult, error)
	DeleteRelation(ctx context.Context, sourceID, targetID, relationType string) error
	PreviewDeletion(ctx context.Context, nodeID string, cascade bool, maxDepth int) (*DeletionPreview, error)
	Close() error
}

type DeletionPreview struct {
	TargetNode       *node.Node    `json:"targetNode"`
	AffectedNodes    []*node.Node  `json:"affectedNodes"`
	TotalNodes       int           `json:"totalNodes"`
	TotalRelations   int           `json:"totalRelations"`
	HasDocuments     bool          `json:"hasDocuments"`
	Warning          string        `json:"warning,omitempty"`
}