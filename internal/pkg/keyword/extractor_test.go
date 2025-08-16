package keyword_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/keyword"
)

var _ = Describe("Keyword Extractor", func() {
	var (
		extractor keyword.Extractor
		ctx       context.Context
		cancel    context.CancelFunc
	)

	BeforeEach(func() {
		var err error
		extractor, err = keyword.NewExtractor()
		Expect(err).NotTo(HaveOccurred())
		Expect(extractor).NotTo(BeNil())

		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	})

	AfterEach(func() {
		cancel()
		if extractor != nil {
			extractor.Close()
		}
	})

	Describe("ExtractKeywords", func() {
		Context("with simple text", func() {
			It("should extract keywords successfully", func() {
				text := "The quick brown fox jumps over the lazy dog. Technology innovation artificial intelligence machine learning."
				options := keyword.DefaultOptions()

				keywords, err := extractor.ExtractKeywords(ctx, text, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(keywords).NotTo(BeEmpty())
				Expect(len(keywords)).To(BeNumerically("<=", options.MaxKeywords))
			})

			It("should respect max keywords limit", func() {
				text := "Apple banana cherry date elderberry fig grape honeydew kiwi lemon mango nectarine orange peach"
				options := keyword.DefaultOptions()
				options.MaxKeywords = 5

				keywords, err := extractor.ExtractKeywords(ctx, text, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(len(keywords)).To(BeNumerically("<=", 5))
			})

			It("should filter by minimum length", func() {
				text := "I am a big dog with a small cat and two tiny mice"
				options := keyword.DefaultOptions()
				options.MinLength = 4

				keywords, err := extractor.ExtractKeywords(ctx, text, options)

				Expect(err).NotTo(HaveOccurred())
				for _, keyword := range keywords {
					Expect(len(keyword)).To(BeNumerically(">=", 4))
				}
			})

			It("should remove stop words when enabled", func() {
				text := "The quick brown fox jumps over the lazy dog"
				options := keyword.DefaultOptions()
				options.RemoveStopWords = true

				keywords, err := extractor.ExtractKeywords(ctx, text, options)

				Expect(err).NotTo(HaveOccurred())

				// Should not contain "the" which is definitely a stop word
				for _, keyword := range keywords {
					Expect(keyword).NotTo(Equal("the"))
				}

				// Should contain meaningful words like "quick", "brown", "fox"
				meaningfulWords := []string{"quick", "brown", "fox", "jumps", "lazy", "dog"}
				foundMeaningful := 0
				for _, keyword := range keywords {
					for _, meaningful := range meaningfulWords {
						if keyword == meaningful {
							foundMeaningful++
							break
						}
					}
				}
				Expect(foundMeaningful).To(BeNumerically(">", 0))
			})

			It("should include stop words when disabled", func() {
				text := "The quick brown fox"
				options := keyword.DefaultOptions()
				options.RemoveStopWords = false

				keywords, err := extractor.ExtractKeywords(ctx, text, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(keywords).To(ContainElement("the"))
			})
		})

		Context("with empty or invalid input", func() {
			It("should handle empty text gracefully", func() {
				text := ""
				options := keyword.DefaultOptions()

				keywords, err := extractor.ExtractKeywords(ctx, text, options)

				Expect(err).NotTo(HaveOccurred())
				Expect(keywords).To(BeEmpty())
			})

			It("should handle nil options by using defaults", func() {
				text := "Test text for keyword extraction"

				keywords, err := extractor.ExtractKeywords(ctx, text, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(keywords).NotTo(BeEmpty())
			})
		})
	})

	Describe("ExtractKeywordsWithScores", func() {
		It("should return keywords with scores", func() {
			text := "Machine learning artificial intelligence technology innovation machine learning"
			options := keyword.DefaultOptions()

			keywords, err := extractor.ExtractKeywordsWithScores(ctx, text, options)

			Expect(err).NotTo(HaveOccurred())
			Expect(keywords).NotTo(BeEmpty())

			for _, kw := range keywords {
				Expect(kw.Text).NotTo(BeEmpty())
				Expect(kw.Frequency).To(BeNumerically(">", 0))
				Expect(kw.Score).To(BeNumerically(">", 0))
			}

			// "machine" and "learning" should appear twice
			found := false
			for _, kw := range keywords {
				if (kw.Text == "machine" || kw.Text == "learning") && kw.Frequency == 2 {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
		})

		It("should sort keywords by score descending", func() {
			text := "apple apple apple banana banana cherry"
			options := keyword.DefaultOptions()

			keywords, err := extractor.ExtractKeywordsWithScores(ctx, text, options)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(keywords)).To(BeNumerically(">=", 2))

			// Should be sorted by score (frequency) descending
			for i := 1; i < len(keywords); i++ {
				Expect(keywords[i-1].Score).To(BeNumerically(">=", keywords[i].Score))
			}
		})
	})

	Describe("Close", func() {
		It("should close without error", func() {
			err := extractor.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
