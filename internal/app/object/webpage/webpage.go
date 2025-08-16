package webpage

import (
	"crypto/sha256"
	"fmt"
	"net/url"

	"mgds/internal/app/services/text_analysis"
	"mgds/internal/pkg/node"
	"mgds/internal/pkg/scrapper"
)

type Webpage struct {
	URL          string                            `json:"url"`
	ScrapedData  *scrapper.ScrapedData             `json:"scraped_data"`
	TextAnalysis *text_analysis.TextAnalysisResult `json:"text_analysis"`
	Links        []*PageLink                       `json:"links"`
}

type PageLink struct {
	*scrapper.Link
}

func NewWebpage(scrapedData *scrapper.ScrapedData, textAnalysis *text_analysis.TextAnalysisResult) *Webpage {
	webpage := &Webpage{
		URL:          scrapedData.URL,
		ScrapedData:  scrapedData,
		TextAnalysis: textAnalysis,
		Links:        make([]*PageLink, len(scrapedData.Links)),
	}

	for i, link := range scrapedData.Links {
		webpage.Links[i] = &PageLink{Link: &link}
	}

	return webpage
}

func Build(scrapedData *scrapper.ScrapedData, textAnalysis *text_analysis.TextAnalysisResult, originalURL string) WebpageInterface {
	// Use original URL if provided, otherwise use scraped URL
	url := originalURL
	if url == "" {
		url = scrapedData.URL
	}

	webpage := &Webpage{
		URL:          url,
		ScrapedData:  scrapedData,
		TextAnalysis: textAnalysis,
		Links:        make([]*PageLink, len(scrapedData.Links)),
	}

	for i, link := range scrapedData.Links {
		webpage.Links[i] = &PageLink{Link: &link}
	}

	return webpage
}

func (w *Webpage) ToNode() *node.Node {
	nodeID := w.GetID()

	displayName := w.ScrapedData.Title
	if displayName == "" {
		displayName = w.URL
	}

	return &node.Node{
		Type:        "URL",
		SubType:     "webpage",
		DisplayName: displayName,
		ID:          nodeID,
	}
}

func (w *Webpage) GetID() string {
	return GenerateURLNodeID(w.URL)
}

func (w *Webpage) GetLinkNodes() []*node.Node {
	linkNodes := make([]*node.Node, len(w.Links))
	for i, link := range w.Links {
		linkNodes[i] = link.ToNode()
	}
	return linkNodes
}

func (w *Webpage) GetURL() string {
	return w.URL
}

func (w *Webpage) GetTitle() string {
	return w.ScrapedData.Title
}

func (w *Webpage) GetText() string {
	return w.ScrapedData.Text
}

func (w *Webpage) GetTextAnalysis() *text_analysis.TextAnalysisResult {
	return w.TextAnalysis
}

func (w *Webpage) GetScrapedData() *scrapper.ScrapedData {
	return w.ScrapedData
}

func (w *Webpage) HasPII() bool {
	return w.TextAnalysis != nil && w.TextAnalysis.HasPII()
}

func (w *Webpage) HasKeywords() bool {
	return w.TextAnalysis != nil && w.TextAnalysis.HasKeywords()
}

func (w *Webpage) ToDocument() map[string]any {
	return map[string]any{
		"url":        w.URL,
		"title":      w.ScrapedData.Title,
		"text":       w.ScrapedData.Text,
		"meta_tags":  w.ScrapedData.MetaTags,
		"pii":        w.TextAnalysis.PIIResult,
		"keywords":   w.TextAnalysis.Keywords,
		"links":      w.ScrapedData.Links,
		"images":     w.ScrapedData.Images,
		"scraped_at": w.ScrapedData.ScrapedAt,
	}
}

func (l *PageLink) ToNode() *node.Node {
	nodeID := l.GetID()

	displayName := l.Text
	if displayName == "" {
		displayName = l.URL
	}

	return &node.Node{
		Type:        "URL",
		SubType:     "webpage",
		DisplayName: displayName,
		ID:          nodeID,
	}
}

func (l *PageLink) GetID() string {
	return GenerateURLNodeID(l.URL)
}

func GenerateURLNodeID(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil || parsedURL.Host == "" {
		hash := sha256.Sum256([]byte(urlStr))
		return fmt.Sprintf("url-%x", hash[:8])
	}

	nodeID := parsedURL.Host + parsedURL.Path
	if parsedURL.RawQuery != "" {
		nodeID += "?" + parsedURL.RawQuery
	}

	return nodeID
}
