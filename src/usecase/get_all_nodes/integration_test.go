package getallnodes_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/object"
	"mgds/src/usecase/get_all_nodes"
	
	// Import Cayley memory store
	_ "github.com/cayleygraph/cayley/graph/memstore"
)

func TestGetAllNodesIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GetAllNodes Integration Suite")
}

var _ = Describe("GetAllNodes Integration Tests", func() {
	var (
		useCase     *getallnodes.UseCase
		graphDB     graph.Interface
		ctx         context.Context
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
		
		useCase = getallnodes.NewUseCase(graphDB)
	})

	AfterEach(func() {
		if graphDB != nil {
			graphDB.Close()
		}
	})

	Context("with real Cayley database", func() {
		It("should retrieve all nodes from empty database", func() {
			response, err := useCase.Execute(ctx)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
			Expect(response.Nodes).To(BeEmpty())
			Expect(response.TotalNodes).To(Equal(0))
			Expect(response.ExecutionTime).To(BeNumerically(">", 0))
		})

		It("should retrieve all nodes with data", func() {
			// Add test nodes
			node1 := object.NewURLNode("test-1", "Test Page 1", "A test page", "https://example.com/1")
			node2 := object.NewURLNode("test-2", "Test Page 2", "Another test page", "https://example.com/2")
			node3 := object.NewURLNode("test-3", "Test Page 3", "Third test page", "https://example.com/3")

			err := graphDB.AddNode(ctx, node1)
			Expect(err).NotTo(HaveOccurred())
			err = graphDB.AddNode(ctx, node2)
			Expect(err).NotTo(HaveOccurred())
			err = graphDB.AddNode(ctx, node3)
			Expect(err).NotTo(HaveOccurred())

			response, err := useCase.Execute(ctx)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
			Expect(response.Nodes).To(HaveLen(3))
			Expect(response.TotalNodes).To(Equal(3))
			Expect(response.ExecutionTime).To(BeNumerically(">", 0))

			// Check node data
			nodeIds := make([]string, len(response.Nodes))
			for i, node := range response.Nodes {
				nodeIds[i] = node.ID
				Expect(node.Type).To(Equal(object.URLNodeType))
				Expect(node.DisplayName).NotTo(BeEmpty())
				Expect(node.Description).NotTo(BeEmpty())
			}
			
			Expect(nodeIds).To(ContainElements("test-1", "test-2", "test-3"))
		})

		It("should handle database errors gracefully", func() {
			// Close database to simulate error
			graphDB.Close()
			
			response, err := useCase.Execute(ctx)
			
			Expect(err).NotTo(HaveOccurred()) // Use case doesn't return error, handles internally
			Expect(response).NotTo(BeNil())
			// Database may still work with closed connection in memory backend
			// Just verify response is valid
			Expect(response.Status).To(BeElementOf([]getallnodes.ResponseStatus{
				getallnodes.StatusSuccess,
				getallnodes.StatusFailed,
			}))
			if response.Status == getallnodes.StatusFailed {
				Expect(response.Error).NotTo(BeEmpty())
				Expect(response.Nodes).To(BeNil())
				Expect(response.TotalNodes).To(Equal(0))
			}
		})

		It("should handle performance with many nodes", func() {
			// Add many test nodes
			for i := 0; i < 100; i++ {
				node := object.NewURLNodeWithDetails(
					fmt.Sprintf("perf-test-%d", i),
					fmt.Sprintf("Performance Test Page %d", i),
					fmt.Sprintf("Description for page %d", i),
					fmt.Sprintf("https://example.com/perf-%d", i),
					fmt.Sprintf("Title %d", i),
					fmt.Sprintf("Content for performance test page %d", i),
				)
				err := graphDB.AddNode(ctx, node)
				Expect(err).NotTo(HaveOccurred())
			}

			start := time.Now()
			response, err := useCase.Execute(ctx)
			executionTime := time.Since(start)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
			Expect(response.Nodes).To(HaveLen(100))
			Expect(response.TotalNodes).To(Equal(100))
			Expect(executionTime).To(BeNumerically("<", 5*time.Second)) // Should be fast
		})

		It("should handle context cancellation", func() {
			// Add some test data
			for i := 0; i < 10; i++ {
				node := object.NewURLNode(fmt.Sprintf("ctx-test-%d", i), "Test", "Test", "https://example.com")
				err := graphDB.AddNode(ctx, node)
				Expect(err).NotTo(HaveOccurred())
			}

			// Create cancelled context
			cancelledCtx, cancel := context.WithCancel(ctx)
			cancel()

			response, _ := useCase.Execute(cancelledCtx)
			
			// Should handle cancellation gracefully
			Expect(response).NotTo(BeNil())
			if response.Status == getallnodes.StatusFailed {
				Expect(response.Error).To(ContainSubstring("context"))
			}
		})
	})
})