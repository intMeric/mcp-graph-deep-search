package webpage

import (
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
	GetScrapedData() *scrapper.ScrapedData
	HasKeywords() bool
}
