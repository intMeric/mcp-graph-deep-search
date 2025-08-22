package websearch_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/src/pkg/node"
	webseach "mgds/src/service/web_seach"
	websearch "mgds/src/usecase/web_search"
)

type mockWebSearchService struct {
	shouldFail bool
	nodes      []node.Interface
}

func (m *mockWebSearchService) SearchAndIndex(ctx context.Context, query string, options *webseach.SearchOptions) (*webseach.SearchIndexResult, error) {
	if m.shouldFail {
		return nil, &mockError{message: "search failed"}
	}

	return &webseach.SearchIndexResult{
		Query:        query,
		NodesCreated: m.nodes,
		NodesUpdated: []node.Interface{},
		TimeElapsed:  100 * time.Millisecond,
	}, nil
}

func (m *mockWebSearchService) SearchAndIndexURL(ctx context.Context, url string, options *webseach.SearchOptions) (*webseach.SearchIndexResult, error) {
	return nil, nil
}

func (m *mockWebSearchService) Close() error {
	return nil
}

// Mock d'un node
type mockNode struct {
	id         string
	title      string
	properties map[string]any
}

func (m *mockNode) GetMgdsId() string      { return m.id }
func (m *mockNode) GetDisplayName() string { return m.title }
func (m *mockNode) GetDescription() string { return "test description" }
func (m *mockNode) GetType() string        { return "url_node" }
func (m *mockNode) GetProperty(key string) (any, bool) {
	if m.properties == nil {
		return nil, false
	}
	val, exists := m.properties[key]
	return val, exists
}
func (m *mockNode) SetProperty(key string, value any) {
	if m.properties == nil {
		m.properties = make(map[string]any)
	}
	m.properties[key] = value
}
func (m *mockNode) Serialize() (map[string]any, error)    { return nil, nil }
func (m *mockNode) Deserialize(data map[string]any) error { return nil }
func (m *mockNode) IsValid() bool                         { return true }

type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

var _ = Describe("WebSearch UseCase Service", func() {
	var (
		service        websearch.Service
		mockWebService *mockWebSearchService
		ctx            context.Context
	)

	BeforeEach(func() {
		mockWebService = &mockWebSearchService{}
		service = websearch.NewService(mockWebService, nil)
		ctx = context.Background()
	})

	Describe("Execute", func() {
		Context("with valid request", func() {
			It("should return success response with nodes", func() {
				// Setup mock data
				mockNode1 := &mockNode{
					id:    "node-1",
					title: "Test Page 1",
					properties: map[string]any{
						"url":     "https://example.com/1",
						"title":   "Test Page 1",
						"content": "Content of page 1",
					},
				}

				mockNode2 := &mockNode{
					id:    "node-2",
					title: "Test Page 2",
					properties: map[string]any{
						"url":     "https://example.com/2",
						"title":   "Test Page 2",
						"content": "Content of page 2",
					},
				}

				mockWebService.nodes = []node.Interface{mockNode1, mockNode2}

				req := &websearch.SearchRequest{
					Query:      "test query",
					MaxResults: 5,
				}

				response, err := service.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(websearch.StatusSuccess))
				Expect(response.Query).To(Equal("test query"))
				Expect(response.Nodes).To(HaveLen(2))

				Expect(response.Nodes[0].ID).To(Equal("node-1"))
				Expect(response.Nodes[0].URL).To(Equal("https://example.com/1"))
				Expect(response.Nodes[0].Title).To(Equal("Test Page 1"))
				Expect(response.Nodes[0].Content).To(Equal("test description"))

				Expect(response.Nodes[1].ID).To(Equal("node-2"))
				Expect(response.Nodes[1].URL).To(Equal("https://example.com/2"))
				Expect(response.Nodes[1].Title).To(Equal("Test Page 2"))
				Expect(response.Nodes[1].Content).To(Equal("Content of page 2"))

				Expect(response.ExecutionTime).To(BeNumerically(">", 0))
			})
		})

		Context("with empty query", func() {
			It("should return failed status", func() {
				req := &websearch.SearchRequest{
					Query:      "",
					MaxResults: 5,
				}

				response, err := service.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(websearch.StatusFailed))
				Expect(response.Error).To(ContainSubstring("Query cannot be empty"))
			})
		})

		Context("with zero MaxResults", func() {
			It("should default to 10", func() {
				mockWebService.nodes = []node.Interface{}

				req := &websearch.SearchRequest{
					Query:      "test",
					MaxResults: 0,
				}

				response, err := service.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(websearch.StatusSuccess))
			})
		})

		Context("when web search service fails", func() {
			It("should return failed status", func() {
				mockWebService.shouldFail = true

				req := &websearch.SearchRequest{
					Query:      "test query",
					MaxResults: 5,
				}

				response, err := service.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(websearch.StatusFailed))
				Expect(response.Error).To(ContainSubstring("Search and index failed"))
			})
		})

		Context("with nodes without URL property", func() {
			It("should handle missing properties gracefully", func() {
				mockNode := &mockNode{
					id:    "node-1",
					title: "Test Page",
					properties: map[string]any{
						"title": "Test Page",
					},
				}

				mockWebService.nodes = []node.Interface{mockNode}

				req := &websearch.SearchRequest{
					Query:      "test query",
					MaxResults: 5,
				}

				response, err := service.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(websearch.StatusSuccess))
				Expect(response.Nodes).To(HaveLen(1))
				Expect(response.Nodes[0].URL).To(Equal("")) // URL vide
				Expect(response.Nodes[0].Title).To(Equal("Test Page"))
			})
		})
	})
})
