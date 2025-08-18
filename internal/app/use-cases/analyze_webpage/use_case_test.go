package analyze_webpage_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/link"
	"mgds/internal/app/use-cases/analyze_webpage"
	"mgds/internal/pkg/node"
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
		Text:      "This is test content about machine learning and artificial intelligence",
		MetaTags:  map[string]string{"description": "Test page"},
		Links:     []scrapper.Link{{URL: "https://example.com/other", Text: "Other Page"}},
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

type mockLinkService struct {
	createdLinks []struct {
		source, target *node.Node
		relType        string
	}
}

func (m *mockLinkService) CreateLink(ctx context.Context, source, target *node.Node, relType string) (*link.LinkResponse, error) {
	m.createdLinks = append(m.createdLinks, struct {
		source, target *node.Node
		relType        string
	}{source, target, relType})

	return &link.LinkResponse{
		SourceCreated: true,
		TargetCreated: true,
	}, nil
}

func (m *mockLinkService) Close() error {
	return nil
}

type mockDatabase struct {
	storedDocuments map[string]any
}

func (m *mockDatabase) Insert(ctx context.Context, collection, id string, document any) error {
	if m.storedDocuments == nil {
		m.storedDocuments = make(map[string]any)
	}
	key := collection + "_" + id
	m.storedDocuments[key] = document
	return nil
}

func (m *mockDatabase) FindByID(ctx context.Context, collection, id string, dest any) error {
	return nil
}

func (m *mockDatabase) Update(ctx context.Context, collection, id string, update any) error {
	return nil
}

func (m *mockDatabase) Delete(ctx context.Context, collection, id string) error {
	return nil
}

func (m *mockDatabase) Close(ctx context.Context) error {
	return nil
}


var _ = Describe("AnalyzeWebpageUseCase", func() {
	var (
		useCase     analyze_webpage.AnalyzeWebpageUseCase
		mockScraper *mockWebScraper
		mockLinkSvc *mockLinkService
		mockDB      *mockDatabase
		ctx         context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockScraper = &mockWebScraper{}
		mockLinkSvc = &mockLinkService{}
		mockDB = &mockDatabase{}

		useCase = analyze_webpage.NewAnalyzeWebpageUseCase(
			mockScraper,
			mockLinkSvc,
			mockDB,
		)
	})

	Describe("Execute", func() {
		Context("with valid URL", func() {
			It("should analyze webpage successfully", func() {
				request := &analyze_webpage.AnalyzeWebpageRequest{
					URL: "https://example.com/test",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.URL).To(Equal("https://example.com/test"))
				Expect(response.DocumentID).To(Equal("example.com/test"))
				Expect(response.Title).To(Equal("Test Page"))
				Expect(response.Text).To(Equal("This is test content about machine learning and artificial intelligence"))
				Expect(response.RelationsCreated).To(BeNumerically(">", 0))
			})

			It("should store document in database", func() {
				request := &analyze_webpage.AnalyzeWebpageRequest{
					URL: "https://example.com/test",
				}

				_, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(mockDB.storedDocuments).To(HaveKey("URL_example.com/test"))
			})

			It("should create graph relations", func() {
				request := &analyze_webpage.AnalyzeWebpageRequest{
					URL: "https://example.com/test",
				}

				_, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(mockLinkSvc.createdLinks).NotTo(BeEmpty())
				Expect(mockLinkSvc.createdLinks[0].relType).To(Equal("navigation_link"))
			})
		})

		Context("with invalid input", func() {
			It("should return error for nil request", func() {
				_, err := useCase.Execute(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid request"))
			})

			It("should return error for empty URL", func() {
				request := &analyze_webpage.AnalyzeWebpageRequest{URL: ""}

				_, err := useCase.Execute(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("URL is required"))
			})

			It("should return error for invalid URL format", func() {
				request := &analyze_webpage.AnalyzeWebpageRequest{URL: "invalid-url"}

				_, err := useCase.Execute(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to scrape webpage"))
			})
		})
	})
})
