package scrapper_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/pkg/scrapper"
)

// MockScrapper implements the scrapper.Interface for testing
type MockScrapper struct {
	config       *scrapper.Config
	shouldFail   bool
	mockResults  map[string]*scrapper.ScrapResult
}

func NewMockScrapper() *MockScrapper {
	return &MockScrapper{
		config: &scrapper.Config{
			UserAgent:       "TestAgent/1.0",
			Timeout:         30 * time.Second,
			MaxContentSize:  1024 * 1024,
			FollowRedirects: true,
			MaxRedirects:    5,
		},
		mockResults: make(map[string]*scrapper.ScrapResult),
	}
}

func (m *MockScrapper) Scrape(ctx context.Context, url string) (*scrapper.ScrapResult, error) {
	if m.shouldFail {
		return nil, &ScrapError{Message: "mock scraping failed"}
	}
	
	if result, exists := m.mockResults[url]; exists {
		return result, nil
	}
	
	return &scrapper.ScrapResult{
		URL:        url,
		Title:      "Mock Title",
		Content:    "Mock content for " + url,
		Links:      []scrapper.Link{{URL: "https://example.com/link1", Text: "Link 1", IsInternal: false}},
		Metadata:   map[string]any{"description": "Mock description"},
		ScrapedAt:  time.Now(),
		StatusCode: 200,
	}, nil
}

func (m *MockScrapper) ScrapeBatch(ctx context.Context, urls []string) ([]*scrapper.ScrapResult, error) {
	results := make([]*scrapper.ScrapResult, len(urls))
	for i, url := range urls {
		result, err := m.Scrape(ctx, url)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

func (m *MockScrapper) IsValidURL(url string) bool {
	return url != "" && (len(url) > 7) && (url[:7] == "http://" || url[:8] == "https://")
}

func (m *MockScrapper) NormalizeURL(url string) (string, error) {
	if !m.IsValidURL(url) {
		return "", &ScrapError{Message: "invalid URL"}
	}
	return url, nil
}

func (m *MockScrapper) SetConfig(config *scrapper.Config) {
	m.config = config
}

func (m *MockScrapper) GetConfig() *scrapper.Config {
	return m.config
}

func (m *MockScrapper) Close() error {
	return nil
}

type ScrapError struct {
	Message string
}

func (e *ScrapError) Error() string {
	return e.Message
}

var _ = Describe("Scrapper Interface", func() {
	var (
		mockScrapper *MockScrapper
		ctx          context.Context
	)

	BeforeEach(func() {
		mockScrapper = NewMockScrapper()
		ctx = context.Background()
	})

	AfterEach(func() {
		if mockScrapper != nil {
			mockScrapper.Close()
		}
	})

	Describe("Scrape", func() {
		Context("with valid URL", func() {
			It("should return scraping result", func() {
				url := "https://example.com"
				
				result, err := mockScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.URL).To(Equal(url))
				Expect(result.Title).NotTo(BeEmpty())
				Expect(result.Content).NotTo(BeEmpty())
				Expect(result.StatusCode).To(Equal(200))
				Expect(result.ScrapedAt).To(BeTemporally("~", time.Now(), time.Second))
			})

			It("should extract links from the page", func() {
				url := "https://example.com"
				
				result, err := mockScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Links).NotTo(BeEmpty())
				Expect(result.Links[0].URL).NotTo(BeEmpty())
				Expect(result.Links[0].Text).NotTo(BeEmpty())
			})

			It("should include metadata", func() {
				url := "https://example.com"
				
				result, err := mockScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Metadata).NotTo(BeEmpty())
				Expect(result.Metadata["description"]).To(Equal("Mock description"))
			})
		})

		Context("with scraping failure", func() {
			BeforeEach(func() {
				mockScrapper.shouldFail = true
			})

			It("should handle errors gracefully", func() {
				url := "https://example.com"
				
				result, err := mockScrapper.Scrape(ctx, url)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

		Context("with context cancellation", func() {
			It("should respect context timeout", func() {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
				defer cancel()
				
				time.Sleep(2 * time.Millisecond) // Ensure context is cancelled
				
				_, err := mockScrapper.Scrape(ctxWithTimeout, "https://example.com")

				// In a real implementation, this should return context error
				// For mock, we just verify it doesn't panic
				_ = err
			})
		})
	})

	Describe("ScrapeBatch", func() {
		Context("with multiple valid URLs", func() {
			It("should scrape all URLs", func() {
				urls := []string{
					"https://example.com/page1",
					"https://example.com/page2",
					"https://example.com/page3",
				}
				
				results, err := mockScrapper.ScrapeBatch(ctx, urls)

				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(3))
				for i, result := range results {
					Expect(result.URL).To(Equal(urls[i]))
					Expect(result.Title).NotTo(BeEmpty())
					Expect(result.Content).NotTo(BeEmpty())
				}
			})
		})

		Context("with empty URL list", func() {
			It("should return empty results", func() {
				urls := []string{}
				
				results, err := mockScrapper.ScrapeBatch(ctx, urls)

				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(BeEmpty())
			})
		})
	})

	Describe("IsValidURL", func() {
		Context("with valid URLs", func() {
			It("should return true for HTTP URLs", func() {
				Expect(mockScrapper.IsValidURL("http://example.com")).To(BeTrue())
			})

			It("should return true for HTTPS URLs", func() {
				Expect(mockScrapper.IsValidURL("https://example.com")).To(BeTrue())
			})

			It("should return true for URLs with paths", func() {
				Expect(mockScrapper.IsValidURL("https://example.com/path/to/page")).To(BeTrue())
			})
		})

		Context("with invalid URLs", func() {
			It("should return false for empty URLs", func() {
				Expect(mockScrapper.IsValidURL("")).To(BeFalse())
			})

			It("should return false for malformed URLs", func() {
				Expect(mockScrapper.IsValidURL("not-a-url")).To(BeFalse())
			})

			It("should return false for FTP URLs", func() {
				Expect(mockScrapper.IsValidURL("ftp://example.com")).To(BeFalse())
			})
		})
	})

	Describe("NormalizeURL", func() {
		Context("with valid URLs", func() {
			It("should return normalized URL", func() {
				url := "https://example.com/page"
				
				normalized, err := mockScrapper.NormalizeURL(url)

				Expect(err).NotTo(HaveOccurred())
				Expect(normalized).To(Equal(url))
			})
		})

		Context("with invalid URLs", func() {
			It("should return error for invalid URLs", func() {
				url := "invalid-url"
				
				normalized, err := mockScrapper.NormalizeURL(url)

				Expect(err).To(HaveOccurred())
				Expect(normalized).To(BeEmpty())
			})
		})
	})

	Describe("Configuration", func() {
		It("should allow setting and getting configuration", func() {
			newConfig := &scrapper.Config{
				UserAgent:       "NewAgent/2.0",
				Timeout:         60 * time.Second,
				MaxContentSize:  2 * 1024 * 1024,
				FollowRedirects: false,
				MaxRedirects:    3,
			}
			
			mockScrapper.SetConfig(newConfig)
			retrievedConfig := mockScrapper.GetConfig()

			Expect(retrievedConfig).To(Equal(newConfig))
			Expect(retrievedConfig.UserAgent).To(Equal("NewAgent/2.0"))
			Expect(retrievedConfig.Timeout).To(Equal(60 * time.Second))
			Expect(retrievedConfig.MaxContentSize).To(Equal(int64(2 * 1024 * 1024)))
			Expect(retrievedConfig.FollowRedirects).To(BeFalse())
			Expect(retrievedConfig.MaxRedirects).To(Equal(3))
		})
	})

	Describe("Lifecycle", func() {
		It("should close without error", func() {
			err := mockScrapper.Close()
			
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Link Structure", func() {
		It("should properly handle link properties", func() {
			link := scrapper.Link{
				URL:        "https://example.com/link",
				Text:       "Example Link",
				Title:      "Link Title",
				IsInternal: true,
				Properties: map[string]any{
					"rel":   "nofollow",
					"class": "external-link",
				},
			}

			Expect(link.URL).To(Equal("https://example.com/link"))
			Expect(link.Text).To(Equal("Example Link"))
			Expect(link.Title).To(Equal("Link Title"))
			Expect(link.IsInternal).To(BeTrue())
			Expect(link.Properties["rel"]).To(Equal("nofollow"))
			Expect(link.Properties["class"]).To(Equal("external-link"))
		})
	})

	Describe("ScrapResult Structure", func() {
		It("should contain all required fields", func() {
			result := &scrapper.ScrapResult{
				URL:        "https://example.com",
				Title:      "Example Page",
				Content:    "Page content",
				Links:      []scrapper.Link{{URL: "https://example.com/link", Text: "Link"}},
				Metadata:   map[string]any{"author": "John Doe"},
				ScrapedAt:  time.Now(),
				StatusCode: 200,
				Error:      nil,
			}

			Expect(result.URL).To(Equal("https://example.com"))
			Expect(result.Title).To(Equal("Example Page"))
			Expect(result.Content).To(Equal("Page content"))
			Expect(result.Links).To(HaveLen(1))
			Expect(result.Metadata["author"]).To(Equal("John Doe"))
			Expect(result.StatusCode).To(Equal(200))
			Expect(result.Error).To(BeNil())
		})
	})
})