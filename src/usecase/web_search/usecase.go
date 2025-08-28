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

// UseCase interface for the web search use case
type UseCase interface {
	Execute(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
}

// useCase implementation of the use case
type useCase struct {
	webSearchService webseach.Interface
	graphDB          graph.Interface
}

// NewUseCase creates a new instance of the use case
func NewUseCase(webSearchService webseach.Interface, graphDB graph.Interface) UseCase {
	return &useCase{
		webSearchService: webSearchService,
		graphDB:          graphDB,
	}
}

// Execute executes the web search and indexing use case
func (u *useCase) Execute(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
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

	result, err := u.webSearchService.SearchAndIndex(ctx, req.Query, options)
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
		summary := u.convertToNodeSummary(nodeInterface)
		if summary != nil {
			nodes = append(nodes, *summary)
		}
	}

	// Process updated nodes
	for _, nodeInterface := range result.NodesUpdated {
		summary := u.convertToNodeSummary(nodeInterface)
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
func (u *useCase) convertToNodeSummary(nodeInterface node.Interface) *NodeSummary {
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

// Factory to create the use case with dependencies
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

// CreateUseCase creates the use case with all dependencies
func (f *Factory) CreateUseCase() UseCase {
	webSearchService := webseach.NewWebSearchService(f.searchEngine, f.scraper, f.graphDB)
	return NewUseCase(webSearchService, f.graphDB)
}