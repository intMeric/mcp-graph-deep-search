package graph

import (
	"context"
	"mgds/src/pkg/node"
)

// Relation represents a relationship between two nodes
type Relation struct {
	FromNodeId string         `json:"fromNodeId"`
	ToNodeId   string         `json:"toNodeId"`
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Interface defines the contract for graph database operations
type Interface interface {
	// Node operations
	AddNode(ctx context.Context, node node.Interface) error
	GetNode(ctx context.Context, mgdsId string) (node.Interface, error)
	UpdateNode(ctx context.Context, node node.Interface) error
	DeleteNode(ctx context.Context, mgdsId string) error

	// Relation operations
	AddRelation(ctx context.Context, relation *Relation) error
	GetRelations(ctx context.Context, fromNodeId string) ([]*Relation, error)
	GetIncomingRelations(ctx context.Context, toNodeId string) ([]*Relation, error)
	DeleteRelation(ctx context.Context, fromNodeId, toNodeId, label string) error

	// Graph traversal
	GetConnectedNodes(ctx context.Context, mgdsId string) ([]node.Interface, error)
	FindPath(ctx context.Context, fromNodeId, toNodeId string) ([]*Relation, error)

	// Node queries
	FindNodesByType(ctx context.Context, nodeType string) ([]node.Interface, error)
	FindNodesByProperty(ctx context.Context, key string, value any) ([]node.Interface, error)

	// Graph operations
	NodeExists(ctx context.Context, mgdsId string) (bool, error)
	GetAllNodes(ctx context.Context) ([]node.Interface, error)

	// Database lifecycle
	Connect(ctx context.Context) error
	Close() error

	// Health check
	Ping(ctx context.Context) error
}

// Config holds configuration for graph database connections
type Config struct {
	DatabasePath string         `json:"databasePath"`
	DatabaseType string         `json:"databaseType"` // "cayley", "neo4j", etc.
	Options      map[string]any `json:"options"`
}
