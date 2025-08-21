package graph_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/pkg/graph"
	"mgds/pkg/object"
)

var _ = Describe("CayleyGraph", func() {
	var (
		g   graph.Interface
		ctx context.Context
	)

	BeforeEach(func() {
		config := &graph.Config{
			DatabasePath: ":memory:",
			DatabaseType: "cayley",
		}
		g = graph.NewCayleyGraph(config)
		ctx = context.Background()
		
		err := g.Connect(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if g != nil {
			g.Close()
		}
	})

	Describe("Connection Management", func() {
		Context("when connecting to database", func() {
			It("should connect successfully", func() {
				err := g.Ping(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Node Operations", func() {
		var testNode *object.URLNode

		BeforeEach(func() {
			node := object.NewURLNode("test-1", "Test Page", "A test page", "https://example.com")
			testNode = node.(*object.URLNode)
		})

		Context("adding nodes", func() {
			It("should add a node successfully", func() {
				err := g.AddNode(ctx, testNode)
				Expect(err).NotTo(HaveOccurred())

				exists, err := g.NodeExists(ctx, testNode.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})
		})

		Context("retrieving nodes", func() {
			BeforeEach(func() {
				err := g.AddNode(ctx, testNode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should get an existing node", func() {
				retrievedNode, err := g.GetNode(ctx, testNode.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(retrievedNode).NotTo(BeNil())
				Expect(retrievedNode.GetMgdsId()).To(Equal(testNode.GetMgdsId()))
				Expect(retrievedNode.GetType()).To(Equal(object.URLNodeType))
			})

			It("should return error for non-existent node", func() {
				_, err := g.GetNode(ctx, "non-existent")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("updating nodes", func() {
			BeforeEach(func() {
				err := g.AddNode(ctx, testNode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should update an existing node", func() {
				testNode.SetProperty("updated", true)
				
				err := g.UpdateNode(ctx, testNode)
				Expect(err).NotTo(HaveOccurred())

				retrievedNode, err := g.GetNode(ctx, testNode.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(retrievedNode).NotTo(BeNil())
			})
		})

		Context("deleting nodes", func() {
			BeforeEach(func() {
				err := g.AddNode(ctx, testNode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should delete an existing node", func() {
				err := g.DeleteNode(ctx, testNode.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())

				exists, err := g.NodeExists(ctx, testNode.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})
	})

	Describe("Relation Operations", func() {
		var node1, node2 *object.URLNode

		BeforeEach(func() {
			n1 := object.NewURLNode("node-1", "Node 1", "First node", "https://example1.com")
			n2 := object.NewURLNode("node-2", "Node 2", "Second node", "https://example2.com")
			node1 = n1.(*object.URLNode)
			node2 = n2.(*object.URLNode)

			err := g.AddNode(ctx, node1)
			Expect(err).NotTo(HaveOccurred())
			err = g.AddNode(ctx, node2)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("adding relations", func() {
			It("should add a relation between nodes", func() {
				relation := &graph.Relation{
					FromNodeId: node1.GetMgdsId(),
					ToNodeId:   node2.GetMgdsId(),
					Label:      "links_to",
					Properties: map[string]any{"weight": 1},
				}

				err := g.AddRelation(ctx, relation)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("retrieving relations", func() {
			BeforeEach(func() {
				relation := &graph.Relation{
					FromNodeId: node1.GetMgdsId(),
					ToNodeId:   node2.GetMgdsId(),
					Label:      "links_to",
				}
				err := g.AddRelation(ctx, relation)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should get outgoing relations", func() {
				relations, err := g.GetRelations(ctx, node1.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(relations).To(HaveLen(1))
				Expect(relations[0].FromNodeId).To(Equal(node1.GetMgdsId()))
				Expect(relations[0].ToNodeId).To(Equal(node2.GetMgdsId()))
				Expect(relations[0].Label).To(Equal("links_to"))
			})

			It("should get incoming relations", func() {
				relations, err := g.GetIncomingRelations(ctx, node2.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(relations).To(HaveLen(1))
				Expect(relations[0].FromNodeId).To(Equal(node1.GetMgdsId()))
				Expect(relations[0].ToNodeId).To(Equal(node2.GetMgdsId()))
				Expect(relations[0].Label).To(Equal("links_to"))
			})
		})

		Context("deleting relations", func() {
			BeforeEach(func() {
				relation := &graph.Relation{
					FromNodeId: node1.GetMgdsId(),
					ToNodeId:   node2.GetMgdsId(),
					Label:      "links_to",
				}
				err := g.AddRelation(ctx, relation)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should delete a relation", func() {
				err := g.DeleteRelation(ctx, node1.GetMgdsId(), node2.GetMgdsId(), "links_to")
				Expect(err).NotTo(HaveOccurred())

				relations, err := g.GetRelations(ctx, node1.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(relations).To(BeEmpty())
			})
		})
	})

	Describe("Graph Traversal", func() {
		var node1, node2, node3 *object.URLNode

		BeforeEach(func() {
			n1 := object.NewURLNode("node-1", "Node 1", "First node", "https://example1.com")
			n2 := object.NewURLNode("node-2", "Node 2", "Second node", "https://example2.com")
			n3 := object.NewURLNode("node-3", "Node 3", "Third node", "https://example3.com")
			node1 = n1.(*object.URLNode)
			node2 = n2.(*object.URLNode)
			node3 = n3.(*object.URLNode)

			err := g.AddNode(ctx, node1)
			Expect(err).NotTo(HaveOccurred())
			err = g.AddNode(ctx, node2)
			Expect(err).NotTo(HaveOccurred())
			err = g.AddNode(ctx, node3)
			Expect(err).NotTo(HaveOccurred())

			// Create connections: node1 -> node2 -> node3
			rel1 := &graph.Relation{FromNodeId: node1.GetMgdsId(), ToNodeId: node2.GetMgdsId(), Label: "connects"}
			rel2 := &graph.Relation{FromNodeId: node2.GetMgdsId(), ToNodeId: node3.GetMgdsId(), Label: "connects"}
			
			err = g.AddRelation(ctx, rel1)
			Expect(err).NotTo(HaveOccurred())
			err = g.AddRelation(ctx, rel2)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("finding connected nodes", func() {
			It("should find all connected nodes", func() {
				connected, err := g.GetConnectedNodes(ctx, node2.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(connected).To(HaveLen(2))
				
				ids := make([]string, len(connected))
				for i, n := range connected {
					ids[i] = n.GetMgdsId()
				}
				Expect(ids).To(ContainElement(node1.GetMgdsId()))
				Expect(ids).To(ContainElement(node3.GetMgdsId()))
			})
		})

		Context("finding paths", func() {
			It("should find direct path between nodes", func() {
				path, err := g.FindPath(ctx, node1.GetMgdsId(), node2.GetMgdsId())
				Expect(err).NotTo(HaveOccurred())
				Expect(path).To(HaveLen(1))
				Expect(path[0].FromNodeId).To(Equal(node1.GetMgdsId()))
				Expect(path[0].ToNodeId).To(Equal(node2.GetMgdsId()))
			})
		})
	})

	Describe("Query Operations", func() {
		BeforeEach(func() {
			node1 := object.NewURLNode("url-1", "Page 1", "First page", "https://example1.com")
			node2 := object.NewURLNode("url-2", "Page 2", "Second page", "https://example2.com")
			
			err := g.AddNode(ctx, node1)
			Expect(err).NotTo(HaveOccurred())
			err = g.AddNode(ctx, node2)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("finding nodes by type", func() {
			It("should find all URL nodes", func() {
				nodes, err := g.FindNodesByType(ctx, object.URLNodeType)
				Expect(err).NotTo(HaveOccurred())
				Expect(nodes).To(HaveLen(2))
				
				for _, node := range nodes {
					Expect(node.GetType()).To(Equal(object.URLNodeType))
				}
			})
		})

		Context("getting all nodes", func() {
			It("should return all nodes in the graph", func() {
				nodes, err := g.GetAllNodes(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(nodes).To(HaveLen(2))
			})
		})
	})
})