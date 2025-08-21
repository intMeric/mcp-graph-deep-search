package search_engine

import (
	"time"
)

// TimeRange represents the time range for search results
type TimeRange string

const (
	TimeRangeShort TimeRange = "short" // Less than 1 month
	TimeRangeAll   TimeRange = "all"   // All time
)

// Category represents search categories
type Category string

const (
	CategoryGeneral Category = "general"
	CategoryNews    Category = "news"
)

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string      `json:"query"`
	Categories []Category  `json:"categories"`
	TimeRange  TimeRange   `json:"timeRange"`
	Language   string      `json:"language,omitempty"`
	SafeSearch int         `json:"safeSearch,omitempty"` // 0=none, 1=moderate, 2=strict
	PageNo     int         `json:"pageNo,omitempty"`
	MaxResults int         `json:"maxResults,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	URL           string            `json:"url"`
	Title         string            `json:"title"`
	Content       string            `json:"content"`
	Snippet       string            `json:"snippet,omitempty"`
	Score         float64           `json:"score,omitempty"`
	Engine        string            `json:"engine,omitempty"`
	Category      Category          `json:"category"`
	PublishedDate *time.Time        `json:"publishedDate,omitempty"`
	Metadata      map[string]any    `json:"metadata,omitempty"`
}

// SearchResponse represents the response from a search request
type SearchResponse struct {
	Results      []SearchResult `json:"results"`
	TotalResults int            `json:"totalResults,omitempty"`
	Query        string         `json:"query"`
	TimeElapsed  time.Duration  `json:"timeElapsed"`
	NextPage     int            `json:"nextPage,omitempty"`
	PrevPage     int            `json:"prevPage,omitempty"`
	CurrentPage  int            `json:"currentPage"`
	HasNextPage  bool           `json:"hasNextPage"`
	HasPrevPage  bool           `json:"hasPrevPage"`
}

// Config holds configuration for search engine
type Config struct {
	BaseURL          string        `json:"baseUrl"`
	Timeout          time.Duration `json:"timeout"`
	UserAgent        string        `json:"userAgent"`
	DefaultLanguage  string        `json:"defaultLanguage"`
	DefaultSafeSearch int          `json:"defaultSafeSearch"`
	MaxResults       int           `json:"maxResults"`
	DefaultTimeRange TimeRange     `json:"defaultTimeRange"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:           "https://searx.example.com",
		Timeout:           30 * time.Second,
		UserAgent:         "mgds/1.0",
		DefaultLanguage:   "en",
		DefaultSafeSearch: 1,
		MaxResults:        50,
		DefaultTimeRange:  TimeRangeAll,
	}
}