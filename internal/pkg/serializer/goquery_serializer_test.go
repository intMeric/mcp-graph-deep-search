package serializer_test

import (
	"context"
	"mgds/internal/pkg/serializer"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const sampleHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
    <meta name="description" content="A test page for HTML serialization">
    <meta property="og:title" content="Test Page OG">
</head>
<body>
    <h1>Welcome to Test Page</h1>
    <p>This is a paragraph with <a href="https://example.com" title="Example Link" rel="nofollow">a link</a>.</p>
    <p>Another paragraph with <a href="https://google.com">Google</a>.</p>
    <img src="https://example.com/image.jpg" alt="Example Image" width="200" height="150" title="Sample Image">
    <img src="https://example.com/logo.png" alt="Logo">
</body>
</html>
`

var _ = Describe("GoquerySerializer", func() {
	var (
		htmlSerializer serializer.HTMLSerializer
		ctx            context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		htmlSerializer = serializer.NewGoquerySerializer(serializer.DefaultDocumentOptions())
	})

	AfterEach(func() {
		if htmlSerializer != nil {
			htmlSerializer.Close()
		}
	})

	Describe("ParseHTML", func() {
		Context("with valid HTML", func() {
			It("should parse HTML successfully", func() {
				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc).NotTo(BeNil())
				Expect(doc.Title).To(Equal("Test Page"))
			})

			It("should extract title correctly", func() {
				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Title).To(Equal("Test Page"))
			})

			It("should extract text content", func() {
				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Text).To(ContainSubstring("Welcome to Test Page"))
				Expect(doc.Text).To(ContainSubstring("This is a paragraph"))
				Expect(doc.Text).NotTo(ContainSubstring("<h1>"))
			})

			It("should extract meta tags", func() {
				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.MetaTags).To(HaveKey("description"))
				Expect(doc.MetaTags["description"]).To(Equal("A test page for HTML serialization"))
				Expect(doc.MetaTags).To(HaveKey("og:title"))
				Expect(doc.MetaTags["og:title"]).To(Equal("Test Page OG"))
			})

			It("should extract links", func() {
				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Links).To(HaveLen(2))

				firstLink := doc.Links[0]
				Expect(firstLink.URL).To(Equal("https://example.com"))
				Expect(firstLink.Text).To(Equal("a link"))
				Expect(firstLink.Title).To(Equal("Example Link"))
				Expect(firstLink.Rel).To(Equal("nofollow"))

				secondLink := doc.Links[1]
				Expect(secondLink.URL).To(Equal("https://google.com"))
				Expect(secondLink.Text).To(Equal("Google"))
			})

			It("should extract images", func() {
				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Images).To(HaveLen(2))

				firstImage := doc.Images[0]
				Expect(firstImage.URL).To(Equal("https://example.com/image.jpg"))
				Expect(firstImage.Alt).To(Equal("Example Image"))
				Expect(firstImage.Width).To(Equal("200"))
				Expect(firstImage.Height).To(Equal("150"))
				Expect(firstImage.Title).To(Equal("Sample Image"))

				secondImage := doc.Images[1]
				Expect(secondImage.URL).To(Equal("https://example.com/logo.png"))
				Expect(secondImage.Alt).To(Equal("Logo"))
			})

			It("should extract head and body content", func() {
				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Head).To(ContainSubstring("<title>Test Page</title>"))
				Expect(doc.Head).To(ContainSubstring("meta"))
				Expect(doc.Body).To(ContainSubstring("<h1>Welcome to Test Page</h1>"))
			})
		})

		Context("with custom options", func() {
			It("should respect ExtractLinks option", func() {
				options := serializer.DefaultDocumentOptions()
				options.ExtractLinks = false
				htmlSerializer = serializer.NewGoquerySerializer(options)

				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Links).To(BeEmpty())
			})

			It("should respect ExtractImages option", func() {
				options := serializer.DefaultDocumentOptions()
				options.ExtractImages = false
				htmlSerializer = serializer.NewGoquerySerializer(options)

				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Images).To(BeEmpty())
			})

			It("should respect ExtractMetaTags option", func() {
				options := serializer.DefaultDocumentOptions()
				options.ExtractMetaTags = false
				htmlSerializer = serializer.NewGoquerySerializer(options)

				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.MetaTags).To(BeEmpty())
			})

			It("should respect ExtractText option", func() {
				options := serializer.DefaultDocumentOptions()
				options.ExtractText = false
				htmlSerializer = serializer.NewGoquerySerializer(options)

				doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Text).To(BeEmpty())
			})
		})

		Context("with invalid HTML", func() {
			It("should handle malformed HTML gracefully", func() {
				malformedHTML := "<html><head><title>Test</title></head><body><p>Unclosed paragraph</body></html>"

				doc, err := htmlSerializer.ParseHTML(ctx, malformedHTML)

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Title).To(Equal("Test"))
			})

			It("should handle empty HTML", func() {
				doc, err := htmlSerializer.ParseHTML(ctx, "")

				Expect(err).NotTo(HaveOccurred())
				Expect(doc.Title).To(BeEmpty())
				Expect(doc.Text).To(BeEmpty())
			})
		})
	})

	Describe("ParseHTMLFromReader", func() {
		It("should parse HTML from reader successfully", func() {
			reader := strings.NewReader(sampleHTML)
			doc, err := htmlSerializer.ParseHTMLFromReader(ctx, reader)

			Expect(err).NotTo(HaveOccurred())
			Expect(doc.Title).To(Equal("Test Page"))
		})
	})

	Describe("SerializeToString", func() {
		It("should serialize document back to HTML", func() {
			doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)
			Expect(err).NotTo(HaveOccurred())

			html, err := htmlSerializer.SerializeToString(ctx, doc)

			Expect(err).NotTo(HaveOccurred())
			Expect(html).To(ContainSubstring("<title>Test Page</title>"))
			Expect(html).To(ContainSubstring("Welcome to Test Page"))
		})

		It("should reconstruct HTML without raw document", func() {
			doc := &serializer.Document{
				Title: "Reconstructed Page",
				Head:  "<meta name=\"test\" content=\"value\">",
				Body:  "<h1>Hello World</h1>",
			}

			htmlStr, err := htmlSerializer.SerializeToString(ctx, doc)

			Expect(err).NotTo(HaveOccurred())
			Expect(htmlStr).To(ContainSubstring("<title>Reconstructed Page</title>"))
			Expect(htmlStr).To(ContainSubstring("<h1>Hello World</h1>"))
			Expect(htmlStr).To(ContainSubstring("<!DOCTYPE html>"))
		})
	})

	Describe("SerializeToWriter", func() {
		It("should serialize document to writer", func() {
			doc, err := htmlSerializer.ParseHTML(ctx, sampleHTML)
			Expect(err).NotTo(HaveOccurred())

			var buffer strings.Builder
			err = htmlSerializer.SerializeToWriter(ctx, doc, &buffer)

			Expect(err).NotTo(HaveOccurred())
			result := buffer.String()
			Expect(result).To(ContainSubstring("<title>Test Page</title>"))
		})
	})

	Describe("Factory", func() {
		It("should create goquery serializer from factory", func() {
			serializer := serializer.NewHTMLSerializer(serializer.GoquerySerializer, nil)

			Expect(serializer).NotTo(BeNil())

			doc, err := serializer.ParseHTML(ctx, sampleHTML)
			Expect(err).NotTo(HaveOccurred())
			Expect(doc.Title).To(Equal("Test Page"))
		})

		It("should use default serializer for unknown type", func() {
			serializer := serializer.NewHTMLSerializer("unknown", nil)

			Expect(serializer).NotTo(BeNil())

			doc, err := serializer.ParseHTML(ctx, sampleHTML)
			Expect(err).NotTo(HaveOccurred())
			Expect(doc.Title).To(Equal("Test Page"))
		})
	})
})
