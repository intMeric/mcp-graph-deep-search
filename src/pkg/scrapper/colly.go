package scrapper

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// CollyScrapper implements the Interface using Colly framework
type CollyScrapper struct {
	collector *colly.Collector
	config    *Config
}

// NewCollyScrapper creates a new Colly-based scrapper
func NewCollyScrapper(config *Config) Interface {
	if config == nil {
		config = &Config{
			UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Timeout:         30 * time.Second,
			MaxContentSize:  10 * 1024 * 1024, // 10MB
			FollowRedirects: true,
			MaxRedirects:    5,
		}
	}

	c := colly.NewCollector(
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
	)

	// Set basic configuration
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	c.SetRequestTimeout(config.Timeout)

	return &CollyScrapper{
		collector: c,
		config:    config,
	}
}

// NewCollyScrapperWithDebug creates a scrapper with debug logging
func NewCollyScrapperWithDebug(config *Config) Interface {
	if config == nil {
		config = &Config{
			UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Timeout:         30 * time.Second,
			MaxContentSize:  10 * 1024 * 1024,
			FollowRedirects: true,
			MaxRedirects:    5,
		}
	}

	c := colly.NewCollector(
		colly.Async(true),
		colly.Debugger(&debug.LogDebugger{}),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	c.SetRequestTimeout(config.Timeout)

	return &CollyScrapper{
		collector: c,
		config:    config,
	}
}

// Scrape implements the Interface.Scrape method
func (cs *CollyScrapper) Scrape(ctx context.Context, targetURL string) (*ScrapResult, error) {
	if !cs.IsValidURL(targetURL) {
		return nil, fmt.Errorf("invalid URL: %s", targetURL)
	}

	normalizedURL, err := cs.NormalizeURL(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize URL: %w", err)
	}

	result := &ScrapResult{
		URL:       normalizedURL,
		Links:     make([]Link, 0),
		Metadata:  make(map[string]any),
		ScrapedAt: time.Now(),
	}

	// Set user agent
	cs.collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", cs.config.UserAgent)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})

	// Handle HTML responses
	cs.collector.OnHTML("html", func(e *colly.HTMLElement) {
		// Extract title
		title := e.ChildText("title")
		result.Title = strings.TrimSpace(title)

		// Extract meta description
		description := e.ChildAttr("meta[name=description]", "content")
		if description != "" {
			result.Metadata["description"] = description
		}

		// Extract meta keywords
		keywords := e.ChildAttr("meta[name=keywords]", "content")
		if keywords != "" {
			result.Metadata["keywords"] = keywords
		}

		// Extract meta author
		author := e.ChildAttr("meta[name=author]", "content")
		if author != "" {
			result.Metadata["author"] = author
		}

		// Extract Open Graph data
		ogTitle := e.ChildAttr("meta[property='og:title']", "content")
		if ogTitle != "" {
			result.Metadata["og:title"] = ogTitle
		}

		ogDescription := e.ChildAttr("meta[property='og:description']", "content")
		if ogDescription != "" {
			result.Metadata["og:description"] = ogDescription
		}

		ogImage := e.ChildAttr("meta[property='og:image']", "content")
		if ogImage != "" {
			result.Metadata["og:image"] = ogImage
		}

		// Extract language
		lang := e.Attr("lang")
		if lang != "" {
			result.Metadata["language"] = lang
		}

		// Extract main content (prioritize main, article, then body)
		content := ""
		if mainContent := e.ChildText("main"); mainContent != "" {
			content = mainContent
		} else if articleContent := e.ChildText("article"); articleContent != "" {
			content = articleContent
		} else {
			content = e.ChildText("body")
		}

		// Clean and limit content
		content = strings.TrimSpace(content)
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.ReplaceAll(content, "\t", " ")
		for strings.Contains(content, "  ") {
			content = strings.ReplaceAll(content, "  ", " ")
		}

		if int64(len(content)) > cs.config.MaxContentSize {
			content = content[:cs.config.MaxContentSize]
		}

		result.Content = content
	})

	// Extract all links
	cs.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		text := strings.TrimSpace(e.Text)
		title := e.Attr("title")

		if href == "" {
			return
		}

		// Resolve relative URLs
		absoluteURL := e.Request.AbsoluteURL(href)
		if absoluteURL == "" {
			return
		}

		// Parse URL to check if it's internal
		baseURL, err := url.Parse(normalizedURL)
		if err != nil {
			return
		}

		linkURL, err := url.Parse(absoluteURL)
		if err != nil {
			return
		}

		isInternal := baseURL.Host == linkURL.Host

		// Extract additional properties
		properties := make(map[string]any)
		if rel := e.Attr("rel"); rel != "" {
			properties["rel"] = rel
		}
		if class := e.Attr("class"); class != "" {
			properties["class"] = class
		}
		if target := e.Attr("target"); target != "" {
			properties["target"] = target
		}

		link := Link{
			URL:        absoluteURL,
			Text:       text,
			Title:      title,
			IsInternal: isInternal,
			Properties: properties,
		}

		result.Links = append(result.Links, link)
	})

	// Handle response
	cs.collector.OnResponse(func(r *colly.Response) {
		result.StatusCode = r.StatusCode
	})

	// Create a done channel to wait for completion (moved up before callbacks)
	done := make(chan bool, 1) // Buffered to prevent blocking

	// Handle errors
	cs.collector.OnError(func(r *colly.Response, err error) {
		result.Error = err
		if r != nil {
			result.StatusCode = r.StatusCode
		} else {
			result.StatusCode = 0
		}
		select {
		case done <- true: // Signal completion even on error
		default:
		}
	})

	cs.collector.OnScraped(func(r *colly.Response) {
		select {
		case done <- true:
		default:
		}
	})

	// Visit the URL
	err = cs.collector.Visit(normalizedURL)
	if err != nil {
		result.Error = err
		result.StatusCode = 0
		return result, nil // Return result with error, not nil
	}

	// Wait for scraping to complete or context to be cancelled
	select {
	case <-done:
		// Scraping completed successfully
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(cs.config.Timeout):
		return nil, fmt.Errorf("scraping timeout after %v", cs.config.Timeout)
	}

	// Wait for all goroutines to finish
	cs.collector.Wait()

	return result, result.Error
}

// ScrapeBatch implements the Interface.ScrapeBatch method
func (cs *CollyScrapper) ScrapeBatch(ctx context.Context, urls []string) ([]*ScrapResult, error) {
	if len(urls) == 0 {
		return []*ScrapResult{}, nil
	}

	results := make([]*ScrapResult, len(urls))
	errors := make([]error, len(urls))

	// Process URLs concurrently with reasonable limits
	maxConcurrency := 5
	semaphore := make(chan struct{}, maxConcurrency)

	done := make(chan bool)
	completed := 0

	for i, targetURL := range urls {
		go func(index int, url string) {
			semaphore <- struct{}{} // Acquire
			defer func() {
				<-semaphore // Release
				completed++
				if completed == len(urls) {
					done <- true
				}
			}()

			result, err := cs.Scrape(ctx, url)
			results[index] = result
			errors[index] = err
		}(i, targetURL)
	}

	// Wait for all to complete or context cancellation
	select {
	case <-done:
		// All completed
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Check if any critical errors occurred
	var firstError error
	for _, err := range errors {
		if err != nil && firstError == nil {
			firstError = err
		}
	}

	return results, firstError
}

// IsValidURL implements the Interface.IsValidURL method
func (cs *CollyScrapper) IsValidURL(targetURL string) bool {
	if targetURL == "" {
		return false
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

// NormalizeURL implements the Interface.NormalizeURL method
func (cs *CollyScrapper) NormalizeURL(targetURL string) (string, error) {
	if !cs.IsValidURL(targetURL) {
		return "", fmt.Errorf("invalid URL: %s", targetURL)
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}

	// Convert to lowercase
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Remove default ports
	if u.Scheme == "http" && strings.HasSuffix(u.Host, ":80") {
		u.Host = strings.TrimSuffix(u.Host, ":80")
	} else if u.Scheme == "https" && strings.HasSuffix(u.Host, ":443") {
		u.Host = strings.TrimSuffix(u.Host, ":443")
	}

	// Remove fragment
	u.Fragment = ""

	// Remove trailing slash from path (unless it's just "/")
	if u.Path != "/" && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	return u.String(), nil
}

// SetConfig implements the Interface.SetConfig method
func (cs *CollyScrapper) SetConfig(config *Config) {
	cs.config = config
	if cs.collector != nil {
		cs.collector.SetRequestTimeout(config.Timeout)
	}
}

// GetConfig implements the Interface.GetConfig method
func (cs *CollyScrapper) GetConfig() *Config {
	return cs.config
}

// Close implements the Interface.Close method
func (cs *CollyScrapper) Close() error {
	if cs.collector != nil {
		cs.collector.OnRequest(nil)
		cs.collector.OnHTML("", nil)
		cs.collector.OnResponse(nil)
		cs.collector.OnError(nil)
		cs.collector.OnScraped(nil)
	}
	return nil
}
