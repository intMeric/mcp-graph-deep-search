package getnodebyid_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/node"
	getnodebyid "mgds/src/usecase/get_node_by_id"
)

// MockGraphDB implements graph.Interface for testing
type MockGraphDB struct {
	node      node.Interface
	exists    bool
	getErr    error
	existsErr error
}

func (m *MockGraphDB) AddNode(ctx context.Context, node node.Interface) error { return nil }
func (m *MockGraphDB) GetNode(ctx context.Context, mgdsId string) (node.Interface, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.node, nil
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
func (m *MockGraphDB) NodeExists(ctx context.Context, mgdsId string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	return m.exists, nil
}
func (m *MockGraphDB) GetAllNodes(ctx context.Context) ([]node.Interface, error) { return nil, nil }
func (m *MockGraphDB) Connect(ctx context.Context) error                         { return nil }
func (m *MockGraphDB) Close() error                                              { return nil }
func (m *MockGraphDB) Ping(ctx context.Context) error                            { return nil }

// MockNode implements node.Interface for testing
type MockNode struct {
	id            string
	displayName   string
	description   string
	nodeType      string
	properties    map[string]any
	serializeErr  error
	deserializeErr error
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
	if m.serializeErr != nil {
		return nil, m.serializeErr
	}
	return map[string]any{
		"id":          m.id,
		"displayName": m.displayName,
		"description": m.description,
		"type":        m.nodeType,
		"properties":  m.properties,
	}, nil
}

func (m *MockNode) Deserialize(data map[string]any) error {
	if m.deserializeErr != nil {
		return m.deserializeErr
	}
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

var _ = Describe("GetNodeById UseCase", func() {
	var (
		useCase *getnodebyid.UseCase
		mockDB  *MockGraphDB
		ctx     context.Context
	)

	BeforeEach(func() {
		mockDB = &MockGraphDB{}
		useCase = getnodebyid.NewUseCase(mockDB)
		ctx = context.Background()
	})

	Describe("Execute", func() {
		Context("with valid node ID", func() {
			var mockNode *MockNode

			BeforeEach(func() {
				mockNode = NewMockNode("node123", "Test Node", "A test node", "url")
				mockNode.SetProperty("url", "https://example.com")
				mockDB.node = mockNode
				mockDB.exists = true
			})

			It("should return success status with deserialized node data", func() {
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "node123",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(getnodebyid.StatusSuccess))
				Expect(response.NodeData).NotTo(BeEmpty())
				Expect(response.NodeData["id"]).To(Equal("node123"))
				Expect(response.NodeData["displayName"]).To(Equal("Test Node"))
				Expect(response.NodeData["type"]).To(Equal("url"))
				Expect(response.Error).To(BeEmpty())
			})

			It("should include node properties in deserialized data", func() {
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "node123",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.NodeData["properties"]).NotTo(BeNil())
				properties := response.NodeData["properties"].(map[string]any)
				Expect(properties["url"]).To(Equal("https://example.com"))
			})
		})

		Context("with non-existent node ID", func() {
			BeforeEach(func() {
				mockDB.exists = false
			})

			It("should return not found status", func() {
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "nonexistent",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(getnodebyid.StatusNotFound))
				Expect(response.NodeData).To(BeNil())
				Expect(response.Error).To(Equal("node not found"))
			})
		})

		Context("with invalid request", func() {
			It("should return failed status for nil request", func() {
				response, err := useCase.Execute(ctx, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
				Expect(response.Error).To(Equal("invalid node ID"))
			})

			It("should return failed status for empty node ID", func() {
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
				Expect(response.Error).To(Equal("invalid node ID"))
			})

			It("should return failed status for whitespace-only node ID", func() {
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "   ",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
				Expect(response.Error).To(Equal("invalid node ID"))
			})
		})

		Context("with database errors", func() {
			It("should return failed status when NodeExists fails", func() {
				mockDB.existsErr = errors.New("database connection error")
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "node123",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
				Expect(response.Error).To(Equal("database connection error"))
			})

			It("should return failed status when GetNode fails", func() {
				mockDB.exists = true
				mockDB.getErr = errors.New("failed to retrieve node")
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "node123",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
				Expect(response.Error).To(Equal("failed to retrieve node"))
			})

			It("should return not found status when GetNode returns nil", func() {
				mockDB.exists = true
				mockDB.node = nil
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "node123",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(getnodebyid.StatusNotFound))
				Expect(response.Error).To(Equal("node not found"))
			})
		})

		Context("with serialization errors", func() {
			BeforeEach(func() {
				mockNode := NewMockNode("node123", "Test Node", "A test node", "url")
				mockNode.serializeErr = errors.New("serialization failed")
				mockDB.node = mockNode
				mockDB.exists = true
			})

			It("should return failed status when node serialization fails", func() {
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: "node123",
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
				Expect(response.Error).To(Equal("serialization failed"))
			})
		})
	})

	Describe("NewUseCase", func() {
		It("should create a new service instance", func() {
			newUseCase := getnodebyid.NewUseCase(mockDB)

			Expect(newUseCase).NotTo(BeNil())
		})
	})
})