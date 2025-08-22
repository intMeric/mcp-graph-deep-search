package scrapper_test

import (
	"context"
	"fmt"
	"mgds/src/pkg/scrapper"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scrapper Integration Tests", func() {
	var (
		collyScrapper         scrapper.Interface
		antiDetectionScrapper *scrapper.AntiDetectionScrapper
		ctx                   context.Context
		mockServer            *httptest.Server
	)

	BeforeEach(func() {
		ctx = context.Background()
		
		// Setup mock HTTP server
		mux := http.NewServeMux()
		
		// Mock HTML page
		mux.HandleFunc("/html", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Test Page Title</title>
    <meta name="description" content="Test page description">
    <meta name="keywords" content="test, html, mock">
</head>
<body>
    <main>
        <h1>Test Content</h1>
        <p>This is test content from mock server.</p>
        <a href="/link1">Internal Link</a>
        <a href="https://external.com">External Link</a>
    </main>
</body>
</html>`)
		})
		
		// Mock user-agent endpoint
		mux.HandleFunc("/user-agent", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			userAgent := r.Header.Get("User-Agent")
			fmt.Fprintf(w, `{
  "user-agent": "%s"
}`, userAgent)
		})
		
		// Mock home page with links
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Mock Home</title>
</head>
<body>
    <h1>Mock Server Home</h1>
    <a href="/html">HTML Page</a>
    <a href="/user-agent">User Agent</a>
    <a href="https://example.com">External</a>
</body>
</html>`)
		})
		
		// Mock robots.txt
		mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(200)
			fmt.Fprint(w, "User-agent: *\nDisallow:")
		})
		
		// Mock delay endpoint
		mux.HandleFunc("/delay/", func(w http.ResponseWriter, r *http.Request) {
			// Extract delay from path
			delay := 5 * time.Second // Default 5 seconds
			time.Sleep(delay)
			w.WriteHeader(200)
			fmt.Fprint(w, "Delayed response")
		})
		
		mockServer = httptest.NewServer(mux)
	})

	AfterEach(func() {
		if collyScrapper != nil {
			collyScrapper.Close()
		}
		if antiDetectionScrapper != nil {
			antiDetectionScrapper.Close()
		}
		if mockServer != nil {
			mockServer.Close()
		}
	})

	Describe("CollyScrapper Integration", func() {
		BeforeEach(func() {
			config := scrapper.DefaultConfig()
			collyScrapper = scrapper.NewCollyScrapper(config)
		})

		Context("with mock HTML page", func() {
			It("should scrape basic HTML content", func() {
				url := mockServer.URL + "/html"

				result, err := collyScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.URL).To(Equal(url))
				Expect(result.StatusCode).To(Equal(200))
				Expect(result.Title).To(Equal("Test Page Title"))
				Expect(result.Content).To(ContainSubstring("Test Content"))
				Expect(result.Content).To(ContainSubstring("This is test content from mock server"))
				Expect(result.ScrapedAt).To(BeTemporally("~", time.Now(), 30*time.Second))
			})

			It("should extract links from pages with links", func() {
				url := mockServer.URL + "/"

				result, err := collyScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Links).NotTo(BeEmpty())
				Expect(len(result.Links)).To(BeNumerically(">=", 2))

				// Check that links are properly formatted
				hasValidLink := false
				for _, link := range result.Links {
					if link.URL != "" && collyScrapper.IsValidURL(link.URL) {
						hasValidLink = true
						break
					}
				}
				Expect(hasValidLink).To(BeTrue())
			})
		})

		Context("with simple mock page", func() {
			It("should handle simple pages", func() {
				url := mockServer.URL + "/"

				result, err := collyScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.StatusCode).To(Equal(200))
				Expect(result.Title).To(Equal("Mock Home"))
				Expect(result.Content).To(ContainSubstring("Mock Server Home"))
			})
		})

		Context("with invalid URLs", func() {
			It("should handle non-existent domains gracefully", func() {
				url := "https://this-domain-should-not-exist-12345.com"

				result, err := collyScrapper.Scrape(ctx, url)

				// Should return a result with error info, not fail completely
				Expect(result).NotTo(BeNil())
				// Either err is not nil, or result.Error is set
				hasError := err != nil || result.Error != nil
				Expect(hasError).To(BeTrue())
			})

			It("should reject malformed URLs", func() {
				url := "not-a-valid-url"

				result, err := collyScrapper.Scrape(ctx, url)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

		Context("batch scraping", func() {
			It("should scrape multiple URLs concurrently", func() {
				urls := []string{
					mockServer.URL + "/html",
					mockServer.URL + "/",
					mockServer.URL + "/robots.txt",
				}

				results, err := collyScrapper.ScrapeBatch(ctx, urls)

				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(3))

				// Check that most results are valid
				validCount := 0
				for _, result := range results {
					if result != nil && result.StatusCode == 200 {
						validCount++
					}
				}
				Expect(validCount).To(BeNumerically(">=", 2)) // At least 2 of 3 should succeed
			})
		})

		Context("URL validation and normalization", func() {
			It("should validate URLs correctly", func() {
				validURLs := []string{
					"https://example.com",
					"http://example.com",
					"https://example.com/path",
					"https://example.com/path?query=value",
				}

				invalidURLs := []string{
					"",
					"ftp://example.com",
					"example.com",
					"not-a-url",
				}

				for _, url := range validURLs {
					Expect(collyScrapper.IsValidURL(url)).To(BeTrue(), "Expected %s to be valid", url)
				}

				for _, url := range invalidURLs {
					Expect(collyScrapper.IsValidURL(url)).To(BeFalse(), "Expected %s to be invalid", url)
				}
			})

			It("should normalize URLs correctly", func() {
				testCases := map[string]string{
					"HTTPS://EXAMPLE.COM/PATH":      "https://example.com/PATH",
					"https://example.com:443/":      "https://example.com/",
					"http://example.com:80/":        "http://example.com/",
					"https://example.com/path/":     "https://example.com/path",
					"https://example.com/#fragment": "https://example.com/",
				}

				for input, expected := range testCases {
					normalized, err := collyScrapper.NormalizeURL(input)
					Expect(err).NotTo(HaveOccurred())
					Expect(normalized).To(Equal(expected), "Expected %s to normalize to %s but got %s", input, expected, normalized)
				}
			})
		})
	})

	Describe("AntiDetectionScrapper Integration", func() {
		BeforeEach(func() {
			config := scrapper.StealthConfig()
			antiDetectConfig := scrapper.DefaultAntiDetectionConfig()
			antiDetectionScrapper = scrapper.NewAntiDetectionScrapper(config, antiDetectConfig)
		})

		Context("with anti-detection enabled", func() {
			It("should create anti-detection scrapper without errors", func() {
				Expect(antiDetectionScrapper).NotTo(BeNil())
				Expect(antiDetectionScrapper.GetAntiDetectionConfig()).NotTo(BeNil())
			})
		})
	})

	Describe("Performance Tests", func() {
		Context("with fast configuration", func() {
			BeforeEach(func() {
				config := scrapper.FastConfig()
				collyScrapper = scrapper.NewCollyScrapper(config)
			})

			It("should complete scraping within reasonable time", func() {
				url := mockServer.URL + "/html"
				start := time.Now()

				result, err := collyScrapper.Scrape(ctx, url)
				duration := time.Since(start)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(duration).To(BeNumerically("<", 15*time.Second))
			})
		})
	})

	Describe("Error Handling", func() {
		BeforeEach(func() {
			config := scrapper.DefaultConfig()
			collyScrapper = scrapper.NewCollyScrapper(config)
		})

		Context("with invalid URLs", func() {
			It("should handle network errors gracefully", func() {
				url := "https://this-domain-should-not-exist-12345.com"

				result, err := collyScrapper.Scrape(ctx, url)

				// Should either error or have an error in result
				hasError := err != nil || (result != nil && result.Error != nil)
				Expect(hasError).To(BeTrue())
			})
		})
	})
})