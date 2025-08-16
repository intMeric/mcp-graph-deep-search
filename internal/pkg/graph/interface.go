package graph

import (
	"context"
	"mgds/internal/pkg/node"
)

type Graph interface {
	CreateNode(ctx context.Context, node *node.Node) error
	CreateRelation(ctx context.Context, relation *Relation) error
	CreateLink(ctx context.Context, sourceNode *node.Node, targetNode *node.Node, relationType string) (*Relation, error)
	GetNode(ctx context.Context, id string) (*node.Node, error)
	NodeExists(ctx context.Context, id string) (bool, error)
	Close(ctx context.Context) error

	// Graph exploration methods
	GetGraphStats(ctx context.Context) (*GraphStats, error)
	GetNodesByType(ctx context.Context, nodeType string, offset, limit int) (*PaginatedNodesResult, error)
	GetNodeRelations(ctx context.Context, nodeID string) ([]*Relation, error)
	GetConnectedNodes(ctx context.Context, nodeID string) ([]*NodeConnection, error)
}
