package pii_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/pii"
)

var _ = Describe("Factory", func() {
	Describe("NewPIIExtractor", func() {
		It("should create extractor successfully", func() {
			extractor, err := pii.NewPIIExtractor()

			Expect(err).NotTo(HaveOccurred())
			Expect(extractor).NotTo(BeNil())

			// Clean up
			extractor.Close()
		})
	})

	Describe("NewExtractor", func() {
		It("should create extractor successfully", func() {
			extractor, err := pii.NewExtractor()

			Expect(err).NotTo(HaveOccurred())
			Expect(extractor).NotTo(BeNil())

			// Clean up
			extractor.Close()
		})
	})
})
