package serializer

import (
	"context"
	"io"
)

// HTMLSerializer provides methods to serialize HTML content
type HTMLSerializer interface {
	// ParseHTML parses HTML content from a string and returns a Document
	ParseHTML(ctx context.Context, htmlContent string) (*Document, error)

	// ParseHTMLFromReader parses HTML content from a reader and returns a Document
	ParseHTMLFromReader(ctx context.Context, reader io.Reader) (*Document, error)

	// SerializeToString converts a Document back to HTML string
	SerializeToString(ctx context.Context, doc *Document) (string, error)

	// SerializeToWriter writes a Document as HTML to a writer
	SerializeToWriter(ctx context.Context, doc *Document, writer io.Writer) error

	// Close releases any resources held by the serializer
	Close() error
}

// Document represents a parsed HTML document with manipulation capabilities
type Document struct {
	// Title of the HTML document
	Title string

	// Body contains the body content as HTML string
	Body string

	// Head contains the head content as HTML string
	Head string

	// Links contains all links found in the document
	Links []Link

	// Images contains all images found in the document
	Images []Image

	// MetaTags contains meta tag key-value pairs
	MetaTags map[string]string

	// Text contains the plain text content (no HTML tags)
	Text string

	// rawDocument holds the internal goquery document (not exported for clean interface)
	rawDocument any
}

// Link represents an HTML link element
type Link struct {
	URL    string `json:"url"`
	Text   string `json:"text"`
	Title  string `json:"title,omitempty"`
	Rel    string `json:"rel,omitempty"`
	Target string `json:"target,omitempty"`
}

// Image represents an HTML image element
type Image struct {
	URL    string `json:"url"`
	Alt    string `json:"alt,omitempty"`
	Title  string `json:"title,omitempty"`
	Width  string `json:"width,omitempty"`
	Height string `json:"height,omitempty"`
}

// DocumentOptions provides options for HTML parsing and serialization
type DocumentOptions struct {
	// ExtractLinks determines if links should be extracted during parsing
	ExtractLinks bool

	// ExtractImages determines if images should be extracted during parsing
	ExtractImages bool

	// ExtractMetaTags determines if meta tags should be extracted during parsing
	ExtractMetaTags bool

	// ExtractText determines if plain text should be extracted during parsing
	ExtractText bool

	// PreserveWhitespace determines if whitespace should be preserved in text extraction
	PreserveWhitespace bool
}

// DefaultDocumentOptions returns default options for document parsing
func DefaultDocumentOptions() *DocumentOptions {
	return &DocumentOptions{
		ExtractLinks:       true,
		ExtractImages:      true,
		ExtractMetaTags:    true,
		ExtractText:        true,
		PreserveWhitespace: false,
	}
}
