package scrapper

import (
	"context"
	"time"
)

// ScrapResult contains the result of scraping a URL
type ScrapResult struct {
	URL        string            `json:"url"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Links      []Link            `json:"links"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
	ScrapedAt  time.Time         `json:"scrapedAt"`
	StatusCode int               `json:"statusCode"`
	Error      error             `json:"error,omitempty"`
}

// Link represents a hyperlink found on a webpage
type Link struct {
	URL        string         `json:"url"`
	Text       string         `json:"text"`
	Title      string         `json:"title,omitempty"`
	IsInternal bool           `json:"isInternal"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Config holds configuration for the scrapper
type Config struct {
	UserAgent       string        `json:"userAgent"`
	Timeout         time.Duration `json:"timeout"`
	MaxContentSize  int64         `json:"maxContentSize"`
	FollowRedirects bool          `json:"followRedirects"`
	MaxRedirects    int           `json:"maxRedirects"`
	Parallelism     int           `json:"parallelism"`     // Number of concurrent requests
	Delay           time.Duration `json:"delay"`           // Delay between requests
	RespectRobotsTxt bool         `json:"respectRobotsTxt"` // Respect robots.txt
}

// Interface defines the contract for web scraping operations
type Interface interface {
	// Scrape a single URL and return the result
	Scrape(ctx context.Context, url string) (*ScrapResult, error)
	
	// Scrape multiple URLs in batch for better performance
	ScrapeBatch(ctx context.Context, urls []string) ([]*ScrapResult, error)
	
	// Validate if a URL is valid and scrapable
	IsValidURL(url string) bool
	
	// Normalize URL to avoid duplicates
	NormalizeURL(url string) (string, error)
	
	// Configuration management
	SetConfig(config *Config)
	GetConfig() *Config
	
	// Lifecycle management
	Close() error
}