package search_engine

type SearchResponse struct {
	Query   string   `json:"query"`
	Results []Result `json:"results"`
}

type Result struct {
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	PublishedDate string   `json:"publishedDate,omitempty"`
	ImgSrc        string   `json:"img_src,omitempty"`
	Engine        string   `json:"engine"`
	Score         float64  `json:"score"`
	Category      string   `json:"category"`
	Query         string   `json:"query"`
	RequestDate   string   `json:"requestDate"`
	Tags          []string `json:"tags"`
}

type SearchEngineType string

const (
	SearXNGType SearchEngineType = "searxng"
)

type SearXNGConfig struct {
	Host     string
	Language string
}