package prune_graph_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/graph_pruner"
	"mgds/internal/app/use-cases/prune_graph"
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

// Mock service for testing
type mockPrunerService struct {
	shouldError bool
	deletedNode *graph.DeletionResult
	preview     *graph_pruner.DeletionPreview
}

func newMockPrunerService() *mockPrunerService {
	return &mockPrunerService{
		deletedNode: &graph.DeletionResult{
			DeletedNodes:     1,
			DeletedRelations: 2,
			DeletedNodeIDs:   []string{"test-node"},
			Errors:           []string{},
		},
		preview: &graph_pruner.DeletionPreview{
			TargetNode: &node.Node{
				ID:          "test-node",
				Type:        "webpage",
				DisplayName: "Test Node",
				Location:    "webpage",
			},
			AffectedNodes:  []*node.Node{},
			TotalNodes:     1,
			TotalRelations: 2,
			HasDocuments:   true,
		},
	}
}

func (m *mockPrunerService) DeleteNode(ctx context.Context, nodeID string) (*graph.DeletionResult, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.deletedNode, nil
}

func (m *mockPrunerService) DeleteNodeCascade(ctx context.Context, nodeID string, maxDepth int) (*graph.DeletionResult, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	result := *m.deletedNode
	result.DeletedNodes = 3 // Simulate cascade deletion
	return &result, nil
}

func (m *mockPrunerService) DeleteRelation(ctx context.Context, sourceID, targetID, relationType string) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	return nil
}

func (m *mockPrunerService) PreviewDeletion(ctx context.Context, nodeID string, cascade bool, maxDepth int) (*graph_pruner.DeletionPreview, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.preview, nil
}

func (m *mockPrunerService) Close() error {
	return nil
}

var _ = Describe("PruneGraphUseCase", func() {
	var (
		useCase     prune_graph.PruneGraphUseCase
		mockService *mockPrunerService
		ctx         context.Context
	)

	BeforeEach(func() {
		mockService = newMockPrunerService()
		useCase = prune_graph.NewPruneGraphUseCase(mockService)
		ctx = context.Background()
	})

	Describe("DeleteNode", func() {
		Context("with valid request", func() {
			It("should delete node successfully", func() {
				request := &prune_graph.DeleteNodeRequest{
					NodeID: "test-node",
				}

				response, err := useCase.DeleteNode(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.NodeID).To(Equal("test-node"))
				Expect(response.Result.DeletedNodes).To(Equal(1))
				Expect(response.Result.DeletedNodeIDs).To(ContainElement("test-node"))
			})
		})

		Context("with invalid request", func() {
			It("should return error for nil request", func() {
				response, err := useCase.DeleteNode(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request cannot be nil"))
				Expect(response).To(BeNil())
			})

			It("should return error for empty nodeID", func() {
				request := &prune_graph.DeleteNodeRequest{
					NodeID: "",
				}

				response, err := useCase.DeleteNode(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nodeID cannot be empty"))
				Expect(response).To(BeNil())
			})
		})
	})

	Describe("DeleteNodeCascade", func() {
		Context("with valid request", func() {
			It("should delete node cascade successfully", func() {
				request := &prune_graph.DeleteNodeCascadeRequest{
					NodeID:   "test-node",
					MaxDepth: 5,
				}

				response, err := useCase.DeleteNodeCascade(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.NodeID).To(Equal("test-node"))
				Expect(response.MaxDepth).To(Equal(5))
				Expect(response.Result.DeletedNodes).To(Equal(3))
			})

			It("should use default maxDepth when not provided", func() {
				request := &prune_graph.DeleteNodeCascadeRequest{
					NodeID:   "test-node",
					MaxDepth: 0,
				}

				response, err := useCase.DeleteNodeCascade(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.MaxDepth).To(Equal(5)) // Default value
			})
		})

		Context("with invalid request", func() {
			It("should return error for nil request", func() {
				response, err := useCase.DeleteNodeCascade(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request cannot be nil"))
				Expect(response).To(BeNil())
			})

			It("should return error for empty nodeID", func() {
				request := &prune_graph.DeleteNodeCascadeRequest{
					NodeID: "",
				}

				response, err := useCase.DeleteNodeCascade(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nodeID cannot be empty"))
				Expect(response).To(BeNil())
			})

			It("should return error for excessive maxDepth", func() {
				request := &prune_graph.DeleteNodeCascadeRequest{
					NodeID:   "test-node",
					MaxDepth: 25,
				}

				response, err := useCase.DeleteNodeCascade(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("maxDepth cannot exceed 20"))
				Expect(response).To(BeNil())
			})
		})
	})

	Describe("DeleteRelation", func() {
		Context("with valid request", func() {
			It("should delete relation successfully", func() {
				request := &prune_graph.DeleteRelationRequest{
					SourceID:     "source-node",
					TargetID:     "target-node",
					RelationType: "navigation_link",
				}

				response, err := useCase.DeleteRelation(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.SourceID).To(Equal("source-node"))
				Expect(response.TargetID).To(Equal("target-node"))
				Expect(response.RelationType).To(Equal("navigation_link"))
				Expect(response.Success).To(BeTrue())
			})
		})

		Context("with invalid request", func() {
			It("should return error for nil request", func() {
				response, err := useCase.DeleteRelation(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request cannot be nil"))
				Expect(response).To(BeNil())
			})

			It("should return error for same source and target", func() {
				request := &prune_graph.DeleteRelationRequest{
					SourceID:     "same-node",
					TargetID:     "same-node",
					RelationType: "navigation_link",
				}

				response, err := useCase.DeleteRelation(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sourceID and targetID cannot be the same"))
				Expect(response).To(BeNil())
			})
		})
	})

	Describe("PreviewDeletion", func() {
		Context("with valid request", func() {
			It("should return preview successfully", func() {
				request := &prune_graph.PreviewDeletionRequest{
					NodeID:   "test-node",
					Cascade:  false,
					MaxDepth: 0,
				}

				response, err := useCase.PreviewDeletion(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.NodeID).To(Equal("test-node"))
				Expect(response.Cascade).To(BeFalse())
				Expect(response.Preview.TargetNode.ID).To(Equal("test-node"))
			})

			It("should use default maxDepth for cascade preview", func() {
				request := &prune_graph.PreviewDeletionRequest{
					NodeID:   "test-node",
					Cascade:  true,
					MaxDepth: 0,
				}

				response, err := useCase.PreviewDeletion(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Cascade).To(BeTrue())
			})
		})

		Context("with invalid request", func() {
			It("should return error for nil request", func() {
				response, err := useCase.PreviewDeletion(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request cannot be nil"))
				Expect(response).To(BeNil())
			})

			It("should return error for empty nodeID", func() {
				request := &prune_graph.PreviewDeletionRequest{
					NodeID: "",
				}

				response, err := useCase.PreviewDeletion(ctx, request)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nodeID cannot be empty"))
				Expect(response).To(BeNil())
			})
		})
	})
})