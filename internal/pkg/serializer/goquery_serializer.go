package serializer

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type goquerySerializer struct {
	options *DocumentOptions
}

// NewGoquerySerializer creates a new HTML serializer using goquery library
func NewGoquerySerializer(options *DocumentOptions) HTMLSerializer {
	if options == nil {
		options = DefaultDocumentOptions()
	}

	return &goquerySerializer{
		options: options,
	}
}

func (g *goquerySerializer) ParseHTML(ctx context.Context, htmlContent string) (*Document, error) {
	reader := strings.NewReader(htmlContent)
	return g.ParseHTMLFromReader(ctx, reader)
}

func (g *goquerySerializer) ParseHTMLFromReader(ctx context.Context, reader io.Reader) (*Document, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	document := &Document{
		rawDocument: doc,
		MetaTags:    make(map[string]string),
		Links:       make([]Link, 0),
		Images:      make([]Image, 0),
	}

	// Extract title
	document.Title = strings.TrimSpace(doc.Find("title").Text())

	// Extract head content
	headHtml, _ := doc.Find("head").Html()
	document.Head = strings.TrimSpace(headHtml)

	// Extract body content
	bodyHtml, _ := doc.Find("body").Html()
	document.Body = strings.TrimSpace(bodyHtml)

	// Extract text content if requested
	if g.options.ExtractText {
		text := doc.Find("body").Text()
		if !g.options.PreserveWhitespace {
			// Clean up whitespace
			text = strings.Join(strings.Fields(text), " ")
		}
		document.Text = strings.TrimSpace(text)
	}

	// Extract meta tags if requested
	if g.options.ExtractMetaTags {
		doc.Find("meta").Each(func(i int, s *goquery.Selection) {
			if name, exists := s.Attr("name"); exists {
				if content, exists := s.Attr("content"); exists {
					document.MetaTags[name] = content
				}
			}
			if property, exists := s.Attr("property"); exists {
				if content, exists := s.Attr("content"); exists {
					document.MetaTags[property] = content
				}
			}
		})
	}

	// Extract links if requested
	if g.options.ExtractLinks {
		doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			if href != "" {
				link := Link{
					URL:  href,
					Text: strings.TrimSpace(s.Text()),
				}

				if title, exists := s.Attr("title"); exists {
					link.Title = title
				}
				if rel, exists := s.Attr("rel"); exists {
					link.Rel = rel
				}
				if target, exists := s.Attr("target"); exists {
					link.Target = target
				}

				document.Links = append(document.Links, link)
			}
		})
	}

	// Extract images if requested
	if g.options.ExtractImages {
		doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
			src, _ := s.Attr("src")
			if src != "" {
				image := Image{
					URL: src,
				}

				if alt, exists := s.Attr("alt"); exists {
					image.Alt = alt
				}
				if title, exists := s.Attr("title"); exists {
					image.Title = title
				}
				if width, exists := s.Attr("width"); exists {
					image.Width = width
				}
				if height, exists := s.Attr("height"); exists {
					image.Height = height
				}

				document.Images = append(document.Images, image)
			}
		})
	}

	return document, nil
}

func (g *goquerySerializer) SerializeToString(ctx context.Context, doc *Document) (string, error) {
	if doc.rawDocument == nil {
		return g.reconstructHTML(doc), nil
	}

	goqueryDoc, ok := doc.rawDocument.(*goquery.Document)
	if !ok {
		return "", fmt.Errorf("invalid document type")
	}

	html, err := goqueryDoc.Html()
	if err != nil {
		return "", fmt.Errorf("failed to serialize HTML: %w", err)
	}

	return html, nil
}

func (g *goquerySerializer) SerializeToWriter(ctx context.Context, doc *Document, writer io.Writer) error {
	html, err := g.SerializeToString(ctx, doc)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(html))
	if err != nil {
		return fmt.Errorf("failed to write HTML: %w", err)
	}

	return nil
}

func (g *goquerySerializer) reconstructHTML(doc *Document) string {
	var builder strings.Builder

	builder.WriteString("<!DOCTYPE html>\n<html>\n")

	// Head section
	builder.WriteString("<head>\n")
	if doc.Title != "" {
		builder.WriteString(fmt.Sprintf("<title>%s</title>\n", doc.Title))
	}
	if doc.Head != "" {
		builder.WriteString(doc.Head)
		builder.WriteString("\n")
	}
	builder.WriteString("</head>\n")

	// Body section
	builder.WriteString("<body>\n")
	if doc.Body != "" {
		builder.WriteString(doc.Body)
		builder.WriteString("\n")
	}
	builder.WriteString("</body>\n")

	builder.WriteString("</html>")

	return builder.String()
}

func (g *goquerySerializer) Close() error {
	// goquery doesn't need explicit cleanup
	return nil
}
