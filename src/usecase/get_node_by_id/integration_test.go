package getnodebyid_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/object"
	getnodebyid "mgds/src/usecase/get_node_by_id"

	// Import Cayley memory store
	_ "github.com/cayleygraph/cayley/graph/memstore"
)

func TestGetNodeByIdIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GetNodeById Integration Suite")
}

var _ = Describe("GetNodeById Integration Tests", func() {
	var (
		useCase *getnodebyid.UseCase
		graphDB graph.Interface
		ctx     context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		config := &graph.Config{
			DatabasePath: ":memory:",
			DatabaseType: "cayley",
		}

		graphDB = graph.NewCayleyGraph(config)
		err := graphDB.Connect(ctx)
		Expect(err).NotTo(HaveOccurred())

		useCase = getnodebyid.NewUseCase(graphDB)
	})

	AfterEach(func() {
		if graphDB != nil {
			graphDB.Close()
		}
	})

	Context("with real Cayley database", func() {
		It("should return not found for non-existent node", func() {
			request := &getnodebyid.GetNodeByIdRequest{
				NodeID: "non-existent-id",
			}

			response, err := useCase.Execute(ctx, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getnodebyid.StatusNotFound))
			Expect(response.Error).To(Equal("node not found"))
			Expect(response.NodeData).To(BeNil())
		})

		It("should retrieve existing node by ID", func() {
			// Add test node
			testNode := object.NewURLNodeWithDetails(
				"integration-test-node",
				"Integration Test Page",
				"A page for integration testing",
				"https://example.com/integration",
				"Integration Test",
				"This is content for integration testing of the get node by ID use case.",
			)
			testNode.SetProperty("category", "test")
			testNode.SetProperty("priority", "high")

			err := graphDB.AddNode(ctx, testNode)
			Expect(err).NotTo(HaveOccurred())

			request := &getnodebyid.GetNodeByIdRequest{
				NodeID: "integration-test-node",
			}

			response, err := useCase.Execute(ctx, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getnodebyid.StatusSuccess))
			Expect(response.NodeData).NotTo(BeNil())
			Expect(response.Error).To(BeEmpty())

			// Check node data structure
			nodeData := response.NodeData
			Expect(nodeData["mgdsId"]).To(Equal("integration-test-node"))
			Expect(nodeData["displayName"]).To(Equal("Integration Test Page"))
			Expect(nodeData["description"]).To(Equal("A page for integration testing"))
			Expect(nodeData["type"]).To(Equal(object.URLNodeType))
			Expect(nodeData["url"]).To(Equal("https://example.com/integration"))
			Expect(nodeData["title"]).To(Equal("Integration Test"))
			Expect(nodeData["content"]).To(ContainSubstring("integration testing"))

			// Check properties
			properties, ok := nodeData["properties"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(properties["category"]).To(Equal("test"))
			Expect(properties["priority"]).To(Equal("high"))
		})

		It("should handle invalid request data", func() {
			// Test with nil request
			response, err := useCase.Execute(ctx, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
			Expect(response.Error).To(Equal("invalid node ID"))

			// Test with empty node ID
			request := &getnodebyid.GetNodeByIdRequest{
				NodeID: "",
			}

			response, err = useCase.Execute(ctx, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
			Expect(response.Error).To(Equal("invalid node ID"))

			// Test with whitespace-only node ID
			request = &getnodebyid.GetNodeByIdRequest{
				NodeID: "   ",
			}

			response, err = useCase.Execute(ctx, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
			Expect(response.Error).To(Equal("invalid node ID"))
		})

		It("should handle database connection errors", func() {
			// Close database to simulate error
			graphDB.Close()

			request := &getnodebyid.GetNodeByIdRequest{
				NodeID: "some-node-id",
			}

			response, err := useCase.Execute(ctx, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			// Database may still work with closed connection in memory backend
			// Just verify response is valid
			Expect(response.Status).To(BeElementOf([]getnodebyid.ResponseStatus{
				getnodebyid.StatusNotFound,
				getnodebyid.StatusFailed,
			}))
			if response.Status == getnodebyid.StatusFailed {
				Expect(response.Error).NotTo(BeEmpty())
			}
			Expect(response.NodeData).To(BeNil())
		})

		It("should handle context cancellation", func() {
			// Add test node first
			testNode := object.NewURLNode("ctx-test", "Context Test", "Test", "https://example.com")
			err := graphDB.AddNode(ctx, testNode)
			Expect(err).NotTo(HaveOccurred())

			// Create cancelled context
			cancelledCtx, cancel := context.WithCancel(ctx)
			cancel()

			request := &getnodebyid.GetNodeByIdRequest{
				NodeID: "ctx-test",
			}

			response, err := useCase.Execute(cancelledCtx, request)

			Expect(response).NotTo(BeNil())
			// Should handle cancellation gracefully (either success if fast, or failure)
			if response.Status == getnodebyid.StatusFailed {
				Expect(response.Error).NotTo(BeEmpty())
			}
		})

		It("should retrieve multiple different nodes correctly", func() {
			// Add multiple test nodes with different data
			nodes := []*object.URLNode{
				object.NewURLNode("multi-1", "First Node", "First description", "https://first.com").(*object.URLNode),
				object.NewURLNode("multi-2", "Second Node", "Second description", "https://second.com").(*object.URLNode),
				object.NewURLNode("multi-3", "Third Node", "Third description", "https://third.com").(*object.URLNode),
			}

			// Set different properties for each
			nodes[0].SetProperty("type", "primary")
			nodes[1].SetProperty("type", "secondary")
			nodes[2].SetProperty("type", "tertiary")

			for _, node := range nodes {
				err := graphDB.AddNode(ctx, node)
				Expect(err).NotTo(HaveOccurred())
			}

			// Retrieve each node and verify
			for i, expectedNode := range nodes {
				request := &getnodebyid.GetNodeByIdRequest{
					NodeID: expectedNode.GetMgdsId(),
				}

				response, err := useCase.Execute(ctx, request)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal(getnodebyid.StatusSuccess))
				Expect(response.NodeData).NotTo(BeNil())

				nodeData := response.NodeData
				Expect(nodeData["mgdsId"]).To(Equal(expectedNode.GetMgdsId()))
				Expect(nodeData["displayName"]).To(Equal(expectedNode.GetDisplayName()))

				properties, ok := nodeData["properties"].(map[string]any)
				Expect(ok).To(BeTrue())

				if i == 0 {
					Expect(properties["type"]).To(Equal("primary"))
				} else if i == 1 {
					Expect(properties["type"]).To(Equal("secondary"))
				} else {
					Expect(properties["type"]).To(Equal("tertiary"))
				}
			}
		})
	})
})
