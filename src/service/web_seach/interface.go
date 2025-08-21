package webseach

import (
	"context"
	"mgds/src/pkg/node"
	"time"
)

// SearchOptions configures the web search and indexing operation
type SearchOptions struct {
	MaxResults   int           `json:"maxResults,omitempty"`
	MaxLinks     int           `json:"maxLinks,omitempty"`     // Max links per page to process
	SkipScraping bool          `json:"skipScraping,omitempty"` // Only search, don't scrape
	SkipLinking  bool          `json:"skipLinking,omitempty"`  // Don't create link relations
	ScrapTimeout time.Duration `json:"scrapTimeout,omitempty"` // Timeout per scraping operation
	Parallelism  int           `json:"parallelism,omitempty"`  // Number of concurrent scraping operations
}

// SearchIndexResult contains the results of a search and indexing operation
type SearchIndexResult struct {
	Query            string           `json:"query"`
	NodesCreated     []node.Interface `json:"nodesCreated"`
	NodesUpdated     []node.Interface `json:"nodesUpdated"`
	RelationsCreated int              `json:"relationsCreated"`
	Errors           []IndexError     `json:"errors,omitempty"`
	TimeElapsed      time.Duration    `json:"timeElapsed"`
	Stats            *IndexStats      `json:"stats"`
}

// IndexError represents an error that occurred during indexing
type IndexError struct {
	URL     string `json:"url"`
	Type    string `json:"type"` // "scraping", "indexing", "relation"
	Message string `json:"message"`
}

// IndexStats provides statistics about the indexing operation
type IndexStats struct {
	SearchResults    int `json:"searchResults"`
	ScrapedPages     int `json:"scrapedPages"`
	SkippedPages     int `json:"skippedPages"`
	LinksFound       int `json:"linksFound"`
	NodesCreated     int `json:"nodesCreated"`
	NodesUpdated     int `json:"nodesUpdated"`
	RelationsCreated int `json:"relationsCreated"`
}

// Interface defines the contract for web search and indexing service
type Interface interface {
	// SearchAndIndex performs a search, scrapes the results, and indexes everything in the graph
	SearchAndIndex(ctx context.Context, query string, options *SearchOptions) (*SearchIndexResult, error)

	// SearchAndIndexURL directly scrapes and indexes a single URL and its links
	SearchAndIndexURL(ctx context.Context, url string, options *SearchOptions) (*SearchIndexResult, error)

	// Close releases resources used by the service
	Close() error
}
