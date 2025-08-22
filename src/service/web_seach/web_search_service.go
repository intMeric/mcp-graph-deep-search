package webseach

import (
	"context"
	"fmt"
	"mgds/src/pkg/graph"
	"mgds/src/pkg/node"
	"mgds/src/pkg/object"
	"mgds/src/pkg/scrapper"
	"mgds/src/pkg/search_engine"
	"sync"
	"time"
)

// WebSearchService implements the Interface for web search and indexing operations
type WebSearchService struct {
	searchEngine search_engine.Interface
	scrapper     scrapper.Interface
	graph        graph.Interface

	// Default options
	defaultOptions *SearchOptions
}

// NewWebSearchService creates a new web search service
func NewWebSearchService(
	searchEngine search_engine.Interface,
	scrapper scrapper.Interface,
	graph graph.Interface,
) Interface {
	return &WebSearchService{
		searchEngine: searchEngine,
		scrapper:     scrapper,
		graph:        graph,
		defaultOptions: &SearchOptions{
			MaxResults:   10,
			MaxLinks:     50,
			SkipLinking:  false,
			ScrapTimeout: 30 * time.Second,
			Parallelism:  3,
		},
	}
}

// SearchAndIndex performs a search, scrapes the results, and indexes everything in the graph
func (s *WebSearchService) SearchAndIndex(ctx context.Context, query string, options *SearchOptions) (*SearchIndexResult, error) {
	startTime := time.Now()

	// Merge options with defaults
	opts := s.mergeOptions(options)

	result := &SearchIndexResult{
		Query:       query,
		Stats:       &IndexStats{},
		TimeElapsed: 0,
	}

	// Step 1: Perform search
	searchResp, err := s.searchEngine.SearchSimple(ctx, query)
	if err != nil {
		return result, fmt.Errorf("search failed: %w", err)
	}

	result.Stats.SearchResults = len(searchResp.Results)

	// Limit results if specified
	results := searchResp.Results
	if opts.MaxResults > 0 && len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}

	// Step 2: Scrape all URLs concurrently
	urls := make([]string, len(results))
	for i, result := range results {
		urls[i] = result.URL
	}

	scrapResults, err := s.scrapeConcurrently(ctx, urls, opts)
	if err != nil {
		return result, fmt.Errorf("scraping failed: %w", err)
	}

	// Step 3: Index scraped content
	for _, scrapResult := range scrapResults {
		if scrapResult.Error != nil {
			result.Errors = append(result.Errors, IndexError{
				URL:     scrapResult.URL,
				Type:    "scraping",
				Message: scrapResult.Error.Error(),
			})
			result.Stats.SkippedPages++
			continue
		}

		result.Stats.ScrapedPages++
		result.Stats.LinksFound += len(scrapResult.Links)

		// Create main page node
		nodeId := s.generateNodeId(scrapResult.URL)
		urlNode := object.NewURLNodeWithDetails(
			nodeId,
			scrapResult.Title,
			fmt.Sprintf("Search result for: %s", query),
			scrapResult.URL,
			scrapResult.Title,
			scrapResult.Content,
		)

		if err := s.addOrUpdateNode(ctx, urlNode, result); err != nil {
			result.Errors = append(result.Errors, IndexError{
				URL:     scrapResult.URL,
				Type:    "indexing",
				Message: err.Error(),
			})
			continue
		}

		// Create nodes for links and relationships if not skipping
		if !opts.SkipLinking {
			s.indexLinks(ctx, nodeId, scrapResult.Links, opts, result)
		}
	}

	result.TimeElapsed = time.Since(startTime)
	return result, nil
}

// SearchAndIndexURL directly scrapes and indexes a single URL and its links
func (s *WebSearchService) SearchAndIndexURL(ctx context.Context, url string, options *SearchOptions) (*SearchIndexResult, error) {
	startTime := time.Now()

	opts := s.mergeOptions(options)

	result := &SearchIndexResult{
		Query: fmt.Sprintf("Direct URL: %s", url),
		Stats: &IndexStats{
			SearchResults: 1,
		},
		TimeElapsed: 0,
	}

	// Scrape the URL
	scrapResult, err := s.scrapper.Scrape(ctx, url)
	if err != nil {
		result.Errors = append(result.Errors, IndexError{
			URL:     url,
			Type:    "scraping",
			Message: err.Error(),
		})
		result.Stats.SkippedPages++
		result.TimeElapsed = time.Since(startTime)
		return result, nil
	}

	result.Stats.ScrapedPages++
	result.Stats.LinksFound = len(scrapResult.Links)

	// Create main page node
	nodeId := s.generateNodeId(scrapResult.URL)
	urlNode := object.NewURLNodeWithDetails(
		nodeId,
		scrapResult.Title,
		"Direct URL indexing",
		scrapResult.URL,
		scrapResult.Title,
		scrapResult.Content,
	)

	if err := s.addOrUpdateNode(ctx, urlNode, result); err != nil {
		result.Errors = append(result.Errors, IndexError{
			URL:     scrapResult.URL,
			Type:    "indexing",
			Message: err.Error(),
		})
	} else {
		// Create nodes for links and relationships if not skipping
		if !opts.SkipLinking {
			s.indexLinks(ctx, nodeId, scrapResult.Links, opts, result)
		}
	}

	result.TimeElapsed = time.Since(startTime)
	return result, nil
}

// Close releases resources used by the service
func (s *WebSearchService) Close() error {
	var errs []error

	if err := s.searchEngine.Close(); err != nil {
		errs = append(errs, fmt.Errorf("search engine close: %w", err))
	}

	if err := s.scrapper.Close(); err != nil {
		errs = append(errs, fmt.Errorf("scrapper close: %w", err))
	}

	if err := s.graph.Close(); err != nil {
		errs = append(errs, fmt.Errorf("graph close: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// Helper methods

func (s *WebSearchService) mergeOptions(options *SearchOptions) *SearchOptions {
	if options == nil {
		return s.defaultOptions
	}

	opts := *s.defaultOptions
	if options.MaxResults > 0 {
		opts.MaxResults = options.MaxResults
	}
	if options.MaxLinks > 0 {
		opts.MaxLinks = options.MaxLinks
	}
	if options.SkipLinking {
		opts.SkipLinking = options.SkipLinking
	}
	if options.ScrapTimeout > 0 {
		opts.ScrapTimeout = options.ScrapTimeout
	}
	if options.Parallelism > 0 {
		opts.Parallelism = options.Parallelism
	}

	return &opts
}

func (s *WebSearchService) scrapeConcurrently(ctx context.Context, urls []string, opts *SearchOptions) ([]*scrapper.ScrapResult, error) {
	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, opts.Parallelism)
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]*scrapper.ScrapResult, 0, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create context with timeout
			scrapCtx, cancel := context.WithTimeout(ctx, opts.ScrapTimeout)
			defer cancel()

			scrapResult, err := s.scrapper.Scrape(scrapCtx, url)
			if err != nil {
				scrapResult = &scrapper.ScrapResult{
					URL:   url,
					Error: err,
				}
			}

			mu.Lock()
			results = append(results, scrapResult)
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	return results, nil
}

func (s *WebSearchService) generateNodeId(url string) string {
	return fmt.Sprintf("url_%s", s.hashURL(url))
}

func (s *WebSearchService) hashURL(url string) string {
	// Simple hash for now - could use crypto/sha256 for better distribution
	hash := 0
	for _, char := range url {
		hash = hash*31 + int(char)
		if hash < 0 {
			hash = -hash
		}
	}
	return fmt.Sprintf("%d", hash)
}

func (s *WebSearchService) addOrUpdateNode(ctx context.Context, node node.Interface, result *SearchIndexResult) error {
	exists, err := s.graph.NodeExists(ctx, node.GetMgdsId())
	if err != nil {
		return fmt.Errorf("failed to check node existence: %w", err)
	}

	if exists {
		if err := s.graph.UpdateNode(ctx, node); err != nil {
			return fmt.Errorf("failed to update node: %w", err)
		}
		result.NodesUpdated = append(result.NodesUpdated, node)
		result.Stats.NodesUpdated++
	} else {
		if err := s.graph.AddNode(ctx, node); err != nil {
			return fmt.Errorf("failed to add node: %w", err)
		}
		result.NodesCreated = append(result.NodesCreated, node)
		result.Stats.NodesCreated++
	}

	return nil
}

func (s *WebSearchService) indexLinks(ctx context.Context, fromNodeId string, links []scrapper.Link, opts *SearchOptions, result *SearchIndexResult) {
	linksToProcess := links
	if opts.MaxLinks > 0 && len(links) > opts.MaxLinks {
		linksToProcess = links[:opts.MaxLinks]
	}

	for _, link := range linksToProcess {
		// Create node for the linked page
		linkNodeId := s.generateNodeId(link.URL)
		linkNode := object.NewURLNode(
			linkNodeId,
			link.Text,
			fmt.Sprintf("Linked from: %s", fromNodeId),
			link.URL,
		)

		// Add the link node (ignore errors for now to keep going)
		_ = s.addOrUpdateNode(ctx, linkNode, result)

		// Create relationship
		relation := &graph.Relation{
			FromNodeId: fromNodeId,
			ToNodeId:   linkNodeId,
			Label:      "links_to",
			Properties: map[string]any{
				"link_text":   link.Text,
				"is_internal": link.IsInternal,
			},
		}

		if err := s.graph.AddRelation(ctx, relation); err == nil {
			result.RelationsCreated++
			result.Stats.RelationsCreated++
		} else {
			result.Errors = append(result.Errors, IndexError{
				URL:     link.URL,
				Type:    "relation",
				Message: err.Error(),
			})
		}
	}
}
