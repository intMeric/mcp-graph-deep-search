package text_analysis_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/text_analysis"
	"mgds/internal/pkg/keyword"
)

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
		keywordExtractor *mockKeywordExtractor
		service          text_analysis.TextAnalysisService
		ctx              context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		keywordExtractor = &mockKeywordExtractor{}
		service = text_analysis.NewTextAnalysisService(keywordExtractor)
	})

	AfterEach(func() {
		if service != nil {
			service.Close()
		}
	})

	Describe("AnalyzeText", func() {
		Context("with valid text containing keywords", func() {
			It("should extract keywords", func() {
				text := "This is a text about machine learning and artificial intelligence"

				keywords := []keyword.Keyword{
					{Text: "machine", Frequency: 1, Score: 0.9},
					{Text: "learning", Frequency: 1, Score: 0.9},
					{Text: "artificial", Frequency: 1, Score: 0.8},
					{Text: "intelligence", Frequency: 1, Score: 0.8},
				}
				keywordExtractor.keywords = keywords

				result, err := service.AnalyzeText(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Text).To(Equal(text))
				Expect(result.HasKeywords()).To(BeTrue())
				Expect(len(result.Keywords)).To(Equal(4))
			})
		})

		Context("with text containing no keywords", func() {
			It("should return empty keywords", func() {
				text := "Some simple text"
				keywordExtractor.keywords = []keyword.Keyword{}

				result, err := service.AnalyzeText(ctx, text)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Text).To(Equal(text))
				Expect(result.HasKeywords()).To(BeFalse())
				Expect(len(result.Keywords)).To(Equal(0))
			})
		})

		Context("when keyword extraction fails", func() {
			It("should return an error", func() {
				text := "Some text"
				keywordExtractor.shouldFail = true

				result, err := service.AnalyzeText(ctx, text)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Close", func() {
		It("should close keyword extractor without error", func() {
			err := service.Close()

			Expect(err).NotTo(HaveOccurred())
			Expect(keywordExtractor.closed).To(BeTrue())
		})
	})
})