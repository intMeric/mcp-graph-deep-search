package webpage_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/object/webpage"
	"mgds/internal/pkg/scrapper"
)

var _ = Describe("Webpage", func() {
	var (
		testWebpage webpage.WebpageInterface
		scrapedData *scrapper.ScrapedData
	)

	BeforeEach(func() {
		scrapedData = &scrapper.ScrapedData{
			URL:      "https://example.com/test",
			Title:    "Test Page",
			Text:     "This is a test page with some content",
			MetaTags: map[string]string{"description": "Test page description"},
			Links: []scrapper.Link{
				{URL: "https://example.com/link1", Text: "Link 1"},
				{URL: "https://example.com/link2", Text: "Link 2"},
			},
			Images: []scrapper.Image{},
		}

		testWebpage = webpage.Build(scrapedData, "")
	})

	Describe("Build", func() {
		It("should create a webpage interface with scraped URL when no original URL provided", func() {
			webpageObj := webpage.Build(scrapedData, "")

			Expect(webpageObj).NotTo(BeNil())
			Expect(webpageObj.GetURL()).To(Equal("https://example.com/test"))
			Expect(webpageObj.GetTitle()).To(Equal("Test Page"))
			Expect(webpageObj.GetText()).To(Equal("This is a test page with some content"))
		})

		It("should use original URL when provided", func() {
			originalURL := "https://original.com/page"
			webpageObj := webpage.Build(scrapedData, originalURL)

			Expect(webpageObj.GetURL()).To(Equal(originalURL))
			Expect(webpageObj.GetTitle()).To(Equal("Test Page"))
		})

		It("should return an object that implements WebpageInterface", func() {
			webpageObj := webpage.Build(scrapedData, "")
			
			// Test that it implements the interface by calling interface methods
			Expect(webpageObj.GetURL()).To(Equal("https://example.com/test"))
			Expect(webpageObj.GetTitle()).To(Equal("Test Page"))
			Expect(webpageObj.GetText()).NotTo(BeEmpty())
			Expect(webpageObj.ToNode()).NotTo(BeNil())
			Expect(webpageObj.ToDocument()).NotTo(BeNil())
		})
	})

	Describe("GetID", func() {
		It("should return a valid ID based on URL", func() {
			id := testWebpage.GetID()
			Expect(id).NotTo(BeEmpty())
			Expect(id).To(ContainSubstring("example.com"))
		})
	})

	Describe("ToNode", func() {
		It("should create a valid node", func() {
			node := testWebpage.ToNode()

			Expect(node).NotTo(BeNil())
			Expect(node.Type).To(Equal("URL"))
			Expect(node.SubType).To(Equal("webpage"))
			Expect(node.DisplayName).To(Equal("Test Page"))
			Expect(node.ID).NotTo(BeEmpty())
		})
	})

	Describe("GetLinkNodes", func() {
		It("should return link nodes from scraped data", func() {
			linkNodes := testWebpage.GetLinkNodes()

			Expect(linkNodes).To(HaveLen(2))
			Expect(linkNodes[0].Type).To(Equal("URL"))
			Expect(linkNodes[0].SubType).To(Equal("webpage"))
			Expect(linkNodes[0].DisplayName).To(Equal("Link 1"))
			Expect(linkNodes[1].DisplayName).To(Equal("Link 2"))
		})
	})

	Describe("HasKeywords", func() {
		It("should always return false since we removed keyword extraction", func() {
			Expect(testWebpage.HasKeywords()).To(BeFalse())
		})
	})

	Describe("GetScrapedData", func() {
		It("should return the original scraped data", func() {
			data := testWebpage.GetScrapedData()

			Expect(data).To(Equal(scrapedData))
			Expect(data.URL).To(Equal("https://example.com/test"))
			Expect(data.Title).To(Equal("Test Page"))
			Expect(data.Text).To(Equal("This is a test page with some content"))
		})
	})

	Describe("ToDocument", func() {
		It("should create a document with expected structure", func() {
			document := testWebpage.ToDocument()

			Expect(document).NotTo(BeNil())
			Expect(document["url"]).To(Equal("https://example.com/test"))
			Expect(document["title"]).To(Equal("Test Page"))
			Expect(document["text"]).To(Equal("This is a test page with some content"))
			Expect(document["meta_tags"]).NotTo(BeNil())
			Expect(document["links"]).NotTo(BeNil())
			Expect(document["scraped_at"]).NotTo(BeNil())
		})

		It("should contain all expected keys", func() {
			document := testWebpage.ToDocument()
			expectedKeys := []string{"url", "title", "text", "meta_tags", "links", "scraped_at"}
			for _, key := range expectedKeys {
				Expect(document).To(HaveKey(key))
			}
		})

		It("should not contain keywords field", func() {
			document := testWebpage.ToDocument()
			Expect(document).NotTo(HaveKey("keywords"))
		})
	})

	Describe("GetTitle", func() {
		It("should return the page title", func() {
			Expect(testWebpage.GetTitle()).To(Equal("Test Page"))
		})
	})

	Describe("GetText", func() {
		It("should return the page text content", func() {
			Expect(testWebpage.GetText()).To(Equal("This is a test page with some content"))
		})
	})

	Describe("GetURL", func() {
		It("should return the page URL", func() {
			Expect(testWebpage.GetURL()).To(Equal("https://example.com/test"))
		})
	})
})