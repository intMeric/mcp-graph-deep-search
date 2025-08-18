package webpage_analysis_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/text_analysis"
	"mgds/internal/app/services/webpage_analysis"
	"mgds/internal/pkg/keyword"
	"mgds/internal/pkg/scrapper"
)

type mockWebScraper struct{}

func (m *mockWebScraper) Scrape(ctx context.Context, url string, options *scrapper.ScrapingOptions) (*scrapper.ScrapedData, error) {
	if url == "invalid-url" {
		return nil, fmt.Errorf("failed to parse URL")
	}

	return &scrapper.ScrapedData{
		URL:       url,
		Title:     "Test Page",
		Text:      "This is test content with email@example.com and phone 123-456-7890",
		MetaTags:  map[string]string{"description": "Test page"},
		Links:     []scrapper.Link{{URL: "https://example.com/other", Text: "Other Page"}},
		Images:    []scrapper.Image{},
		ScrapedAt: time.Now(),
	}, nil
}

func (m *mockWebScraper) ScrapeMultiple(ctx context.Context, urls []string, options *scrapper.ScrapingOptions) ([]*scrapper.ScrapedData, error) {
	return nil, nil
}

func (m *mockWebScraper) SetUserAgent(userAgent string) {}

func (m *mockWebScraper) SetTimeout(timeout time.Duration) {}

func (m *mockWebScraper) Close() error {
	return nil
}

type mockTextAnalysisService struct{}

func (m *mockTextAnalysisService) AnalyzeText(ctx context.Context, text string) (*text_analysis.TextAnalysisResult, error) {
	return &text_analysis.TextAnalysisResult{
		Text: text,
		Keywords: []keyword.Keyword{
			{Text: "test", Score: 0.9, Frequency: 1},
			{Text: "content", Score: 0.8, Frequency: 2},
			{Text: "example", Score: 0.7, Frequency: 1},
		},
	}, nil
}

func (m *mockTextAnalysisService) Close() error {
	return nil
}

var _ = Describe("WebpageAnalysisService", func() {
	var (
		service          webpage_analysis.WebpageAnalysisService
		mockScraper      *mockWebScraper
		mockTextAnalysis *mockTextAnalysisService
		ctx              context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockScraper = &mockWebScraper{}
		mockTextAnalysis = &mockTextAnalysisService{}
		service = webpage_analysis.NewWebpageAnalysisService(mockScraper, mockTextAnalysis)
	})

	AfterEach(func() {
		if service != nil {
			service.Close()
		}
	})

	Describe("AnalyzeWebpage", func() {
		Context("with valid URL", func() {
			It("should analyze webpage successfully", func() {
				webpageObj, err := service.AnalyzeWebpage(ctx, "https://example.com/test", nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(webpageObj).NotTo(BeNil())
				Expect(webpageObj.ToNode()).NotTo(BeNil())
				Expect(webpageObj.GetLinkNodes()).To(HaveLen(1))
			})

			It("should create correct webpage node", func() {
				webpageObj, err := service.AnalyzeWebpage(ctx, "https://example.com/test", nil)

				Expect(err).NotTo(HaveOccurred())
				webpageNode := webpageObj.ToNode()
				Expect(webpageNode.Type).To(Equal("URL"))
				Expect(webpageNode.SubType).To(Equal("webpage"))
				Expect(webpageNode.DisplayName).To(Equal("Test Page"))
				Expect(webpageNode.ID).To(ContainSubstring("example.com"))
			})

			It("should create correct link nodes", func() {
				webpageObj, err := service.AnalyzeWebpage(ctx, "https://example.com/test", nil)

				Expect(err).NotTo(HaveOccurred())
				linkNodes := webpageObj.GetLinkNodes()
				Expect(linkNodes).To(HaveLen(1))
				Expect(linkNodes[0].Type).To(Equal("URL"))
				Expect(linkNodes[0].SubType).To(Equal("webpage"))
				Expect(linkNodes[0].DisplayName).To(Equal("Other Page"))
			})

			It("should include text analysis results", func() {
				webpageObj, err := service.AnalyzeWebpage(ctx, "https://example.com/test", nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(webpageObj.HasKeywords()).To(BeTrue())
			})
		})

		Context("with invalid input", func() {
			It("should return error for scraping failure", func() {
				_, err := service.AnalyzeWebpage(ctx, "invalid-url", nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse URL"))
			})
		})
	})
})
