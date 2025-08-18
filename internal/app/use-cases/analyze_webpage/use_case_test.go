package analyze_webpage_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/object/webpage"
	"mgds/internal/app/services/link"
	"mgds/internal/app/services/text_analysis"
	"mgds/internal/app/use-cases/analyze_webpage"
	"mgds/internal/pkg/keyword"
	"mgds/internal/pkg/node"
	"mgds/internal/pkg/scrapper"
)

type mockWebpageAnalysisService struct{}

func (m *mockWebpageAnalysisService) AnalyzeWebpage(ctx context.Context, url string, options *scrapper.ScrapingOptions) (webpage.WebpageInterface, error) {
	if url == "invalid-url" {
		return nil, fmt.Errorf("failed to parse URL")
	}

	scrapedData := &scrapper.ScrapedData{
		URL:       url,
		Title:     "Test Page",
		Text:      "This is test content about machine learning and artificial intelligence",
		MetaTags:  map[string]string{"description": "Test page"},
		Links:     []scrapper.Link{{URL: "https://example.com/other", Text: "Other Page"}},
		Images:    []scrapper.Image{},
		ScrapedAt: time.Now(),
	}

	textAnalysis := &text_analysis.TextAnalysisResult{
		Text: "This is test content about machine learning and artificial intelligence",
		Keywords: []keyword.Keyword{
			{Text: "machine", Score: 0.9, Frequency: 1},
			{Text: "learning", Score: 0.9, Frequency: 1},
			{Text: "artificial", Score: 0.8, Frequency: 1},
			{Text: "intelligence", Score: 0.8, Frequency: 1},
		},
	}

	// Create webpage object with original URL
	return webpage.Build(scrapedData, textAnalysis, url), nil
}

func (m *mockWebpageAnalysisService) Close() error {
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
		useCase             analyze_webpage.AnalyzeWebpageUseCase
		mockWebpageAnalysis *mockWebpageAnalysisService
		mockLinkSvc         *mockLinkService
		mockDB              *mockDatabase
		ctx                 context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockWebpageAnalysis = &mockWebpageAnalysisService{}
		mockLinkSvc = &mockLinkService{}
		mockDB = &mockDatabase{}

		useCase = analyze_webpage.NewAnalyzeWebpageUseCase(
			mockWebpageAnalysis,
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
				Expect(response.ExtractedKeywords).To(ContainElement("machine"))
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
				Expect(err.Error()).To(ContainSubstring("failed to analyze webpage"))
			})
		})
	})
})
