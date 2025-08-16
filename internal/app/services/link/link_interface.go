package link

import (
	"context"

	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

type LinkService interface {
	CreateLink(ctx context.Context, sourceNode *node.Node, targetNode *node.Node, linkType string) (*LinkResponse, error)
	Close() error
}

type LinkResponse struct {
	Success       bool            `json:"success"`
	Message       string          `json:"message,omitempty"`
	Relation      *graph.Relation `json:"relation,omitempty"`
	SourceCreated bool            `json:"source_created"`
	TargetCreated bool            `json:"target_created"`
}
