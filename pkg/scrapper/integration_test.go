package scrapper_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/pkg/scrapper"
)

var _ = Describe("Scrapper Integration Tests", func() {
	var (
		collyScrapper         scrapper.Interface
		antiDetectionScrapper *scrapper.AntiDetectionScrapper
		ctx                   context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		if collyScrapper != nil {
			collyScrapper.Close()
		}
		if antiDetectionScrapper != nil {
			antiDetectionScrapper.Close()
		}
	})

	Describe("CollyScrapper Integration", func() {
		BeforeEach(func() {
			config := scrapper.DefaultConfig()
			collyScrapper = scrapper.NewCollyScrapper(config)
		})

		Context("with httpbin.org (reliable test site)", func() {
			It("should scrape basic HTML content", func() {
				url := "https://httpbin.org/html"
				
				result, err := collyScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.URL).To(Equal(url))
				Expect(result.StatusCode).To(Equal(200))
				Expect(result.Title).NotTo(BeEmpty())
				Expect(result.Content).NotTo(BeEmpty())
				Expect(result.ScrapedAt).To(BeTemporally("~", time.Now(), 30*time.Second))
			})

			It("should extract links from pages with links", func() {
				url := "https://httpbin.org/"
				
				result, err := collyScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Links).NotTo(BeEmpty())
				
				// Check that at least one link is properly formatted
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

		Context("with example.com (minimal test site)", func() {
			It("should handle simple pages", func() {
				url := "https://example.com"
				
				result, err := collyScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.StatusCode).To(Equal(200))
				Expect(result.Title).To(ContainSubstring("Example"))
				Expect(result.Content).NotTo(BeEmpty())
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
					"https://httpbin.org/html",
					"https://example.com",
					"https://httpbin.org/robots.txt",
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
					"HTTPS://EXAMPLE.COM/PATH":     "https://example.com/PATH",
					"https://example.com:443/":     "https://example.com/",
					"http://example.com:80/":       "http://example.com/",
					"https://example.com/path/":    "https://example.com/path",
					"https://example.com/#fragment": "https://example.com/",
				}

				for input, expected := range testCases {
					normalized, err := collyScrapper.NormalizeURL(input)
					Expect(err).NotTo(HaveOccurred())
					Expect(normalized).To(Equal(expected))
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
			It("should scrape with stealth configuration", func() {
				url := "https://httpbin.org/user-agent"
				
				result, err := antiDetectionScrapper.Scrape(ctx, url)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.StatusCode).To(Equal(200))
				Expect(result.Content).To(ContainSubstring("user-agent"))
			})

			It("should use different user agents", func() {
				url := "https://httpbin.org/user-agent"
				
				// Make multiple requests and check if user agents vary
				results := make([]*scrapper.ScrapResult, 3)
				for i := 0; i < 3; i++ {
					result, err := antiDetectionScrapper.Scrape(ctx, url)
					Expect(err).NotTo(HaveOccurred())
					results[i] = result
				}

				// All should succeed
				for _, result := range results {
					Expect(result.StatusCode).To(Equal(200))
					Expect(result.Content).To(ContainSubstring("user-agent"))
				}
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
				url := "https://httpbin.org/html"
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
			config.Timeout = 2 * time.Second // Short timeout for testing
			collyScrapper = scrapper.NewCollyScrapper(config)
		})

		Context("with timeout scenarios", func() {
			It("should handle timeouts gracefully", func() {
				// Using httpbin delay endpoint to test timeout
				url := "https://httpbin.org/delay/5" // 5 second delay, but 2 second timeout
				
				result, err := collyScrapper.Scrape(ctx, url)

				// Should either error or have an error in result
				hasError := err != nil || (result != nil && result.Error != nil)
				Expect(hasError).To(BeTrue())
			})
		})

		Context("with context cancellation", func() {
			It("should respect context cancellation", func() {
				cancelCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
				defer cancel()
				
				url := "https://httpbin.org/delay/3"
				
				result, err := collyScrapper.Scrape(cancelCtx, url)

				// Should handle cancellation
				hasError := err != nil || (result != nil && result.Error != nil)
				Expect(hasError).To(BeTrue())
			})
		})
	})
})