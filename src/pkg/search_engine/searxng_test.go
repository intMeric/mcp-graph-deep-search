package search_engine_test

import (
	"context"
	"encoding/json"
	"mgds/src/pkg/search_engine"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SearXNGClient", func() {
	var (
		client search_engine.Interface
		ctx    context.Context
		server *httptest.Server
		config *search_engine.Config
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
		if server != nil {
			server.Close()
		}
	})

	Describe("NewSearXNG", func() {
		Context("with nil config", func() {
			It("should use default config", func() {
				client = search_engine.NewSearXNG(nil)

				Expect(client).NotTo(BeNil())
				Expect(client.GetConfig()).NotTo(BeNil())
				Expect(client.GetConfig().BaseURL).NotTo(BeEmpty())
			})
		})

		Context("with custom config", func() {
			It("should use provided config", func() {
				config = &search_engine.Config{
					BaseURL:   "https://custom.searx.example.com",
					Timeout:   10 * time.Second,
					UserAgent: "test-agent",
				}

				client = search_engine.NewSearXNG(config)

				Expect(client.GetConfig().BaseURL).To(Equal("https://custom.searx.example.com"))
				Expect(client.GetConfig().Timeout).To(Equal(10 * time.Second))
				Expect(client.GetConfig().UserAgent).To(Equal("test-agent"))
			})
		})
	})

	Describe("Search", func() {
		var mockResponse map[string]interface{}

		BeforeEach(func() {
			mockResponse = map[string]interface{}{
				"query": "test query",
				"results": []map[string]interface{}{
					{
						"url":       "https://example.com/1",
						"title":     "Test Result 1",
						"content":   "This is test content 1",
						"engine":    "google",
						"score":     0.95,
						"category":  "general",
						"engines":   []string{"google"},
						"positions": []int{1},
					},
					{
						"url":           "https://example.com/2",
						"title":         "Test Result 2",
						"content":       "This is test content 2",
						"engine":        "bing",
						"score":         0.87,
						"category":      "news",
						"engines":       []string{"bing"},
						"positions":     []int{2},
						"publishedDate": "2024-01-15T10:30:00Z",
					},
				},
				"answers":              []any{},
				"corrections":          []any{},
				"infoboxes":            []any{},
				"suggestions":          []string{"test suggestion"},
				"unresponsive_engines": []any{},
			}

			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/search" {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(mockResponse)
				} else if r.URL.Path == "/config" {
					w.WriteHeader(http.StatusOK)
				}
			}))

			config = &search_engine.Config{
				BaseURL:   server.URL,
				Timeout:   5 * time.Second,
				UserAgent: "test-agent",
			}

			client = search_engine.NewSearXNG(config)
		})

		Context("with valid request", func() {
			It("should return search results", func() {
				req := &search_engine.SearchRequest{
					Query:      "test query",
					Categories: []search_engine.Category{search_engine.CategoryGeneral, search_engine.CategoryNews},
					TimeRange:  search_engine.TimeRangeAll,
					Language:   "en",
					SafeSearch: 1,
					PageNo:     1,
				}

				resp, err := client.Search(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Query).To(Equal("test query"))
				Expect(resp.Results).To(HaveLen(2))
				Expect(resp.TimeElapsed).To(BeNumerically(">", 0))

				// Check first result
				result1 := resp.Results[0]
				Expect(result1.URL).To(Equal("https://example.com/1"))
				Expect(result1.Title).To(Equal("Test Result 1"))
				Expect(result1.Content).To(Equal("This is test content 1"))
				Expect(result1.Score).To(Equal(0.95))
				Expect(result1.Engine).To(Equal("google"))
				Expect(result1.Category).To(Equal(search_engine.CategoryGeneral))

				// Check second result with published date
				result2 := resp.Results[1]
				Expect(result2.URL).To(Equal("https://example.com/2"))
				Expect(result2.PublishedDate).NotTo(BeNil())
			})
		})

		Context("with empty query", func() {
			It("should return error", func() {
				req := &search_engine.SearchRequest{
					Query: "",
				}

				resp, err := client.Search(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("query cannot be empty"))
				Expect(resp).To(BeNil())
			})
		})

		Context("with server error", func() {
			BeforeEach(func() {
				server.Close()
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))

				config.BaseURL = server.URL
				client.SetConfig(config)
			})

			It("should return error", func() {
				req := &search_engine.SearchRequest{
					Query: "test query",
				}

				resp, err := client.Search(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("SearchXNG returned status 500"))
				Expect(resp).To(BeNil())
			})
		})

		Context("with time range short", func() {
			It("should set month parameter", func() {
				var capturedURL string
				server.Close()
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					capturedURL = r.URL.RawQuery
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(mockResponse)
				}))

				config.BaseURL = server.URL
				client.SetConfig(config)

				req := &search_engine.SearchRequest{
					Query:     "test query",
					TimeRange: search_engine.TimeRangeShort,
				}

				_, err := client.Search(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(capturedURL).To(ContainSubstring("time_range=month"))
			})
		})
	})

	Describe("SearchSimple", func() {
		BeforeEach(func() {
			mockResponse := map[string]interface{}{
				"query": "simple test",
				"results": []map[string]interface{}{
					{
						"url":       "https://example.com/simple",
						"title":     "Simple Test",
						"content":   "Simple content",
						"engine":    "google",
						"score":     0.9,
						"category":  "general",
						"engines":   []string{"google"},
						"positions": []int{1},
					},
				},
				"answers":              []any{},
				"corrections":          []any{},
				"infoboxes":            []any{},
				"suggestions":          []string{},
				"unresponsive_engines": []any{},
			}

			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(mockResponse)
			}))

			config = &search_engine.Config{
				BaseURL: server.URL,
				Timeout: 5 * time.Second,
			}

			client = search_engine.NewSearXNG(config)
		})

		Context("with valid query", func() {
			It("should return search results with default settings", func() {
				resp, err := client.SearchSimple(ctx, "simple test")

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Query).To(Equal("simple test"))
				Expect(resp.Results).To(HaveLen(1))
				Expect(resp.Results[0].Title).To(Equal("Simple Test"))
			})
		})
	})

	Describe("IsHealthy", func() {
		Context("when server is healthy", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/config" {
						w.WriteHeader(http.StatusOK)
					}
				}))

				config = &search_engine.Config{
					BaseURL: server.URL,
					Timeout: 5 * time.Second,
				}

				client = search_engine.NewSearXNG(config)
			})

			It("should return no error", func() {
				err := client.IsHealthy(ctx)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when server is unhealthy", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusServiceUnavailable)
				}))

				config = &search_engine.Config{
					BaseURL: server.URL,
					Timeout: 5 * time.Second,
				}

				client = search_engine.NewSearXNG(config)
			})

			It("should return error", func() {
				err := client.IsHealthy(ctx)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("health check returned status 503"))
			})
		})
	})

	Describe("Configuration", func() {
		Context("setting and getting config", func() {
			It("should update configuration", func() {
				client = search_engine.NewSearXNG(nil)

				newConfig := &search_engine.Config{
					BaseURL:   "https://new.searx.example.com",
					Timeout:   15 * time.Second,
					UserAgent: "new-agent",
				}

				client.SetConfig(newConfig)

				retrievedConfig := client.GetConfig()
				Expect(retrievedConfig.BaseURL).To(Equal("https://new.searx.example.com"))
				Expect(retrievedConfig.Timeout).To(Equal(15 * time.Second))
				Expect(retrievedConfig.UserAgent).To(Equal("new-agent"))
			})
		})
	})

	Describe("DefaultConfig", func() {
		It("should return valid default configuration", func() {
			defaultConfig := search_engine.DefaultConfig()

			Expect(defaultConfig).NotTo(BeNil())
			Expect(defaultConfig.BaseURL).To(Equal("https://searx.example.com"))
			Expect(defaultConfig.Timeout).To(Equal(30 * time.Second))
			Expect(defaultConfig.UserAgent).To(Equal("mgds/1.0"))
			Expect(defaultConfig.DefaultLanguage).To(Equal("en"))
			Expect(defaultConfig.DefaultSafeSearch).To(Equal(1))
			Expect(defaultConfig.MaxResults).To(Equal(50))
			Expect(defaultConfig.DefaultTimeRange).To(Equal(search_engine.TimeRangeAll))
		})
	})
})
