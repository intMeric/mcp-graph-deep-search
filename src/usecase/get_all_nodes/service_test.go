package getallnodes_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/node"
	getallnodes "mgds/src/usecase/get_all_nodes"
)

// MockGraphDB implements graph.Interface for testing
type MockGraphDB struct {
	nodes []node.Interface
	err   error
}

func (m *MockGraphDB) AddNode(ctx context.Context, node node.Interface) error { return nil }
func (m *MockGraphDB) GetNode(ctx context.Context, mgdsId string) (node.Interface, error) {
	return nil, nil
}
func (m *MockGraphDB) UpdateNode(ctx context.Context, node node.Interface) error       { return nil }
func (m *MockGraphDB) DeleteNode(ctx context.Context, mgdsId string) error             { return nil }
func (m *MockGraphDB) AddRelation(ctx context.Context, relation *graph.Relation) error { return nil }
func (m *MockGraphDB) GetRelations(ctx context.Context, fromNodeId string) ([]*graph.Relation, error) {
	return nil, nil
}
func (m *MockGraphDB) GetIncomingRelations(ctx context.Context, toNodeId string) ([]*graph.Relation, error) {
	return nil, nil
}
func (m *MockGraphDB) DeleteRelation(ctx context.Context, fromNodeId, toNodeId, label string) error {
	return nil
}
func (m *MockGraphDB) GetConnectedNodes(ctx context.Context, mgdsId string) ([]node.Interface, error) {
	return nil, nil
}
func (m *MockGraphDB) FindPath(ctx context.Context, fromNodeId, toNodeId string) ([]*graph.Relation, error) {
	return nil, nil
}
func (m *MockGraphDB) FindNodesByType(ctx context.Context, nodeType string) ([]node.Interface, error) {
	return nil, nil
}
func (m *MockGraphDB) FindNodesByProperty(ctx context.Context, key string, value any) ([]node.Interface, error) {
	return nil, nil
}
func (m *MockGraphDB) NodeExists(ctx context.Context, mgdsId string) (bool, error) { return false, nil }
func (m *MockGraphDB) Connect(ctx context.Context) error                           { return nil }
func (m *MockGraphDB) Close() error                                                { return nil }
func (m *MockGraphDB) Ping(ctx context.Context) error                              { return nil }

func (m *MockGraphDB) GetAllNodes(ctx context.Context) ([]node.Interface, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.nodes, nil
}

// MockNode implements node.Interface for testing
type MockNode struct {
	id          string
	displayName string
	description string
	nodeType    string
	properties  map[string]any
}

func NewMockNode(id, displayName, description, nodeType string) *MockNode {
	return &MockNode{
		id:          id,
		displayName: displayName,
		description: description,
		nodeType:    nodeType,
		properties:  make(map[string]any),
	}
}

func (m *MockNode) GetMgdsId() string      { return m.id }
func (m *MockNode) GetDisplayName() string { return m.displayName }
func (m *MockNode) GetDescription() string { return m.description }
func (m *MockNode) GetType() string        { return m.nodeType }

func (m *MockNode) GetProperty(key string) (any, bool) {
	val, exists := m.properties[key]
	return val, exists
}

func (m *MockNode) SetProperty(key string, value any) {
	m.properties[key] = value
}

func (m *MockNode) Serialize() (map[string]any, error) {
	return map[string]any{
		"id":          m.id,
		"displayName": m.displayName,
		"description": m.description,
		"type":        m.nodeType,
		"properties":  m.properties,
	}, nil
}

func (m *MockNode) Deserialize(data map[string]any) error {
	if id, ok := data["id"].(string); ok {
		m.id = id
	}
	if displayName, ok := data["displayName"].(string); ok {
		m.displayName = displayName
	}
	if description, ok := data["description"].(string); ok {
		m.description = description
	}
	if nodeType, ok := data["type"].(string); ok {
		m.nodeType = nodeType
	}
	return nil
}

func (m *MockNode) IsValid() bool {
	return m.id != "" && m.nodeType != ""
}

var _ = Describe("GetAllNodes UseCase", func() {
	var (
		useCase getallnodes.UseCase
		mockDB  *MockGraphDB
		ctx     context.Context
	)

	BeforeEach(func() {
		mockDB = &MockGraphDB{}
		useCase = getallnodes.NewUseCase(mockDB)
		ctx = context.Background()
	})

	Describe("Execute", func() {
		Context("with successful node retrieval", func() {
			BeforeEach(func() {
				mockDB.nodes = []node.Interface{
					NewMockNode("node1", "Node 1", "Description 1", "url"),
					NewMockNode("node2", "Node 2", "Description 2", "document"),
					NewMockNode("node3", "Node 3", "Description 3", "url"),
				}
			})

			It("should return success status with all nodes", func() {
				response, err := useCase.Execute(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
				Expect(response.TotalNodes).To(Equal(3))
				Expect(response.Nodes).To(HaveLen(3))
				Expect(response.Error).To(BeEmpty())
				Expect(response.ExecutionTime).To(BeNumerically(">", 0))
			})

			It("should convert nodes to proper NodeSummary format", func() {
				response, err := useCase.Execute(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Nodes[0].ID).To(Equal("node1"))
				Expect(response.Nodes[0].DisplayName).To(Equal("Node 1"))
				Expect(response.Nodes[0].Description).To(Equal("Description 1"))
				Expect(response.Nodes[0].Type).To(Equal("url"))

				Expect(response.Nodes[1].ID).To(Equal("node2"))
				Expect(response.Nodes[1].Type).To(Equal("document"))
			})

			It("should measure execution time", func() {
				response, err := useCase.Execute(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.ExecutionTime).To(BeNumerically(">", 0))
				Expect(response.ExecutionTime).To(BeNumerically("<", time.Second))
			})
		})

		Context("with empty node collection", func() {
			BeforeEach(func() {
				mockDB.nodes = []node.Interface{}
			})

			It("should return success status with empty results", func() {
				response, err := useCase.Execute(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
				Expect(response.TotalNodes).To(Equal(0))
				Expect(response.Nodes).To(HaveLen(0))
				Expect(response.Error).To(BeEmpty())
			})
		})

		Context("with nil nodes", func() {
			BeforeEach(func() {
				mockDB.nodes = nil
			})

			It("should return success status with zero count", func() {
				response, err := useCase.Execute(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
				Expect(response.TotalNodes).To(Equal(0))
				Expect(response.Nodes).To(HaveLen(0))
				Expect(response.Error).To(BeEmpty())
			})
		})

		Context("with database error", func() {
			BeforeEach(func() {
				mockDB.err = errors.New("database connection failed")
			})

			It("should return failed status with error message", func() {
				response, err := useCase.Execute(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(getallnodes.StatusFailed))
				Expect(response.TotalNodes).To(Equal(0))
				Expect(response.Nodes).To(BeNil())
				Expect(response.Error).To(Equal("database connection failed"))
				Expect(response.ExecutionTime).To(BeNumerically(">", 0))
			})
		})

		Context("with context cancellation", func() {
			var cancelFunc context.CancelFunc

			BeforeEach(func() {
				ctx, cancelFunc = context.WithCancel(context.Background())
				mockDB.nodes = []node.Interface{
					NewMockNode("node1", "Node 1", "Description 1", "url"),
				}
			})

			AfterEach(func() {
				if cancelFunc != nil {
					cancelFunc()
				}
			})

			It("should handle context cancellation gracefully", func() {
				cancelFunc()
				response, err := useCase.Execute(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
			})
		})
	})

	Describe("NewUseCase", func() {
		It("should create a new service instance", func() {
			newUseCase := getallnodes.NewUseCase(mockDB)

			Expect(newUseCase).NotTo(BeNil())
		})
	})
})
