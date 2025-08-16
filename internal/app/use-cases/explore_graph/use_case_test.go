package explore_graph_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/graph_explorer"
	"mgds/internal/app/use-cases/explore_graph"
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

type mockGraphExplorerService struct {
	shouldFail  bool
	overview    *graph_explorer.GraphOverview
	paginated   *graph.PaginatedNodesResult
	relations   []*graph.Relation
	connections []*graph.NodeConnection
}

func (m *mockGraphExplorerService) GetGraphOverview(ctx context.Context) (*graph_explorer.GraphOverview, error) {
	if m.shouldFail {
		return nil, errors.New("service failed")
	}
	return m.overview, nil
}

func (m *mockGraphExplorerService) GetNodesByType(ctx context.Context, nodeType string, offset, limit int) (*graph.PaginatedNodesResult, error) {
	if m.shouldFail {
		return nil, errors.New("service failed")
	}
	return m.paginated, nil
}

func (m *mockGraphExplorerService) GetNodeRelations(ctx context.Context, nodeID string) ([]*graph.Relation, error) {
	if m.shouldFail {
		return nil, errors.New("service failed")
	}
	return m.relations, nil
}

func (m *mockGraphExplorerService) GetConnectedNodes(ctx context.Context, nodeID string) ([]*graph.NodeConnection, error) {
	if m.shouldFail {
		return nil, errors.New("service failed")
	}
	return m.connections, nil
}

func (m *mockGraphExplorerService) Close() error {
	return nil
}

var _ = Describe("ExploreGraphUseCase", func() {
	var (
		useCase     explore_graph.ExploreGraphUseCase
		mockService *mockGraphExplorerService
		ctx         context.Context
	)

	BeforeEach(func() {
		mockService = &mockGraphExplorerService{
			overview: &graph_explorer.GraphOverview{
				Stats: &graph.GraphStats{
					TotalNodes:      10,
					NodesByType:     map[string]int{"URL": 5, "User": 5},
					TotalRelations:  8,
					RelationsByType: map[string]int{"navigation_link": 8},
				},
				AvailableTypes: []string{"URL", "User"},
				TotalNodes:     10,
				TotalRelations: 8,
			},
			paginated: &graph.PaginatedNodesResult{
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
		useCase = explore_graph.NewExploreGraphUseCase(mockService)
		ctx = context.Background()
	})

	Describe("GetGraphOverview", func() {
		Context("with successful service call", func() {
			It("should return graph overview response", func() {
				response, err := useCase.GetGraphOverview(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Overview).To(Equal(mockService.overview))
			})
		})

		Context("with service failure", func() {
			It("should return error", func() {
				mockService.shouldFail = true

				response, err := useCase.GetGraphOverview(ctx)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to get graph overview"))
			})
		})
	})

	Describe("GetNodesByType", func() {
		Context("with valid request", func() {
			It("should return nodes by type response", func() {
				request := &explore_graph.GetNodesByTypeRequest{
					NodeType: "URL",
					Offset:   0,
					Limit:    10,
				}

				response, err := useCase.GetNodesByType(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Result).To(Equal(mockService.paginated))
			})
		})

		Context("with nil request", func() {
			It("should return validation error", func() {
				response, err := useCase.GetNodesByType(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("request cannot be nil"))
			})
		})

		Context("with empty nodeType", func() {
			It("should return validation error", func() {
				request := &explore_graph.GetNodesByTypeRequest{
					NodeType: "",
					Offset:   0,
					Limit:    10,
				}

				response, err := useCase.GetNodesByType(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("nodeType is required"))
			})
		})

		Context("with negative offset", func() {
			It("should return validation error", func() {
				request := &explore_graph.GetNodesByTypeRequest{
					NodeType: "URL",
					Offset:   -1,
					Limit:    10,
				}

				response, err := useCase.GetNodesByType(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("offset must be >= 0"))
			})
		})

		Context("with invalid limit", func() {
			It("should return validation error for zero limit", func() {
				request := &explore_graph.GetNodesByTypeRequest{
					NodeType: "URL",
					Offset:   0,
					Limit:    0,
				}

				response, err := useCase.GetNodesByType(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("limit must be > 0"))
			})

			It("should return validation error for excessive limit", func() {
				request := &explore_graph.GetNodesByTypeRequest{
					NodeType: "URL",
					Offset:   0,
					Limit:    150,
				}

				response, err := useCase.GetNodesByType(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("limit cannot exceed 100"))
			})
		})

		Context("with service failure", func() {
			It("should return error", func() {
				mockService.shouldFail = true
				request := &explore_graph.GetNodesByTypeRequest{
					NodeType: "URL",
					Offset:   0,
					Limit:    10,
				}

				response, err := useCase.GetNodesByType(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to get nodes by type"))
			})
		})
	})

	Describe("GetNodeRelations", func() {
		Context("with valid request", func() {
			It("should return node relations response", func() {
				request := &explore_graph.GetNodeRelationsRequest{
					NodeID: "url1",
				}

				response, err := useCase.GetNodeRelations(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.NodeID).To(Equal("url1"))
				Expect(response.Relations).To(Equal(mockService.relations))
			})
		})

		Context("with nil request", func() {
			It("should return validation error", func() {
				response, err := useCase.GetNodeRelations(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("request cannot be nil"))
			})
		})

		Context("with empty nodeID", func() {
			It("should return validation error", func() {
				request := &explore_graph.GetNodeRelationsRequest{
					NodeID: "",
				}

				response, err := useCase.GetNodeRelations(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("nodeId is required"))
			})
		})

		Context("with service failure", func() {
			It("should return error", func() {
				mockService.shouldFail = true
				request := &explore_graph.GetNodeRelationsRequest{
					NodeID: "url1",
				}

				response, err := useCase.GetNodeRelations(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to get node relations"))
			})
		})
	})

	Describe("GetConnectedNodes", func() {
		Context("with valid request", func() {
			It("should return connected nodes response", func() {
				request := &explore_graph.GetConnectedNodesRequest{
					NodeID: "url1",
				}

				response, err := useCase.GetConnectedNodes(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.NodeID).To(Equal("url1"))
				Expect(response.Connections).To(Equal(mockService.connections))
			})
		})

		Context("with nil request", func() {
			It("should return validation error", func() {
				response, err := useCase.GetConnectedNodes(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("request cannot be nil"))
			})
		})

		Context("with empty nodeID", func() {
			It("should return validation error", func() {
				request := &explore_graph.GetConnectedNodesRequest{
					NodeID: "",
				}

				response, err := useCase.GetConnectedNodes(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("nodeId is required"))
			})
		})

		Context("with service failure", func() {
			It("should return error", func() {
				mockService.shouldFail = true
				request := &explore_graph.GetConnectedNodesRequest{
					NodeID: "url1",
				}

				response, err := useCase.GetConnectedNodes(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to get connected nodes"))
			})
		})
	})
})
