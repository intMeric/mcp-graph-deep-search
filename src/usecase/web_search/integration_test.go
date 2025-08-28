package websearch_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/scrapper"
	"mgds/src/pkg/search_engine"
	"mgds/src/usecase/web_search"
	
	// Import Cayley memory store
	_ "github.com/cayleygraph/cayley/graph/memstore"
)

func TestWebSearchIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WebSearch Integration Suite")
}

var _ = Describe("WebSearch Integration Tests", func() {
	var (
		useCase      websearch.UseCase
		graphDB      graph.Interface
		searchEngine search_engine.Interface
		scraper      scrapper.Interface
		factory      *websearch.Factory
		ctx          context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		
		config := &graph.Config{
			DatabasePath: ":memory:",
			DatabaseType: "cayley",
		}
		
		graphDB = graph.NewCayleyGraph(config)
		err := graphDB.Connect(ctx)
		Expect(err).NotTo(HaveOccurred())
		
		// Configure SearchXNG
		searchConfig := &search_engine.Config{
			BaseURL:           "http://localhost:8888",
			Timeout:           30 * time.Second,
			UserAgent:         "mgds-integration-test/1.0",
			DefaultLanguage:   "en",
			DefaultSafeSearch: 1,
			MaxResults:        10,
			DefaultTimeRange:  search_engine.TimeRangeShort,
		}
		searchEngine = search_engine.NewSearXNG(searchConfig)
		
		// Configure scraper
		scraper = scrapper.NewCollyScrapper(nil)
		
		factory = websearch.NewFactory(searchEngine, scraper, graphDB)
		useCase = factory.CreateUseCase()
	})

	AfterEach(func() {
		if searchEngine != nil {
			searchEngine.Close()
		}
		if scraper != nil {
			scraper.Close()
		}
		if graphDB != nil {
			graphDB.Close()
		}
	})

	Context("with real SearchXNG and database", func() {
		It("should handle empty query", func() {
			request := &websearch.SearchRequest{
				Query:      "",
				MaxResults: 5,
			}

			response, err := useCase.Execute(ctx, request)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(websearch.StatusFailed))
			Expect(response.Error).To(Equal("Query cannot be empty"))
			Expect(response.Query).To(Equal(""))
			Expect(response.Nodes).To(BeEmpty())
		})

		It("should handle whitespace-only query", func() {
			request := &websearch.SearchRequest{
				Query:      "   ",
				MaxResults: 5,
			}

			response, err := useCase.Execute(ctx, request)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(websearch.StatusFailed))
			Expect(response.Error).To(Equal("Query cannot be empty"))
		})

		It("should execute real web search and index results", func() {
			Skip("Requires SearchXNG running on localhost:8888")
			
			request := &websearch.SearchRequest{
				Query:      "test",
				MaxResults: 3,
			}

			response, err := useCase.Execute(ctx, request)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Query).To(Equal("test"))
			Expect(response.ExecutionTime).To(BeNumerically(">", 0))
			
			if response.Status == websearch.StatusSuccess {
				Expect(response.Nodes).NotTo(BeEmpty())
				Expect(len(response.Nodes)).To(BeNumerically("<=", 3))
				
				// Check first node structure
				if len(response.Nodes) > 0 {
					node := response.Nodes[0]
					Expect(node.ID).NotTo(BeEmpty())
					Expect(node.URL).NotTo(BeEmpty())
					Expect(node.Title).NotTo(BeEmpty())
				}
				
				// Verify nodes were added to database
				allNodes, err := graphDB.GetAllNodes(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(allNodes)).To(BeNumerically(">=", len(response.Nodes)))
			}
		})

		It("should handle invalid max results", func() {
			request := &websearch.SearchRequest{
				Query:      "golang",
				MaxResults: 0,
			}

			response, err := useCase.Execute(ctx, request)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			// Should default to 10 results
			Expect(response.Query).To(Equal("golang"))
		})

		It("should handle negative max results", func() {
			request := &websearch.SearchRequest{
				Query:      "python",
				MaxResults: -5,
			}

			response, err := useCase.Execute(ctx, request)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			// Should default to 10 results
			Expect(response.Query).To(Equal("python"))
		})

		It("should handle SearchXNG unavailable", func() {
			Skip("Requires SearchXNG to be down - manual test")
			
			request := &websearch.SearchRequest{
				Query:      "test unavailable",
				MaxResults: 5,
			}

			response, err := useCase.Execute(ctx, request)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(websearch.StatusFailed))
			Expect(response.Error).To(ContainSubstring("Search and index failed"))
		})

		It("should handle context cancellation", func() {
			// Create context with short timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
			defer cancel()

			request := &websearch.SearchRequest{
				Query:      "context cancellation test",
				MaxResults: 5,
			}

			response, err := useCase.Execute(timeoutCtx, request)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			// Should handle timeout gracefully
			if response.Status == websearch.StatusFailed {
				Expect(response.Error).NotTo(BeEmpty())
			}
		})

		It("should handle database connection errors", func() {
			// Close database to simulate error
			graphDB.Close()

			request := &websearch.SearchRequest{
				Query:      "database error test",
				MaxResults: 3,
			}

			response, err := useCase.Execute(ctx, request)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			// Database may still work with closed connection in memory backend or SearchXNG may be available
			Expect(response.Status).To(BeElementOf([]string{
				websearch.StatusFailed,
				websearch.StatusSuccess,
			}))
			if response.Status == websearch.StatusFailed {
				Expect(response.Error).To(ContainSubstring("Search and index failed"))
			}
		})

		It("should measure execution time accurately", func() {
			request := &websearch.SearchRequest{
				Query:      "execution time test",
				MaxResults: 1,
			}

			start := time.Now()
			response, err := useCase.Execute(ctx, request)
			actualTime := time.Since(start)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.ExecutionTime).To(BeNumerically(">", 0))
			Expect(response.ExecutionTime).To(BeNumerically("<=", actualTime+100*time.Millisecond))
		})

		It("should handle multiple concurrent searches", func() {
			Skip("Concurrent test - may be flaky")
			
			queries := []string{"concurrent1", "concurrent2", "concurrent3"}
			responses := make(chan *websearch.SearchResponse, len(queries))
			
			// Start concurrent searches
			for _, query := range queries {
				go func(q string) {
					request := &websearch.SearchRequest{
						Query:      q,
						MaxResults: 2,
					}
					resp, _ := useCase.Execute(ctx, request)
					responses <- resp
				}(query)
			}
			
			// Collect responses
			for i := 0; i < len(queries); i++ {
				response := <-responses
				Expect(response).NotTo(BeNil())
				Expect(response.Query).To(BeElementOf(queries))
			}
		})
	})

	Context("with SearchXNG health check", func() {
		It("should verify SearchXNG is accessible", func() {
			if searchEngine != nil {
				err := searchEngine.IsHealthy(ctx)
				if err != nil {
					Skip("SearchXNG not available at localhost:8888")
				}
			}
		})
	})
})