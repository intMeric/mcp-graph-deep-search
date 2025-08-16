package text_analysis_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/text_analysis"
	"mgds/internal/pkg/keyword"
	"mgds/internal/pkg/pii"
)

type mockPIIExtractor struct {
	result     *pii.Result
	shouldFail bool
	closed     bool
}

func (m *mockPIIExtractor) ExtractPII(ctx context.Context, text string) (*pii.Result, error) {
	if m.shouldFail {
		return nil, errors.New("PII extraction failed")
	}
	return m.result, nil
}

func (m *mockPIIExtractor) Close() error {
	m.closed = true
	return nil
}

type mockKeywordExtractor struct {
	keywords   []keyword.Keyword
	shouldFail bool
	closed     bool
}

func (m *mockKeywordExtractor) ExtractKeywords(ctx context.Context, text string, options *keyword.Options) ([]string, error) {
	var result []string
	for _, kw := range m.keywords {
		result = append(result, kw.Text)
	}
	return result, nil
}

func (m *mockKeywordExtractor) ExtractKeywordsWithScores(ctx context.Context, text string, options *keyword.Options) ([]keyword.Keyword, error) {
	if m.shouldFail {
		return nil, errors.New("keyword extraction failed")
	}
	return m.keywords, nil
}

func (m *mockKeywordExtractor) Close() error {
	m.closed = true
	return nil
}

var _ = Describe("TextAnalysisService", func() {
	var (
		service          text_analysis.TextAnalysisService
		ctx              context.Context
		piiExtractor     *mockPIIExtractor
		keywordExtractor *mockKeywordExtractor
	)

	BeforeEach(func() {
		ctx = context.Background()
		piiExtractor = &mockPIIExtractor{}
		keywordExtractor = &mockKeywordExtractor{}
		service = text_analysis.NewTextAnalysisService(piiExtractor, keywordExtractor)
	})

	AfterEach(func() {
		if service != nil {
			service.Close()
		}
	})

	Describe("AnalyzeText", func() {
		Context("with valid text containing PII and keywords", func() {
			It("should extract both PII and keywords", func() {
				text := "Contact john.doe@example.com for more information about machine learning"

				piiResult := &pii.Result{
					Total: 1,
					Entities: []pii.Entity{
						{
							Type:  pii.PIITypeEmail,
							Value: "john.doe@example.com",
							Count: 1,
						},
					},
					Stats: map[string]int{"email": 1},
				}
				piiExtractor.result = piiResult

				keywords := []keyword.Keyword{
					{Text: "contact", Frequency: 1, Score: 0.8},
					{Text: "information", Frequency: 1, Score: 0.7},
					{Text: "machine", Frequency: 1, Score: 0.9},
					{Text: "learning", Frequency: 1, Score: 0.9},
				}
				keywordExtractor.keywords = keywords

				result, err := service.AnalyzeText(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Text).To(Equal(text))
				Expect(result.HasPII()).To(BeTrue())
				Expect(result.HasKeywords()).To(BeTrue())
				Expect(result.PIIResult.Total).To(Equal(1))
				Expect(len(result.Keywords)).To(Equal(4))
			})
		})

		Context("when PII extraction fails", func() {
			It("should return an error", func() {
				text := "Some text"
				piiExtractor.shouldFail = true

				result, err := service.AnalyzeText(ctx, text)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

		Context("when keyword extraction fails", func() {
			It("should return an error", func() {
				text := "Some text"
				piiResult := &pii.Result{
					Total:    0,
					Entities: []pii.Entity{},
					Stats:    map[string]int{},
				}
				piiExtractor.result = piiResult
				keywordExtractor.shouldFail = true

				result, err := service.AnalyzeText(ctx, text)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Close", func() {
		It("should close both extractors without error", func() {
			err := service.Close()

			Expect(err).NotTo(HaveOccurred())
			Expect(piiExtractor.closed).To(BeTrue())
			Expect(keywordExtractor.closed).To(BeTrue())
		})
	})
})
