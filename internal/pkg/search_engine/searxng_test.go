package search_engine_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/search_engine"
)

var _ = Describe("SearXNG", func() {
	var (
		searxng search_engine.SearchEngine
		server  *httptest.Server
		ctx     context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("NewSearXNG", func() {
		Context("with valid config", func() {
			It("should create a SearXNG instance", func() {
				config := &search_engine.SearXNGConfig{
					Host:     "http://localhost:8080",
					Language: "en",
				}
				engine, err := search_engine.NewSearXNG(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(engine).NotTo(BeNil())
			})
		})

		Context("with invalid URL", func() {
			It("should return an error", func() {
				config := &search_engine.SearXNGConfig{
					Host:     "://invalid-url",
					Language: "en",
				}
				engine, err := search_engine.NewSearXNG(config)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(search_engine.ErrUrlIsInvalid))
				Expect(engine).To(BeNil())
			})
		})

		Context("with empty language", func() {
			It("should default to 'en'", func() {
				config := &search_engine.SearXNGConfig{
					Host:     "http://localhost:8080",
					Language: "",
				}
				engine, err := search_engine.NewSearXNG(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(engine).NotTo(BeNil())
			})
		})
	})

	Describe("Search", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"query": "test query",
					"results": [
						{
							"title": "Test Result 1",
							"url": "https://example.com/1",
							"engine": "test-engine",
							"score": 0.9,
							"category": "general",
							"tags": ["test"]
						},
						{
							"title": "Test Result 2",
							"url": "https://example.com/2",
							"engine": "test-engine",
							"score": 0.8,
							"category": "general"
						}
					]
				}`))
			}))

			config := &search_engine.SearXNGConfig{
				Host:     server.URL,
				Language: "en",
			}
			var err error
			searxng, err = search_engine.NewSearXNG(config)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("with valid query", func() {
			It("should return search results", func() {
				response, err := searxng.Search(ctx, "test query")
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Query).To(Equal("test query"))
				Expect(response.Results).To(HaveLen(2))

				Expect(response.Results[0].Title).To(Equal("Test Result 1"))
				Expect(response.Results[0].URL).To(Equal("https://example.com/1"))
				Expect(response.Results[0].Score).To(Equal(90.0))
				Expect(response.Results[0].Engine).To(Equal("test-engine"))
				Expect(response.Results[0].Category).To(Equal("general"))
				Expect(response.Results[0].Tags).To(Equal([]string{"test"}))
				Expect(response.Results[0].Query).To(Equal("test query"))
				Expect(response.Results[0].RequestDate).NotTo(BeEmpty())

				Expect(response.Results[1].Score).To(Equal(80.0))
			})

			It("should sort results by score descending", func() {
				response, err := searxng.Search(ctx, "test query")
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Results).To(HaveLen(2))
				Expect(response.Results[0].Score).To(BeNumerically(">", response.Results[1].Score))
			})
		})
	})

	Describe("Configuration methods", func() {
		BeforeEach(func() {
			config := &search_engine.SearXNGConfig{
				Host:     "http://localhost:8080",
				Language: "en",
			}
			var err error
			searxng, err = search_engine.NewSearXNG(config)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("SetTimeRange", func() {
			It("should return a new instance with updated time range", func() {
				newEngine := searxng.SetTimeRange("day")
				Expect(newEngine).NotTo(BeIdenticalTo(searxng))
			})
		})

		Describe("SetCategory", func() {
			It("should return a new instance with valid category", func() {
				newEngine := searxng.SetCategory("news")
				Expect(newEngine).NotTo(BeIdenticalTo(searxng))
			})

			It("should use default category for invalid input", func() {
				newEngine := searxng.SetCategory("invalid")
				Expect(newEngine).NotTo(BeIdenticalTo(searxng))
			})
		})

		Describe("SetLanguage", func() {
			It("should return a new instance with updated language", func() {
				newEngine := searxng.SetLanguage("fr")
				Expect(newEngine).NotTo(BeIdenticalTo(searxng))
			})
		})
	})
})
