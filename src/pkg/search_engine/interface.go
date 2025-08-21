package search_engine

import (
	"context"
)

// Interface defines the contract for search engine operations
type Interface interface {
	// Search performs a search with the given request
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
	
	// SearchSimple performs a simple search with just a query string
	// Uses default categories (general, news) and time range (all)
	SearchSimple(ctx context.Context, query string) (*SearchResponse, error)
	
	// IsHealthy checks if the search engine is accessible
	IsHealthy(ctx context.Context) error
	
	// Configuration management
	SetConfig(config *Config)
	GetConfig() *Config
	
	// Lifecycle management
	Close() error
}