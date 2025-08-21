package webseach_test

import (
	"context"
	"errors"
	"mgds/src/pkg/graph"
	"mgds/src/pkg/node"
	"mgds/src/pkg/object"
	"mgds/src/pkg/scrapper"
	"mgds/src/pkg/search_engine"
	webseach "mgds/src/service/web_seach"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Mock implementations
type mockSearchEngine struct {
	searchResult *search_engine.SearchResponse
	searchError  error
	closed       bool
}

func (m *mockSearchEngine) Search(ctx context.Context, req *search_engine.SearchRequest) (*search_engine.SearchResponse, error) {
	return m.searchResult, m.searchError
}

func (m *mockSearchEngine) SearchSimple(ctx context.Context, query string) (*search_engine.SearchResponse, error) {
	return m.searchResult, m.searchError
}

func (m *mockSearchEngine) IsHealthy(ctx context.Context) error {
	return nil
}

func (m *mockSearchEngine) SetConfig(config *search_engine.Config) {}

func (m *mockSearchEngine) GetConfig() *search_engine.Config {
	return search_engine.DefaultConfig()
}

func (m *mockSearchEngine) Close() error {
	m.closed = true
	return nil
}

type mockScrapper struct {
	scrapResults map[string]*scrapper.ScrapResult
	scrapError   error
	closed       bool
}

func (m *mockScrapper) Scrape(ctx context.Context, url string) (*scrapper.ScrapResult, error) {
	if m.scrapError != nil {
		return nil, m.scrapError
	}

	result, exists := m.scrapResults[url]
	if !exists {
		return &scrapper.ScrapResult{
			URL:        url,
			Title:      "Default Title",
			Content:    "Default content",
			Links:      []scrapper.Link{},
			ScrapedAt:  time.Now(),
			StatusCode: 200,
		}, nil
	}

	return result, nil
}

func (m *mockScrapper) ScrapeBatch(ctx context.Context, urls []string) ([]*scrapper.ScrapResult, error) {
	var results []*scrapper.ScrapResult
	for _, url := range urls {
		result, err := m.Scrape(ctx, url)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (m *mockScrapper) IsValidURL(url string) bool {
	return true
}

func (m *mockScrapper) NormalizeURL(url string) (string, error) {
	return url, nil
}

func (m *mockScrapper) SetConfig(config *scrapper.Config) {}

func (m *mockScrapper) GetConfig() *scrapper.Config {
	return &scrapper.Config{}
}

func (m *mockScrapper) Close() error {
	m.closed = true
	return nil
}

type mockGraph struct {
	nodes     map[string]node.Interface
	relations map[string][]*graph.Relation
	closed    bool
}

func newMockGraph() *mockGraph {
	return &mockGraph{
		nodes:     make(map[string]node.Interface),
		relations: make(map[string][]*graph.Relation),
	}
}

func (m *mockGraph) AddNode(ctx context.Context, node node.Interface) error {
	m.nodes[node.GetMgdsId()] = node
	return nil
}

func (m *mockGraph) GetNode(ctx context.Context, mgdsId string) (node.Interface, error) {
	node, exists := m.nodes[mgdsId]
	if !exists {
		return nil, errors.New("node not found")
	}
	return node, nil
}

func (m *mockGraph) UpdateNode(ctx context.Context, node node.Interface) error {
	m.nodes[node.GetMgdsId()] = node
	return nil
}

func (m *mockGraph) DeleteNode(ctx context.Context, mgdsId string) error {
	delete(m.nodes, mgdsId)
	return nil
}

func (m *mockGraph) AddRelation(ctx context.Context, relation *graph.Relation) error {
	key := relation.FromNodeId + "-" + relation.Label
	m.relations[key] = append(m.relations[key], relation)
	return nil
}

func (m *mockGraph) GetRelations(ctx context.Context, fromNodeId string) ([]*graph.Relation, error) {
	var allRelations []*graph.Relation
	for key, relations := range m.relations {
		if key[:len(fromNodeId)] == fromNodeId {
			allRelations = append(allRelations, relations...)
		}
	}
	return allRelations, nil
}

func (m *mockGraph) GetIncomingRelations(ctx context.Context, toNodeId string) ([]*graph.Relation, error) {
	var incomingRelations []*graph.Relation
	for _, relations := range m.relations {
		for _, relation := range relations {
			if relation.ToNodeId == toNodeId {
				incomingRelations = append(incomingRelations, relation)
			}
		}
	}
	return incomingRelations, nil
}

func (m *mockGraph) DeleteRelation(ctx context.Context, fromNodeId, toNodeId, label string) error {
	key := fromNodeId + "-" + label
	relations := m.relations[key]
	var filtered []*graph.Relation
	for _, relation := range relations {
		if relation.ToNodeId != toNodeId {
			filtered = append(filtered, relation)
		}
	}
	m.relations[key] = filtered
	return nil
}

func (m *mockGraph) GetConnectedNodes(ctx context.Context, mgdsId string) ([]node.Interface, error) {
	return nil, nil
}

func (m *mockGraph) FindPath(ctx context.Context, fromNodeId, toNodeId string) ([]*graph.Relation, error) {
	return nil, nil
}

func (m *mockGraph) FindNodesByType(ctx context.Context, nodeType string) ([]node.Interface, error) {
	var result []node.Interface
	for _, node := range m.nodes {
		if node.GetType() == nodeType {
			result = append(result, node)
		}
	}
	return result, nil
}

func (m *mockGraph) FindNodesByProperty(ctx context.Context, key string, value any) ([]node.Interface, error) {
	var result []node.Interface
	for _, node := range m.nodes {
		if prop, exists := node.GetProperty(key); exists && prop == value {
			result = append(result, node)
		}
	}
	return result, nil
}

func (m *mockGraph) NodeExists(ctx context.Context, mgdsId string) (bool, error) {
	_, exists := m.nodes[mgdsId]
	return exists, nil
}

func (m *mockGraph) GetAllNodes(ctx context.Context) ([]node.Interface, error) {
	var result []node.Interface
	for _, node := range m.nodes {
		result = append(result, node)
	}
	return result, nil
}

func (m *mockGraph) Connect(ctx context.Context) error {
	return nil
}

func (m *mockGraph) Close() error {
	m.closed = true
	return nil
}

func (m *mockGraph) Ping(ctx context.Context) error {
	return nil
}

var _ = Describe("WebSearchService", func() {
	var (
		webSearchService webseach.Interface
		mockSearchEng    *mockSearchEngine
		mockScrap        *mockScrapper
		mockGr           *mockGraph
		ctx              context.Context
	)

	BeforeEach(func() {
		mockSearchEng = &mockSearchEngine{}
		mockScrap = &mockScrapper{
			scrapResults: make(map[string]*scrapper.ScrapResult),
		}
		mockGr = newMockGraph()

		webSearchService = webseach.NewWebSearchService(mockSearchEng, mockScrap, mockGr)
		ctx = context.Background()
	})

	AfterEach(func() {
		if webSearchService != nil {
			webSearchService.Close()
		}
	})

	Describe("SearchAndIndex", func() {
		Context("with successful search and scraping", func() {
			BeforeEach(func() {
				mockSearchEng.searchResult = &search_engine.SearchResponse{
					Results: []search_engine.SearchResult{
						{
							URL:     "https://example.com/page1",
							Title:   "Page 1",
							Content: "Content 1",
						},
						{
							URL:     "https://example.com/page2",
							Title:   "Page 2",
							Content: "Content 2",
						},
					},
					Query: "test query",
				}

				mockScrap.scrapResults["https://example.com/page1"] = &scrapper.ScrapResult{
					URL:     "https://example.com/page1",
					Title:   "Page 1 Title",
					Content: "Scraped content 1",
					Links: []scrapper.Link{
						{
							URL:        "https://example.com/link1",
							Text:       "Link 1",
							IsInternal: true,
						},
					},
					ScrapedAt:  time.Now(),
					StatusCode: 200,
				}

				mockScrap.scrapResults["https://example.com/page2"] = &scrapper.ScrapResult{
					URL:     "https://example.com/page2",
					Title:   "Page 2 Title",
					Content: "Scraped content 2",
					Links: []scrapper.Link{
						{
							URL:        "https://example.com/link2",
							Text:       "Link 2",
							IsInternal: false,
						},
					},
					ScrapedAt:  time.Now(),
					StatusCode: 200,
				}
			})

			It("should successfully search, scrape, and index pages", func() {
				result, err := webSearchService.SearchAndIndex(ctx, "test query", nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Query).To(Equal("test query"))
				Expect(result.Stats.SearchResults).To(Equal(2))
				Expect(result.Stats.ScrapedPages).To(Equal(2))
				Expect(result.Stats.NodesCreated).To(BeNumerically(">=", 2))
				Expect(result.Stats.RelationsCreated).To(Equal(2))
				Expect(result.TimeElapsed).To(BeNumerically(">", 0))
			})

			It("should create nodes in the graph", func() {
				_, err := webSearchService.SearchAndIndex(ctx, "test query", nil)
				Expect(err).NotTo(HaveOccurred())

				nodes, err := mockGr.GetAllNodes(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(nodes)).To(BeNumerically(">=", 2))

				// Check that URL nodes were created
				urlNodes, err := mockGr.FindNodesByType(ctx, object.URLNodeType)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(urlNodes)).To(BeNumerically(">=", 2))
			})

			It("should create link relations", func() {
				_, err := webSearchService.SearchAndIndex(ctx, "test query", nil)
				Expect(err).NotTo(HaveOccurred())

				// Check that relations were created
				Expect(len(mockGr.relations)).To(BeNumerically(">=", 2))
			})
		})

		Context("with search engine failure", func() {
			BeforeEach(func() {
				mockSearchEng.searchError = errors.New("search engine failed")
			})

			It("should return error when search fails", func() {
				result, err := webSearchService.SearchAndIndex(ctx, "test query", nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("search failed"))
				Expect(result.Stats.SearchResults).To(Equal(0))
			})
		})

		Context("with scraping failure", func() {
			BeforeEach(func() {
				mockSearchEng.searchResult = &search_engine.SearchResponse{
					Results: []search_engine.SearchResult{
						{
							URL:     "https://example.com/page1",
							Title:   "Page 1",
							Content: "Content 1",
						},
					},
					Query: "test query",
				}

				mockScrap.scrapError = errors.New("scraping failed")
			})

			It("should handle scraping errors gracefully", func() {
				result, err := webSearchService.SearchAndIndex(ctx, "test query", nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Stats.SkippedPages).To(Equal(1))
				Expect(len(result.Errors)).To(Equal(1))
				Expect(result.Errors[0].Type).To(Equal("scraping"))
			})
		})

		Context("with skip scraping option", func() {
			BeforeEach(func() {
				mockSearchEng.searchResult = &search_engine.SearchResponse{
					Results: []search_engine.SearchResult{
						{
							URL:     "https://example.com/page1",
							Title:   "Page 1",
							Content: "Content 1",
						},
					},
					Query: "test query",
				}
			})

			It("should only create nodes from search results without scraping", func() {
				options := &webseach.SearchOptions{
					SkipScraping: true,
				}

				result, err := webSearchService.SearchAndIndex(ctx, "test query", options)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Stats.ScrapedPages).To(Equal(0))
				Expect(result.Stats.NodesCreated).To(Equal(1))
				Expect(result.Stats.RelationsCreated).To(Equal(0))
			})
		})
	})

	Describe("SearchAndIndexURL", func() {
		Context("with successful scraping", func() {
			BeforeEach(func() {
				mockScrap.scrapResults["https://example.com/direct"] = &scrapper.ScrapResult{
					URL:     "https://example.com/direct",
					Title:   "Direct Page",
					Content: "Direct content",
					Links: []scrapper.Link{
						{
							URL:        "https://example.com/linked",
							Text:       "Linked Page",
							IsInternal: true,
						},
					},
					ScrapedAt:  time.Now(),
					StatusCode: 200,
				}
			})

			It("should scrape and index a single URL", func() {
				result, err := webSearchService.SearchAndIndexURL(ctx, "https://example.com/direct", nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Query).To(ContainSubstring("Direct URL"))
				Expect(result.Stats.SearchResults).To(Equal(1))
				Expect(result.Stats.ScrapedPages).To(Equal(1))
				Expect(result.Stats.LinksFound).To(Equal(1))
				Expect(result.Stats.NodesCreated).To(BeNumerically(">=", 1))
				Expect(result.Stats.RelationsCreated).To(Equal(1))
			})
		})

		Context("with scraping failure", func() {
			BeforeEach(func() {
				mockScrap.scrapError = errors.New("scraping failed")
			})

			It("should handle scraping errors gracefully", func() {
				result, err := webSearchService.SearchAndIndexURL(ctx, "https://example.com/fail", nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Stats.SkippedPages).To(Equal(1))
				Expect(len(result.Errors)).To(Equal(1))
				Expect(result.Errors[0].Type).To(Equal("scraping"))
			})
		})
	})

	Describe("Close", func() {
		It("should close all dependencies", func() {
			err := webSearchService.Close()

			Expect(err).NotTo(HaveOccurred())
			Expect(mockSearchEng.closed).To(BeTrue())
			Expect(mockScrap.closed).To(BeTrue())
			Expect(mockGr.closed).To(BeTrue())
		})
	})
})
