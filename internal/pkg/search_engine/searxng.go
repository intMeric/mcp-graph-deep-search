package search_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type SearXNG struct {
	config    *SearXNGConfig
	timeRange string
	category  string
	language  string
}

func NewSearXNG(config *SearXNGConfig) (*SearXNG, error) {
	if _, err := url.Parse(config.Host); err != nil {
		return nil, ErrUrlIsInvalid
	}

	language := config.Language
	if language == "" {
		language = "en"
	}

	return &SearXNG{
		config:   config,
		language: language,
		category: "general",
	}, nil
}

func (s *SearXNG) Search(ctx context.Context, query string) (*SearchResponse, error) {
	searchURL, err := s.buildSearchURL(query)
	if err != nil {
		return nil, err
	}

	log.Printf("search url %s", searchURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var searxngResponse struct {
		Query   string `json:"query"`
		Results []struct {
			Title         string   `json:"title"`
			URL           string   `json:"url"`
			PublishedDate string   `json:"publishedDate,omitempty"`
			ImgSrc        string   `json:"img_src,omitempty"`
			Engine        string   `json:"engine"`
			Score         float64  `json:"score"`
			Category      string   `json:"category"`
			Tags          []string `json:"tags,omitempty"`
		} `json:"results"`
	}

	err = json.NewDecoder(response.Body).Decode(&searxngResponse)
	if err != nil {
		return nil, err
	}

	var results []Result
	now := time.Now()
	formattedTime := now.Format(time.RFC3339)

	for _, res := range searxngResponse.Results {
		results = append(results, Result{
			Title:         res.Title,
			URL:           res.URL,
			PublishedDate: res.PublishedDate,
			ImgSrc:        res.ImgSrc,
			Engine:        res.Engine,
			Score:         res.Score * 100,
			Category:      res.Category,
			Query:         query,
			RequestDate:   formattedTime,
			Tags:          res.Tags,
		})
	}

	searchResponse := &SearchResponse{
		Query:   searxngResponse.Query,
		Results: results,
	}

	sort.Slice(searchResponse.Results, func(i, j int) bool {
		return searchResponse.Results[i].Score > searchResponse.Results[j].Score
	})

	return searchResponse, nil
}

func (s *SearXNG) SetTimeRange(timeRange string) SearchEngine {
	newSearXNG := *s
	newSearXNG.timeRange = timeRange
	return &newSearXNG
}

func (s *SearXNG) SetCategory(category string) SearchEngine {
	if err := s.validCategory(category); err != nil {
		log.Printf("Invalid category %s, using default 'general'", category)
		category = "general"
	}
	newSearXNG := *s
	newSearXNG.category = category
	return &newSearXNG
}

func (s *SearXNG) SetLanguage(lang string) SearchEngine {
	newSearXNG := *s
	newSearXNG.language = lang
	return &newSearXNG
}

func (s *SearXNG) buildSearchURL(query string) (string, error) {
	queryParam := url.QueryEscape(query)
	
	switch s.category {
	case "news":
		return fmt.Sprintf("%s/search?q=%s&time_range=%s&format=json&language=%s&category_news=", 
			s.config.Host, queryParam, s.timeRange, s.language), nil
	case "rs":
		return fmt.Sprintf("%s/search?q=%s&time_range=%s&format=json&language=%s&category_social+media=", 
			s.config.Host, queryParam, s.timeRange, s.language), nil
	case "general":
		return fmt.Sprintf("%s/search?q=%s&time_range=%s&format=json&language=%s", 
			s.config.Host, queryParam, s.timeRange, s.language), nil
	default:
		return "", ErrInvalidCategory
	}
}

func (s *SearXNG) validCategory(category string) error {
	switch category {
	case "news", "rs", "general":
		return nil
	default:
		return ErrInvalidCategory
	}
}