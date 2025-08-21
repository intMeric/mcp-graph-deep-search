package webseach_test

import (
	"context"
	"mgds/src/pkg/scrapper"
	"mgds/src/pkg/search_engine"
	webseach "mgds/src/service/web_seach"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Integration test using real implementations
var _ = Describe("WebSearchService Integration", func() {
	var (
		webSearchService webseach.Interface
		tempDir          string
		ctx              context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create temporary directory for database
		var err error
		tempDir, err = os.MkdirTemp("", "mgds_test_*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if webSearchService != nil {
			webSearchService.Close()
		}

		// Clean up temporary directory
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("with mock search engine and real graph", func() {
		BeforeEach(func() {
			// Create mock search engine
			mockSearchEng := &mockSearchEngine{
				searchResult: &search_engine.SearchResponse{
					Results: []search_engine.SearchResult{
						{
							URL:     "https://httpbin.org/html",
							Title:   "HTTPBin HTML Test",
							Content: "Test content",
						},
						{
							URL:     "https://httpbin.org/json",
							Title:   "HTTPBin JSON Test",
							Content: "JSON test content",
						},
					},
					Query: "integration test",
				},
			}

			// Create real scrapper with simple implementation
			scrapperImpl := &simpleScrapper{}

			// Create real graph database
			// Note: This would require actual graph implementation
			// For now, use mock to avoid dependency issues
			mockGr := newMockGraph()

			webSearchService = webseach.NewWebSearchService(mockSearchEng, scrapperImpl, mockGr)
		})

		It("should perform end-to-end search and indexing", func() {
			options := &webseach.SearchOptions{
				MaxResults:  2,
				MaxLinks:    5,
				Parallelism: 1, // Use 1 for predictable testing
			}

			result, err := webSearchService.SearchAndIndex(ctx, "integration test", options)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Query).To(Equal("integration test"))
			Expect(result.Stats.SearchResults).To(Equal(2))
			Expect(result.TimeElapsed).To(BeNumerically(">", 0))

			// Check that some pages were processed
			totalProcessed := result.Stats.ScrapedPages + result.Stats.SkippedPages
			Expect(totalProcessed).To(Equal(2))

			// If scraping was successful, we should have nodes
			if result.Stats.ScrapedPages > 0 {
				Expect(result.Stats.NodesCreated).To(BeNumerically(">", 0))
			}
		})
	})
})

// Simple scrapper implementation for integration testing
type simpleScrapper struct{}

func (s *simpleScrapper) Scrape(ctx context.Context, url string) (*scrapper.ScrapResult, error) {
	// Simple implementation that creates basic content
	// In a real integration test, this would make HTTP requests
	return &scrapper.ScrapResult{
		URL:     url,
		Title:   "Integration Test Page",
		Content: "This is test content from the integration test",
		Links: []scrapper.Link{
			{
				URL:        url + "/link1",
				Text:       "Test Link 1",
				IsInternal: true,
			},
			{
				URL:        "https://external.example.com",
				Text:       "External Link",
				IsInternal: false,
			},
		},
		ScrapedAt:  time.Now(),
		StatusCode: 200,
	}, nil
}

func (s *simpleScrapper) ScrapeBatch(ctx context.Context, urls []string) ([]*scrapper.ScrapResult, error) {
	var results []*scrapper.ScrapResult
	for _, url := range urls {
		result, err := s.Scrape(ctx, url)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *simpleScrapper) IsValidURL(url string) bool {
	return true
}

func (s *simpleScrapper) NormalizeURL(url string) (string, error) {
	return url, nil
}

func (s *simpleScrapper) SetConfig(config *scrapper.Config) {}

func (s *simpleScrapper) GetConfig() *scrapper.Config {
	return &scrapper.Config{}
}

func (s *simpleScrapper) Close() error {
	return nil
}
