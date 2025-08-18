package webpage_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/object/webpage"
	"mgds/internal/app/services/text_analysis"
	"mgds/internal/pkg/keyword"
	"mgds/internal/pkg/scrapper"
)

var _ = Describe("Webpage", func() {
	var (
		testWebpage  *webpage.Webpage
		scrapedData  *scrapper.ScrapedData
		textAnalysis *text_analysis.TextAnalysisResult
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

		textAnalysis = &text_analysis.TextAnalysisResult{
			Text: scrapedData.Text,
			Keywords: []keyword.Keyword{
				{Text: "test", Score: 0.8},
				{Text: "content", Score: 0.6},
			},
		}

		testWebpage = webpage.NewWebpage(scrapedData, textAnalysis)
	})

	Describe("NewWebpage", func() {
		It("should create a webpage with scraped data and text analysis", func() {
			Expect(testWebpage.ScrapedData).To(Equal(scrapedData))
			Expect(testWebpage.TextAnalysis).To(Equal(textAnalysis))
			Expect(testWebpage.Links).To(HaveLen(2))
		})
	})

	Describe("Build", func() {
		It("should create a webpage interface with scraped URL when no original URL provided", func() {
			webpageObj := webpage.Build(scrapedData, textAnalysis, "")

			Expect(webpageObj).NotTo(BeNil())
			Expect(webpageObj.GetURL()).To(Equal("https://example.com/test"))
			Expect(webpageObj.GetTitle()).To(Equal("Test Page"))
			Expect(webpageObj.GetText()).To(Equal("This is a test page with some content"))
		})

		It("should use original URL when provided", func() {
			originalURL := "https://original.com/page"
			webpageObj := webpage.Build(scrapedData, textAnalysis, originalURL)

			Expect(webpageObj.GetURL()).To(Equal(originalURL))
			Expect(webpageObj.GetTitle()).To(Equal("Test Page"))
		})

		It("should return an object that implements WebpageInterface", func() {
			webpageObj := webpage.Build(scrapedData, textAnalysis, "")

			var _ webpage.WebpageInterface = webpageObj
		})
	})

	Describe("GetID", func() {
		It("should generate consistent URL-based ID", func() {
			id := testWebpage.GetID()
			Expect(id).To(Equal("example.com/test"))
		})
	})

	Describe("GetURL", func() {
		It("should return the webpage URL", func() {
			Expect(testWebpage.GetURL()).To(Equal("https://example.com/test"))
		})
	})

	Describe("GetTitle", func() {
		It("should return the webpage title", func() {
			Expect(testWebpage.GetTitle()).To(Equal("Test Page"))
		})
	})

	Describe("GetText", func() {
		It("should return the webpage text content", func() {
			Expect(testWebpage.GetText()).To(Equal("This is a test page with some content"))
		})
	})

	Describe("ToNode", func() {
		It("should create a node with correct properties", func() {
			node := testWebpage.ToNode()

			Expect(node.Type).To(Equal("URL"))
			Expect(node.SubType).To(Equal("webpage"))
			Expect(node.DisplayName).To(Equal("Test Page"))
			Expect(node.ID).To(Equal("example.com/test"))
		})

		Context("when title is empty", func() {
			BeforeEach(func() {
				scrapedData.Title = ""
				testWebpage = webpage.NewWebpage(scrapedData, textAnalysis)
			})

			It("should use URL as display name", func() {
				node := testWebpage.ToNode()
				Expect(node.DisplayName).To(Equal("https://example.com/test"))
			})
		})
	})

	Describe("GetLinkNodes", func() {
		It("should return nodes for all links", func() {
			linkNodes := testWebpage.GetLinkNodes()

			Expect(linkNodes).To(HaveLen(2))
			Expect(linkNodes[0].DisplayName).To(Equal("Link 1"))
			Expect(linkNodes[1].DisplayName).To(Equal("Link 2"))
		})
	})

	Describe("HasKeywords", func() {
		It("should return true when keywords exist", func() {
			Expect(testWebpage.HasKeywords()).To(BeTrue())
		})

		Context("when no keywords exist", func() {
			BeforeEach(func() {
				textAnalysis.Keywords = []keyword.Keyword{}
				testWebpage = webpage.NewWebpage(scrapedData, textAnalysis)
			})

			It("should return false", func() {
				Expect(testWebpage.HasKeywords()).To(BeFalse())
			})
		})
	})


	Describe("ToDocument", func() {
		It("should return a document with all webpage data", func() {
			document := testWebpage.ToDocument()

			Expect(document).NotTo(BeNil())
			Expect(document["url"]).To(Equal("https://example.com/test"))
			Expect(document["title"]).To(Equal("Test Page"))
			Expect(document["text"]).To(Equal("This is a test page with some content"))
			Expect(document["meta_tags"]).NotTo(BeNil())
			Expect(document["keywords"]).NotTo(BeNil())
			Expect(document["links"]).NotTo(BeNil())
			Expect(document["images"]).NotTo(BeNil())
			Expect(document["scraped_at"]).NotTo(BeNil())
		})

		It("should contain all expected keys", func() {
			document := testWebpage.ToDocument()

			expectedKeys := []string{"url", "title", "text", "meta_tags", "keywords", "links", "images", "scraped_at"}
			for _, key := range expectedKeys {
				Expect(document).To(HaveKey(key))
			}
		})

		Context("with minimal data", func() {
			BeforeEach(func() {
				minimalScrapedData := &scrapper.ScrapedData{
					URL:   "https://minimal.com",
					Title: "",
					Text:  "",
				}
				minimalTextAnalysis := &text_analysis.TextAnalysisResult{}
				testWebpage = webpage.NewWebpage(minimalScrapedData, minimalTextAnalysis)
			})

			It("should still return a valid document", func() {
				document := testWebpage.ToDocument()

				Expect(document).NotTo(BeNil())
				Expect(document["url"]).To(Equal("https://minimal.com"))
				Expect(document["title"]).To(Equal(""))
				Expect(document["text"]).To(Equal(""))
			})
		})
	})
})

var _ = Describe("Link", func() {
	var (
		testLink *webpage.PageLink
	)

	BeforeEach(func() {
		testLink = &webpage.PageLink{
			Link: &scrapper.Link{
				URL:  "https://example.com/link",
				Text: "Example Link",
			},
		}
	})

	Describe("GetID", func() {
		It("should generate consistent URL-based ID", func() {
			id := testLink.GetID()
			Expect(id).To(Equal("example.com/link"))
		})
	})

	Describe("ToNode", func() {
		It("should create a node with correct properties", func() {
			node := testLink.ToNode()

			Expect(node.Type).To(Equal("URL"))
			Expect(node.SubType).To(Equal("webpage"))
			Expect(node.DisplayName).To(Equal("Example Link"))
			Expect(node.ID).To(Equal("example.com/link"))
		})

		Context("when text is empty", func() {
			BeforeEach(func() {
				testLink.Text = ""
			})

			It("should use URL as display name", func() {
				node := testLink.ToNode()
				Expect(node.DisplayName).To(Equal("https://example.com/link"))
			})
		})
	})
})
