package keyword_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/keyword"
)

var _ = Describe("Options", func() {
	Describe("DefaultOptions", func() {
		It("should return sensible defaults", func() {
			options := keyword.DefaultOptions()

			Expect(options).NotTo(BeNil())
			Expect(options.MinLength).To(Equal(3))
			Expect(options.MaxKeywords).To(Equal(20))
			Expect(options.RemoveStopWords).To(BeTrue())
		})
	})

	Describe("Keyword struct", func() {
		It("should have proper structure", func() {
			kw := keyword.Keyword{
				Text:      "test",
				Frequency: 5,
				Score:     0.8,
			}

			Expect(kw.Text).To(Equal("test"))
			Expect(kw.Frequency).To(Equal(5))
			Expect(kw.Score).To(Equal(0.8))
		})
	})
})
