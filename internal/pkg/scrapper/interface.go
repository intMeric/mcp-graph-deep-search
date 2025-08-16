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
	URL         string            `json:"url"`
	Title       string            `json:"title"`
	Text        string            `json:"text"`
	HTMLBody    string            `json:"html_body"`
	Links       []Link            `json:"links"`
	Images      []Image           `json:"images"`
	Forms       []Form            `json:"forms"`
	Scripts     []string          `json:"scripts"`
	Stylesheets []string          `json:"stylesheets"`
	MetaTags    map[string]string `json:"meta_tags"`
	Headers     map[string]string `json:"headers"`
	StatusCode  int               `json:"status_code"`
	ScrapedAt   time.Time         `json:"scraped_at"`
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

type Image struct {
	URL    string `json:"url"`
	Alt    string `json:"alt"`
	Title  string `json:"title,omitempty"`
	Width  string `json:"width,omitempty"`
	Height string `json:"height,omitempty"`
}

type Form struct {
	Action string      `json:"action"`
	Method string      `json:"method"`
	Name   string      `json:"name,omitempty"`
	ID     string      `json:"id,omitempty"`
	Fields []FormField `json:"fields"`
}

type FormField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       string `json:"value,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Required    bool   `json:"required"`
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
	ExtractImages     bool          `json:"extract_images"`
	ExtractForms      bool          `json:"extract_forms"`
	ExtractScripts    bool          `json:"extract_scripts"`
	ExtractMeta       bool          `json:"extract_meta"`
}

func DefaultScrapingOptions() *ScrapingOptions {
	return &ScrapingOptions{
		Timeout:         30 * time.Second,
		UserAgent:       "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0; +https://openai.com/bot",
		FollowRedirects: true,
		MaxDepth:        1,
		RateLimitDelay:  1 * time.Second,
		ExtractText:     true,
		ExtractHTML:     true,
		ExtractLinks:    true,
		ExtractImages:   true,
		ExtractForms:    true,
		ExtractScripts:  true,
		ExtractMeta:     true,
	}
}

