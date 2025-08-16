package scrapper

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interface", func() {
	Describe("ScrapedData", func() {
		It("should have proper structure", func() {
			scrapedData := &ScrapedData{
				URL:       "https://example.com/page",
				Title:     "Example Page",
				Text:      "Some text content",
				ScrapedAt: time.Now(),
			}

			Expect(scrapedData.URL).To(Equal("https://example.com/page"))
			Expect(scrapedData.Title).To(Equal("Example Page"))
			Expect(scrapedData.Text).To(Equal("Some text content"))
		})
	})

	Describe("Link", func() {
		It("should have proper structure", func() {
			link := &Link{
				URL:  "https://example.com/link",
				Text: "Example Link",
				Rel:  "nofollow",
			}

			Expect(link.URL).To(Equal("https://example.com/link"))
			Expect(link.Text).To(Equal("Example Link"))
			Expect(link.Rel).To(Equal("nofollow"))
		})
	})
})