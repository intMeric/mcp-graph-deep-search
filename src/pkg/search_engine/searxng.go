package search_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SearXNGClient implements the Interface for SearchXNG
type SearXNGClient struct {
	config     *Config
	httpClient *http.Client
}

// searxngResponse represents the JSON response from SearchXNG API
type searxngResponse struct {
	Query   string `json:"query"`
	Results []struct {
		URL           string `json:"url"`
		Title         string `json:"title"`
		Content       string `json:"content"`
		Engine        string `json:"engine"`
		ParsedURL     []string `json:"parsed_url"`
		Template      string `json:"template,omitempty"`
		Engines       []string `json:"engines"`
		Positions     []int    `json:"positions"`
		Score         float64  `json:"score"`
		Category      string   `json:"category"`
		PublishedDate string   `json:"publishedDate,omitempty"`
	} `json:"results"`
	Answers            []any `json:"answers"`
	Corrections        []any `json:"corrections"`
	Infoboxes          []any `json:"infoboxes"`
	Suggestions        []string      `json:"suggestions"`
	UnresponsiveEngines []any `json:"unresponsive_engines"`
}

// NewSearXNG creates a new SearchXNG client
func NewSearXNG(config *Config) *SearXNGClient {
	if config == nil {
		config = DefaultConfig()
	}

	return &SearXNGClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Search performs a search with the given request
func (s *SearXNGClient) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	startTime := time.Now()

	// Build the search URL
	searchURL, err := s.buildSearchURL(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build search URL: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("User-Agent", s.config.UserAgent)
	httpReq.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SearchXNG returned status %d", resp.StatusCode)
	}

	// Parse response
	var searxngResp searxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&searxngResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our format
	return s.convertResponse(&searxngResp, req, time.Since(startTime)), nil
}

// SearchSimple performs a simple search with default settings
func (s *SearXNGClient) SearchSimple(ctx context.Context, query string) (*SearchResponse, error) {
	req := &SearchRequest{
		Query:      query,
		Categories: []Category{CategoryGeneral, CategoryNews},
		TimeRange:  s.config.DefaultTimeRange,
		Language:   s.config.DefaultLanguage,
		SafeSearch: s.config.DefaultSafeSearch,
		PageNo:     1,
		MaxResults: s.config.MaxResults,
	}

	return s.Search(ctx, req)
}

// IsHealthy checks if SearchXNG is accessible
func (s *SearXNGClient) IsHealthy(ctx context.Context) error {
	healthURL := strings.TrimSuffix(s.config.BaseURL, "/") + "/config"

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SearchXNG health check returned status %d", resp.StatusCode)
	}

	return nil
}

// SetConfig updates the configuration
func (s *SearXNGClient) SetConfig(config *Config) {
	s.config = config
	s.httpClient.Timeout = config.Timeout
}

// GetConfig returns the current configuration
func (s *SearXNGClient) GetConfig() *Config {
	return s.config
}

// Close closes the client and cleans up resources
func (s *SearXNGClient) Close() error {
	s.httpClient.CloseIdleConnections()
	return nil
}

// buildSearchURL constructs the search URL with parameters
func (s *SearXNGClient) buildSearchURL(req *SearchRequest) (string, error) {
	baseURL := strings.TrimSuffix(s.config.BaseURL, "/")
	u, err := url.Parse(baseURL + "/search")
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("q", req.Query)
	params.Set("format", "json")

	// Set categories
	if len(req.Categories) > 0 {
		categories := make([]string, len(req.Categories))
		for i, cat := range req.Categories {
			categories[i] = string(cat)
		}
		params.Set("categories", strings.Join(categories, ","))
	}

	// Set time range
	if req.TimeRange != "" {
		switch req.TimeRange {
		case TimeRangeShort:
			params.Set("time_range", "month")
		case TimeRangeAll:
			// Don't set time_range parameter for all time
		}
	}

	// Set language
	language := req.Language
	if language == "" {
		language = s.config.DefaultLanguage
	}
	if language != "" {
		params.Set("language", language)
	}

	// Set safe search
	safeSearch := req.SafeSearch
	if safeSearch == 0 {
		safeSearch = s.config.DefaultSafeSearch
	}
	if safeSearch > 0 {
		params.Set("safesearch", fmt.Sprintf("%d", safeSearch))
	}

	// Set page number
	if req.PageNo > 0 {
		params.Set("pageno", fmt.Sprintf("%d", req.PageNo))
	}

	u.RawQuery = params.Encode()
	return u.String(), nil
}

// convertResponse converts SearchXNG response to our format
func (s *SearXNGClient) convertResponse(resp *searxngResponse, req *SearchRequest, elapsed time.Duration) *SearchResponse {
	results := make([]SearchResult, 0, len(resp.Results))

	for _, result := range resp.Results {
		searchResult := SearchResult{
			URL:      result.URL,
			Title:    result.Title,
			Content:  result.Content,
			Snippet:  result.Content, // SearchXNG uses content as snippet
			Score:    result.Score,
			Engine:   result.Engine,
			Category: Category(result.Category),
		}

		// Parse published date if available
		if result.PublishedDate != "" {
			if parsedTime, err := time.Parse(time.RFC3339, result.PublishedDate); err == nil {
				searchResult.PublishedDate = &parsedTime
			}
		}

		// Add metadata
		searchResult.Metadata = map[string]any{
			"engines":   result.Engines,
			"positions": result.Positions,
			"template":  result.Template,
		}

		results = append(results, searchResult)
	}

	// Calculate pagination
	currentPage := req.PageNo
	if currentPage == 0 {
		currentPage = 1
	}

	maxResults := req.MaxResults
	if maxResults == 0 {
		maxResults = s.config.MaxResults
	}

	response := &SearchResponse{
		Results:     results,
		Query:       req.Query,
		TimeElapsed: elapsed,
		CurrentPage: currentPage,
		HasNextPage: len(results) >= maxResults,
		HasPrevPage: currentPage > 1,
	}

	if response.HasNextPage {
		response.NextPage = currentPage + 1
	}
	if response.HasPrevPage {
		response.PrevPage = currentPage - 1
	}

	return response
}