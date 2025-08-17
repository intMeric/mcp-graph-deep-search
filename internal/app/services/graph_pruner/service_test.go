package graph_pruner_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/graph_pruner"
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

// Mock implementations for testing
type mockGraph struct {
	nodes       map[string]*node.Node
	relations   map[string][]*graph.Relation
	descendants map[string][]*node.Node
	shouldError bool
}

func newMockGraph() *mockGraph {
	return &mockGraph{
		nodes:       make(map[string]*node.Node),
		relations:   make(map[string][]*graph.Relation),
		descendants: make(map[string][]*node.Node),
	}
}

func (m *mockGraph) addNode(n *node.Node) {
	m.nodes[n.ID] = n
}

func (m *mockGraph) addRelation(nodeID string, rel *graph.Relation) {
	m.relations[nodeID] = append(m.relations[nodeID], rel)
}

func (m *mockGraph) addRelations(nodeID string, relations []*graph.Relation) {
	m.relations[nodeID] = relations
}

func (m *mockGraph) addDescendants(nodeID string, descendants []*node.Node) {
	m.descendants[nodeID] = descendants
}

func (m *mockGraph) CreateNode(ctx context.Context, node *node.Node) error {
	return fmt.Errorf("not implemented")
}

func (m *mockGraph) CreateRelation(ctx context.Context, relation *graph.Relation) error {
	return fmt.Errorf("not implemented")
}

func (m *mockGraph) CreateLink(ctx context.Context, sourceNode *node.Node, targetNode *node.Node, relationType string) (*graph.Relation, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockGraph) GetNode(ctx context.Context, id string) (*node.Node, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	if n, exists := m.nodes[id]; exists {
		return n, nil
	}
	return nil, fmt.Errorf("node not found")
}

func (m *mockGraph) NodeExists(ctx context.Context, id string) (bool, error) {
	_, exists := m.nodes[id]
	return exists, nil
}

func (m *mockGraph) Close(ctx context.Context) error {
	return nil
}

func (m *mockGraph) GetGraphStats(ctx context.Context) (*graph.GraphStats, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockGraph) GetNodesByType(ctx context.Context, nodeType string, offset, limit int) (*graph.PaginatedNodesResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockGraph) GetNodeRelations(ctx context.Context, nodeID string) ([]*graph.Relation, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.relations[nodeID], nil
}

func (m *mockGraph) GetConnectedNodes(ctx context.Context, nodeID string) ([]*graph.NodeConnection, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockGraph) DeleteNode(ctx context.Context, nodeID string) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	if _, exists := m.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found")
	}
	delete(m.nodes, nodeID)
	return nil
}

func (m *mockGraph) DeleteRelation(ctx context.Context, sourceID, targetID, relationType string) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	return nil
}

func (m *mockGraph) GetDescendantNodes(ctx context.Context, nodeID string, maxDepth int) ([]*node.Node, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	return m.descendants[nodeID], nil
}

type mockDatabase struct {
	documents   map[string]map[string]interface{}
	shouldError bool
}

func newMockDatabase() *mockDatabase {
	return &mockDatabase{
		documents: make(map[string]map[string]interface{}),
	}
}

func (m *mockDatabase) Insert(ctx context.Context, id string, document interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockDatabase) FindByID(ctx context.Context, collection, id string, dest interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockDatabase) Update(ctx context.Context, collection, id string, update interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockDatabase) Delete(ctx context.Context, collection, id string) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	if m.documents[collection] != nil {
		delete(m.documents[collection], id)
	}
	return nil
}

func (m *mockDatabase) Close(ctx context.Context) error {
	return nil
}

func (m *mockDatabase) GetLocation() string {
	return "test_location"
}

var _ = Describe("GraphPrunerService", func() {
	var (
		service    graph_pruner.GraphPrunerService
		mockGraph  *mockGraph
		mockDB     *mockDatabase
		ctx        context.Context
		testNode   *node.Node
		testNode2  *node.Node
		testNode3  *node.Node
	)

	BeforeEach(func() {
		mockGraph = newMockGraph()
		mockDB = newMockDatabase()
		service = graph_pruner.NewGraphPrunerService(mockGraph, mockDB)
		ctx = context.Background()

		testNode = &node.Node{
			ID:          "test-node-1",
			Type:        "webpage",
			DisplayName: "Test Node 1",
			Location:    "webpage",
		}

		testNode2 = &node.Node{
			ID:          "test-node-2",
			Type:        "webpage",
			DisplayName: "Test Node 2",
			Location:    "",
		}

		testNode3 = &node.Node{
			ID:          "test-node-3",
			Type:        "webpage",
			DisplayName: "Test Node 3",
			Location:    "webpage",
		}

		mockGraph.addNode(testNode)
		mockGraph.addNode(testNode2)
		mockGraph.addNode(testNode3)
	})

	AfterEach(func() {
		if service != nil {
			service.Close()
		}
	})

	Describe("DeleteNode", func() {
		Context("with valid node", func() {
			It("should delete node successfully", func() {
				result, err := service.DeleteNode(ctx, testNode.ID)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.DeletedNodes).To(Equal(1))
				Expect(result.DeletedNodeIDs).To(ContainElement(testNode.ID))
			})

			It("should handle node without location", func() {
				result, err := service.DeleteNode(ctx, testNode2.ID)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.DeletedNodes).To(Equal(1))
				Expect(result.DeletedNodeIDs).To(ContainElement(testNode2.ID))
			})
		})

		Context("with invalid input", func() {
			It("should return error for empty nodeID", func() {
				result, err := service.DeleteNode(ctx, "")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nodeID cannot be empty"))
				Expect(result).To(BeNil())
			})

			It("should return error for non-existent node", func() {
				result, err := service.DeleteNode(ctx, "non-existent")

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("DeleteNodeCascade", func() {
		Context("with descendants", func() {
			BeforeEach(func() {
				descendants := []*node.Node{testNode2, testNode3}
				mockGraph.addDescendants(testNode.ID, descendants)
			})

			It("should delete node and all descendants", func() {
				result, err := service.DeleteNodeCascade(ctx, testNode.ID, 5)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.DeletedNodes).To(Equal(3)) // Root + 2 descendants
				Expect(result.DeletedNodeIDs).To(ContainElement(testNode.ID))
				Expect(result.DeletedNodeIDs).To(ContainElement(testNode2.ID))
				Expect(result.DeletedNodeIDs).To(ContainElement(testNode3.ID))
			})
		})

		Context("with invalid input", func() {
			It("should return error for empty nodeID", func() {
				result, err := service.DeleteNodeCascade(ctx, "", 5)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nodeID cannot be empty"))
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("DeleteRelation", func() {
		Context("with valid parameters", func() {
			It("should delete relation successfully", func() {
				err := service.DeleteRelation(ctx, testNode.ID, testNode2.ID, "navigation_link")

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with invalid input", func() {
			It("should return error for empty sourceID", func() {
				err := service.DeleteRelation(ctx, "", testNode2.ID, "navigation_link")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sourceID cannot be empty"))
			})

			It("should return error for empty targetID", func() {
				err := service.DeleteRelation(ctx, testNode.ID, "", "navigation_link")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("targetID cannot be empty"))
			})

			It("should return error for empty relationType", func() {
				err := service.DeleteRelation(ctx, testNode.ID, testNode2.ID, "")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("relationType cannot be empty"))
			})
		})
	})

	Describe("PreviewDeletion", func() {
		Context("for single node", func() {
			BeforeEach(func() {
				relations := []*graph.Relation{
					{Type: "navigation_link", SourceID: testNode.ID, TargetID: testNode2.ID},
				}
				mockGraph.addRelations(testNode.ID, relations)
			})

			It("should return preview for single node deletion", func() {
				preview, err := service.PreviewDeletion(ctx, testNode.ID, false, 0)

				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.TargetNode.ID).To(Equal(testNode.ID))
				Expect(preview.TotalNodes).To(Equal(1))
				Expect(preview.TotalRelations).To(Equal(1))
				Expect(preview.HasDocuments).To(BeTrue())
			})
		})

		Context("for cascade deletion", func() {
			BeforeEach(func() {
				descendants := []*node.Node{testNode2, testNode3}
				mockGraph.addDescendants(testNode.ID, descendants)
			})

			It("should return preview for cascade deletion", func() {
				preview, err := service.PreviewDeletion(ctx, testNode.ID, true, 5)

				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.TargetNode.ID).To(Equal(testNode.ID))
				Expect(preview.TotalNodes).To(Equal(3)) // Root + 2 descendants
				Expect(preview.AffectedNodes).To(HaveLen(2))
			})
		})

		Context("with invalid input", func() {
			It("should return error for empty nodeID", func() {
				preview, err := service.PreviewDeletion(ctx, "", false, 0)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nodeID cannot be empty"))
				Expect(preview).To(BeNil())
			})
		})
	})
})