package link_test

import (
	"context"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/services/link"
	"mgds/internal/constant"
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

type testGraph struct {
	shouldFail       bool
	shouldFailExists bool
	nodeExists       bool
	relations        []*graph.Relation
	createdNodes     []*node.Node
	closed           bool
}

func (m *testGraph) CreateNode(ctx context.Context, node *node.Node) error {
	if m.shouldFail {
		return errors.New("node creation failed")
	}
	m.createdNodes = append(m.createdNodes, node)
	return nil
}

func (m *testGraph) CreateRelation(ctx context.Context, relation *graph.Relation) error {
	if m.shouldFail {
		return errors.New("graph creation failed")
	}
	m.relations = append(m.relations, relation)
	return nil
}

func (m *testGraph) GetNode(ctx context.Context, id string) (*node.Node, error) {
	return nil, nil
}

func (m *testGraph) NodeExists(ctx context.Context, id string) (bool, error) {
	if m.shouldFailExists {
		return false, errors.New("node exists check failed")
	}
	return m.nodeExists, nil
}

func (m *testGraph) CreateLink(ctx context.Context, sourceNode *node.Node, targetNode *node.Node, relationType string) (*graph.Relation, error) {
	if m.shouldFail {
		return nil, errors.New("graph creation failed")
	}
	// Graph package should handle validation
	if sourceNode == nil || targetNode == nil {
		return nil, errors.New("nodes cannot be nil")
	}
	if sourceNode.ID == targetNode.ID {
		return nil, errors.New("sourceNode and targetNode cannot have the same ID")
	}
	if err := sourceNode.Validate(); err != nil {
		return nil, err
	}
	if err := targetNode.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(relationType) == "" {
		return nil, errors.New("relation type cannot be empty")
	}

	relation := &graph.Relation{
		Type:     relationType,
		SourceID: sourceNode.ID,
		TargetID: targetNode.ID,
	}
	m.relations = append(m.relations, relation)
	return relation, nil
}

func (m *testGraph) Close(ctx context.Context) error {
	m.closed = true
	return nil
}

func (m *testGraph) GetGraphStats(ctx context.Context) (*graph.GraphStats, error) {
	return nil, nil
}

func (m *testGraph) GetNodesByType(ctx context.Context, nodeType string, offset, limit int) (*graph.PaginatedNodesResult, error) {
	return nil, nil
}

func (m *testGraph) GetNodeRelations(ctx context.Context, nodeID string) ([]*graph.Relation, error) {
	return nil, nil
}

func (m *testGraph) GetConnectedNodes(ctx context.Context, nodeID string) ([]*graph.NodeConnection, error) {
	return nil, nil
}

type testQueue struct {
	messages   []*link.QueuedLinkRequest
	shouldFail bool
	closed     bool
}

func (m *testQueue) PublishMessage(msg *link.QueuedLinkRequest) error {
	if m.shouldFail {
		return errors.New("queue publish failed")
	}
	m.messages = append(m.messages, msg)
	return nil
}

func (m *testQueue) ConsumeMessages(handler func(*link.QueuedLinkRequest) error) error {
	return nil
}

func (m *testQueue) Close() error {
	m.closed = true
	return nil
}

var _ = Describe("LinkService", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("DirectLinkService", func() {
		var (
			service   link.LinkService
			graphMock *testGraph
		)

		BeforeEach(func() {
			graphMock = &testGraph{nodeExists: true}
			service = link.NewDirectLinkService(graphMock)
		})

		AfterEach(func() {
			if service != nil {
				service.Close()
			}
		})

		Describe("CreateLink", func() {
			Context("with valid nodes", func() {
				It("should create relation successfully with existing nodes", func() {
					sourceNode := &node.Node{
						Type:        "URL",
						DisplayName: "Source Node",
						ID:          "source-1",
					}
					targetNode := &node.Node{
						Type:        "URL",
						DisplayName: "Target Node",
						ID:          "target-1",
					}

					response, err := service.CreateLink(ctx, sourceNode, targetNode, constant.NavigationLink)

					Expect(err).NotTo(HaveOccurred())
					Expect(response).NotTo(BeNil())
					Expect(response.Success).To(BeTrue())
					Expect(response.Message).To(Equal("Link created successfully"))
					Expect(response.Relation).NotTo(BeNil())
					Expect(response.Relation.SourceID).To(Equal("source-1"))
					Expect(response.Relation.TargetID).To(Equal("target-1"))
					Expect(response.Relation.Type).To(Equal(constant.NavigationLink))
					Expect(response.SourceCreated).To(BeFalse())
					Expect(response.TargetCreated).To(BeFalse())
					Expect(len(graphMock.relations)).To(Equal(1))
					Expect(len(graphMock.createdNodes)).To(Equal(0))
				})

				It("should create nodes and relation when nodes don't exist", func() {
					graphMock.nodeExists = false
					sourceNode := &node.Node{
						Type:        "URL",
						DisplayName: "Source Node",
						ID:          "source-1",
					}
					targetNode := &node.Node{
						Type:        "URL",
						DisplayName: "Target Node",
						ID:          "target-1",
					}

					response, err := service.CreateLink(ctx, sourceNode, targetNode, constant.NavigationLink)

					Expect(err).NotTo(HaveOccurred())
					Expect(response).NotTo(BeNil())
					Expect(response.Success).To(BeTrue())
					Expect(response.SourceCreated).To(BeTrue())
					Expect(response.TargetCreated).To(BeTrue())
					Expect(len(graphMock.createdNodes)).To(Equal(2))
					Expect(len(graphMock.relations)).To(Equal(1))
				})
			})

			Context("with invalid input", func() {
				It("should let graph package handle validation", func() {
					sameNode := &node.Node{
						Type:        "URL",
						DisplayName: "Same Node",
						ID:          "same-id",
					}

					response, err := service.CreateLink(ctx, sameNode, sameNode, constant.NavigationLink)

					Expect(err).To(HaveOccurred())
					Expect(response).To(BeNil())
				})
			})

			Context("when graph fails", func() {
				It("should return error when CreateLink fails", func() {
					graphMock.shouldFail = true
					sourceNode := &node.Node{
						Type:        "URL",
						DisplayName: "Source Node",
						ID:          "source-1",
					}
					targetNode := &node.Node{
						Type:        "URL",
						DisplayName: "Target Node",
						ID:          "target-1",
					}

					response, err := service.CreateLink(ctx, sourceNode, targetNode, constant.NavigationLink)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("graph creation failed"))
					Expect(response).To(BeNil())
				})

				It("should return error when node creation fails", func() {
					graphMock.nodeExists = false
					graphMock.shouldFail = true
					sourceNode := &node.Node{
						Type:        "URL",
						DisplayName: "Source Node",
						ID:          "source-1",
					}
					targetNode := &node.Node{
						Type:        "URL",
						DisplayName: "Target Node",
						ID:          "target-1",
					}

					response, err := service.CreateLink(ctx, sourceNode, targetNode, constant.NavigationLink)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("node creation failed"))
					Expect(response).To(BeNil())
				})
			})
		})

		Describe("Close", func() {
			It("should close graph successfully", func() {
				err := service.Close()

				Expect(err).NotTo(HaveOccurred())
				Expect(graphMock.closed).To(BeTrue())
			})
		})
	})

	Describe("QueuedLinkService", func() {
		var (
			service   link.LinkService
			queueMock *testQueue
		)

		BeforeEach(func() {
			queueMock = &testQueue{}
			service = link.NewQueuedLinkService(queueMock)
		})

		AfterEach(func() {
			if service != nil {
				service.Close()
			}
		})

		Describe("CreateLink", func() {
			Context("with valid nodes", func() {
				It("should queue message successfully", func() {
					sourceNode := &node.Node{
						Type:        "URL",
						DisplayName: "Source Node",
						ID:          "source-1",
					}
					targetNode := &node.Node{
						Type:        "URL",
						DisplayName: "Target Node",
						ID:          "target-1",
					}

					response, err := service.CreateLink(ctx, sourceNode, targetNode, constant.NavigationLink)

					Expect(err).NotTo(HaveOccurred())
					Expect(response).NotTo(BeNil())
					Expect(response.Success).To(BeTrue())
					Expect(response.Message).To(Equal("Link request queued successfully"))
					Expect(response.Relation).To(BeNil())
					Expect(len(queueMock.messages)).To(Equal(1))
					Expect(queueMock.messages[0].SourceNode.ID).To(Equal("source-1"))
					Expect(queueMock.messages[0].TargetNode.ID).To(Equal("target-1"))
					Expect(queueMock.messages[0].LinkType).To(Equal(constant.NavigationLink))
				})
			})

			Context("with valid input", func() {
				It("should queue any input - let graph handle validation", func() {
					sameNode := &node.Node{
						Type:        "URL",
						DisplayName: "Same Node",
						ID:          "same-id",
					}

					response, err := service.CreateLink(ctx, sameNode, sameNode, constant.NavigationLink)

					Expect(err).NotTo(HaveOccurred())
					Expect(response).NotTo(BeNil())
					Expect(response.Success).To(BeTrue())
					Expect(response.Message).To(Equal("Link request queued successfully"))
					Expect(len(queueMock.messages)).To(Equal(1))
				})
			})

			Context("when queue fails", func() {
				It("should return error", func() {
					queueMock.shouldFail = true
					sourceNode := &node.Node{
						Type:        "URL",
						DisplayName: "Source Node",
						ID:          "source-1",
					}
					targetNode := &node.Node{
						Type:        "URL",
						DisplayName: "Target Node",
						ID:          "target-1",
					}

					response, err := service.CreateLink(ctx, sourceNode, targetNode, constant.NavigationLink)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("queue publish failed"))
					Expect(response).To(BeNil())
				})
			})
		})

		Describe("Close", func() {
			It("should close queue successfully", func() {
				err := service.Close()

				Expect(err).NotTo(HaveOccurred())
				Expect(queueMock.closed).To(BeTrue())
			})
		})
	})
})
