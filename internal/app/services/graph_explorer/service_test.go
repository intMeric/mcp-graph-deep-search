package graph_explorer_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/graph_explorer"
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

type mockGraph struct {
	shouldFail      bool
	stats           *graph.GraphStats
	paginatedResult *graph.PaginatedNodesResult
	relations       []*graph.Relation
	connections     []*graph.NodeConnection
}

func (m *mockGraph) CreateNode(ctx context.Context, node *node.Node) error {
	return nil
}

func (m *mockGraph) CreateRelation(ctx context.Context, relation *graph.Relation) error {
	return nil
}

func (m *mockGraph) CreateLink(ctx context.Context, sourceNode *node.Node, targetNode *node.Node, relationType string) (*graph.Relation, error) {
	return nil, nil
}

func (m *mockGraph) GetNode(ctx context.Context, id string) (*node.Node, error) {
	return nil, nil
}

func (m *mockGraph) NodeExists(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (m *mockGraph) Close(ctx context.Context) error {
	return nil
}

func (m *mockGraph) GetGraphStats(ctx context.Context) (*graph.GraphStats, error) {
	if m.shouldFail {
		return nil, errors.New("graph stats failed")
	}
	return m.stats, nil
}

func (m *mockGraph) GetNodesByType(ctx context.Context, nodeType string, offset, limit int) (*graph.PaginatedNodesResult, error) {
	if m.shouldFail {
		return nil, errors.New("get nodes by type failed")
	}
	return m.paginatedResult, nil
}

func (m *mockGraph) GetNodeRelations(ctx context.Context, nodeID string) ([]*graph.Relation, error) {
	if m.shouldFail {
		return nil, errors.New("get node relations failed")
	}
	return m.relations, nil
}

func (m *mockGraph) GetConnectedNodes(ctx context.Context, nodeID string) ([]*graph.NodeConnection, error) {
	if m.shouldFail {
		return nil, errors.New("get connected nodes failed")
	}
	return m.connections, nil
}

var _ = Describe("GraphExplorerService", func() {
	var (
		service graph_explorer.GraphExplorerService
		mockG   *mockGraph
		ctx     context.Context
	)

	BeforeEach(func() {
		mockG = &mockGraph{
			stats: &graph.GraphStats{
				TotalNodes:      10,
				NodesByType:     map[string]int{"URL": 5, "User": 5},
				TotalRelations:  8,
				RelationsByType: map[string]int{"navigation_link": 8},
			},
			paginatedResult: &graph.PaginatedNodesResult{
				Nodes: []*node.Node{
					{Type: "URL", ID: "url1", DisplayName: "Example URL", Location: "db"},
				},
				Total:  1,
				Offset: 0,
				Limit:  10,
			},
			relations: []*graph.Relation{
				{Type: "navigation_link", SourceID: "url1", TargetID: "url2"},
			},
			connections: []*graph.NodeConnection{
				{
					Node:     &node.Node{Type: "URL", ID: "url2", DisplayName: "Connected URL", Location: "db"},
					Relation: &graph.Relation{Type: "navigation_link", SourceID: "url1", TargetID: "url2"},
					IsSource: true,
				},
			},
		}
		service = graph_explorer.NewGraphExplorerService(mockG)
		ctx = context.Background()
	})

	Describe("GetGraphOverview", func() {
		Context("with successful stats retrieval", func() {
			It("should return graph overview", func() {
				overview, err := service.GetGraphOverview(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(overview).NotTo(BeNil())
				Expect(overview.TotalNodes).To(Equal(10))
				Expect(overview.TotalRelations).To(Equal(8))
				Expect(overview.AvailableTypes).To(ContainElements("URL", "User"))
				Expect(overview.Stats).To(Equal(mockG.stats))
			})
		})

		Context("with stats retrieval failure", func() {
			It("should return error", func() {
				mockG.shouldFail = true

				overview, err := service.GetGraphOverview(ctx)

				Expect(err).To(HaveOccurred())
				Expect(overview).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to get graph stats"))
			})
		})
	})

	Describe("GetNodesByType", func() {
		Context("with valid parameters", func() {
			It("should return paginated nodes result", func() {
				result, err := service.GetNodesByType(ctx, "URL", 0, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(mockG.paginatedResult))
			})
		})

		Context("with empty nodeType", func() {
			It("should return error", func() {
				result, err := service.GetNodesByType(ctx, "", 0, 10)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("nodeType cannot be empty"))
			})
		})

		Context("with negative offset", func() {
			It("should normalize offset to 0", func() {
				result, err := service.GetNodesByType(ctx, "URL", -5, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(mockG.paginatedResult))
			})
		})

		Context("with invalid limit", func() {
			It("should normalize limit to 10", func() {
				result, err := service.GetNodesByType(ctx, "URL", 0, -1)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(mockG.paginatedResult))
			})

			It("should cap limit at 10 when exceeding 100", func() {
				result, err := service.GetNodesByType(ctx, "URL", 0, 150)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(mockG.paginatedResult))
			})
		})

		Context("with graph failure", func() {
			It("should return error", func() {
				mockG.shouldFail = true

				result, err := service.GetNodesByType(ctx, "URL", 0, 10)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to get nodes by type"))
			})
		})
	})

	Describe("GetNodeRelations", func() {
		Context("with valid nodeID", func() {
			It("should return node relations", func() {
				relations, err := service.GetNodeRelations(ctx, "url1")

				Expect(err).NotTo(HaveOccurred())
				Expect(relations).To(Equal(mockG.relations))
			})
		})

		Context("with empty nodeID", func() {
			It("should return error", func() {
				relations, err := service.GetNodeRelations(ctx, "")

				Expect(err).To(HaveOccurred())
				Expect(relations).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("nodeID cannot be empty"))
			})
		})

		Context("with graph failure", func() {
			It("should return error", func() {
				mockG.shouldFail = true

				relations, err := service.GetNodeRelations(ctx, "url1")

				Expect(err).To(HaveOccurred())
				Expect(relations).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to get node relations"))
			})
		})
	})

	Describe("GetConnectedNodes", func() {
		Context("with valid nodeID", func() {
			It("should return connected nodes", func() {
				connections, err := service.GetConnectedNodes(ctx, "url1")

				Expect(err).NotTo(HaveOccurred())
				Expect(connections).To(Equal(mockG.connections))
			})
		})

		Context("with empty nodeID", func() {
			It("should return error", func() {
				connections, err := service.GetConnectedNodes(ctx, "")

				Expect(err).To(HaveOccurred())
				Expect(connections).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("nodeID cannot be empty"))
			})
		})

		Context("with graph failure", func() {
			It("should return error", func() {
				mockG.shouldFail = true

				connections, err := service.GetConnectedNodes(ctx, "url1")

				Expect(err).To(HaveOccurred())
				Expect(connections).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to get connected nodes"))
			})
		})
	})

	Describe("Close", func() {
		It("should close without error", func() {
			err := service.Close()

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
