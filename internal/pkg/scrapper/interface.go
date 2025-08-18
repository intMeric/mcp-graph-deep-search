package scrapper

import (
	"context"
	"time"
)

type WebScraper interface {
	Scrape(ctx context.Context, url string, options *ScrapingOptions) (*ScrapedData, error)
	ScrapeMultiple(ctx context.Context, urls []string, options *ScrapingOptions) ([]*ScrapedData, error)
	SetUserAgent(userAgent string)
	SetTimeout(timeout time.Duration)
	Close() error
}

type ScrapedData struct {
	URL        string            `json:"url"`
	Title      string            `json:"title"`
	Text       string            `json:"text"`
	HTMLBody   string            `json:"html_body"`
	Links      []Link            `json:"links"`
	MetaTags   map[string]string `json:"meta_tags"`
	Headers    map[string]string `json:"headers"`
	StatusCode int               `json:"status_code"`
	ScrapedAt  time.Time         `json:"scraped_at"`
}

type Link struct {
	URL          string `json:"url"`
	Text         string `json:"text"`
	Rel          string `json:"rel,omitempty"`
	Target       string `json:"target,omitempty"`
	Download     string `json:"download,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	IsExternal   bool   `json:"is_external"`
}

type ScrapingOptions struct {
	Timeout           time.Duration `json:"timeout"`
	UserAgent         string        `json:"user_agent"`
	FollowRedirects   bool          `json:"follow_redirects"`
	MaxDepth          int           `json:"max_depth"`
	AllowedDomains    []string      `json:"allowed_domains"`
	DisallowedDomains []string      `json:"disallowed_domains"`
	RateLimitDelay    time.Duration `json:"rate_limit_delay"`
	ExtractText       bool          `json:"extract_text"`
	ExtractHTML       bool          `json:"extract_html"`
	ExtractLinks      bool          `json:"extract_links"`
	ExtractMeta       bool          `json:"extract_meta"`
}

func DefaultScrapingOptions() *ScrapingOptions {
	return &ScrapingOptions{
		Timeout:         10 * time.Second,
		UserAgent:       "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; MGDS",
		FollowRedirects: true,
		MaxDepth:        1,
		RateLimitDelay:  200 * time.Millisecond,
		ExtractText:     true,
		ExtractHTML:     false,
		ExtractLinks:    true,
		ExtractMeta:     true,
	}
}
