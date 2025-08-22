package websearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/node"
	"mgds/src/pkg/object"
	"mgds/src/pkg/scrapper"
	"mgds/src/pkg/search_engine"
	webseach "mgds/src/service/web_seach"
)

// Service interface for the web search use case
type Service interface {
	Execute(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
}

// service implementation of the use case
type service struct {
	webSearchService webseach.Interface
	graphDB          graph.Interface
}

// NewService creates a new instance of the use case service
func NewService(webSearchService webseach.Interface, graphDB graph.Interface) Service {
	return &service{
		webSearchService: webSearchService,
		graphDB:          graphDB,
	}
}

// Execute executes the web search and indexing use case
func (s *service) Execute(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()

	// Simple validation
	if strings.TrimSpace(req.Query) == "" {
		return &SearchResponse{
			Query:         req.Query,
			Status:        StatusFailed,
			Error:         "Query cannot be empty",
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	if req.MaxResults <= 0 {
		req.MaxResults = 10
	}

	options := &webseach.SearchOptions{
		MaxResults:  req.MaxResults,
		SkipLinking: false,
		MaxLinks:    300,
	}

	result, err := s.webSearchService.SearchAndIndex(ctx, req.Query, options)
	if err != nil {
		return &SearchResponse{
			Query:         req.Query,
			Status:        StatusFailed,
			Error:         fmt.Sprintf("Search and index failed: %v", err),
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	// Convert nodes to NodeSummary
	var nodes []NodeSummary

	// Process created nodes
	for _, nodeInterface := range result.NodesCreated {
		summary := s.convertToNodeSummary(nodeInterface)
		if summary != nil {
			nodes = append(nodes, *summary)
		}
	}

	// Process updated nodes
	for _, nodeInterface := range result.NodesUpdated {
		summary := s.convertToNodeSummary(nodeInterface)
		if summary != nil {
			nodes = append(nodes, *summary)
		}
	}

	return &SearchResponse{
		Query:         req.Query,
		Status:        StatusSuccess,
		Nodes:         nodes,
		ExecutionTime: time.Since(startTime),
	}, nil
}

// convertToNodeSummary converts a node.Interface to NodeSummary
func (s *service) convertToNodeSummary(nodeInterface node.Interface) *NodeSummary {
	if nodeInterface == nil {
		return nil
	}

	summary := &NodeSummary{
		ID: nodeInterface.GetMgdsId(),
	}

	// If it's a URLNode, use specific methods
	if urlNode, ok := nodeInterface.(*object.URLNode); ok {
		summary.URL = urlNode.GetURL()
		summary.Title = urlNode.GetTitle()
		summary.Content = urlNode.GetContent()
	} else {
		// Fallback for other node types
		summary.Title = nodeInterface.GetDisplayName()
		summary.Content = nodeInterface.GetDescription()

		// Try to get URL via properties
		if url, exists := nodeInterface.GetProperty("url"); exists {
			if urlStr, ok := url.(string); ok {
				summary.URL = urlStr
			}
		}
	}

	return summary
}

// Factory to create the service with dependencies
type Factory struct {
	searchEngine search_engine.Interface
	scraper      scrapper.Interface
	graphDB      graph.Interface
}

// NewFactory creates a new factory
func NewFactory(searchEngine search_engine.Interface, scraper scrapper.Interface, graphDB graph.Interface) *Factory {
	return &Factory{
		searchEngine: searchEngine,
		scraper:      scraper,
		graphDB:      graphDB,
	}
}

// CreateService creates the use case service with all dependencies
func (f *Factory) CreateService() Service {
	webSearchService := webseach.NewWebSearchService(f.searchEngine, f.scraper, f.graphDB)
	return NewService(webSearchService, f.graphDB)
}
