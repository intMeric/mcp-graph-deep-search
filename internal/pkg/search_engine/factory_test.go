package search_engine_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/search_engine"
)

var _ = Describe("SearchEngine Factory", func() {
	Describe("NewSearchEngine", func() {
		Context("with SearXNG type", func() {
			It("should create a SearXNG instance", func() {
				engine, err := search_engine.NewSearchEngine(search_engine.SearXNGType)
				Expect(err).NotTo(HaveOccurred())
				Expect(engine).NotTo(BeNil())
			})
		})

		Context("with unknown type", func() {
			It("should return an error", func() {
				engine, err := search_engine.NewSearchEngine("unknown")
				Expect(err).To(HaveOccurred())
				Expect(engine).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("unsupported search engine type"))
			})
		})
	})

})
