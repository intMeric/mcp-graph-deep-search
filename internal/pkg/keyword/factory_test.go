package keyword_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/keyword"
)

var _ = Describe("Factory", func() {
	Describe("NewExtractor", func() {
		It("should create extractor successfully", func() {
			extractor, err := keyword.NewExtractor()

			Expect(err).NotTo(HaveOccurred())
			Expect(extractor).NotTo(BeNil())

			// Clean up
			extractor.Close()
		})
	})
})
