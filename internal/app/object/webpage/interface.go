package webpage

import (
	"mgds/internal/app/services/text_analysis"
	"mgds/internal/pkg/node"
	"mgds/internal/pkg/scrapper"
)

type WebpageInterface interface {
	ToNode() *node.Node
	ToDocument() map[string]any
	GetID() string
	GetLinkNodes() []*node.Node
	GetURL() string
	GetTitle() string
	GetText() string
	GetTextAnalysis() *text_analysis.TextAnalysisResult
	GetScrapedData() *scrapper.ScrapedData
	HasPII() bool
	HasKeywords() bool
}
