package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/object"
	"mgds/src/pkg/scrapper"
	"mgds/src/pkg/search_engine"
	getallnodes "mgds/src/usecase/get_all_nodes"
	getnodebyid "mgds/src/usecase/get_node_by_id"
	websearch "mgds/src/usecase/web_search"

	// Import Cayley memory store
	_ "github.com/cayleygraph/cayley/graph/memstore"
)

func TestServerIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MCP Server Integration Suite")
}

var _ = Describe("MCP Server Components Integration", func() {
	var (
		graphDB      graph.Interface
		searchEngine search_engine.Interface
		scraper      scrapper.Interface
		ctx          context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Initialize graph database
		config := &graph.Config{
			DatabasePath: ":memory:",
			DatabaseType: "cayley",
		}
		graphDB = graph.NewCayleyGraph(config)
		err := graphDB.Connect(ctx)
		Expect(err).NotTo(HaveOccurred())

		// Initialize search engine and scraper
		searchConfig := &search_engine.Config{
			BaseURL: "http://localhost:8888",
			Timeout: 5 * time.Second,
		}
		searchEngine = search_engine.NewSearXNG(searchConfig)
		scraper = scrapper.NewCollyScrapper(nil)
	})

	AfterEach(func() {
		if searchEngine != nil {
			searchEngine.Close()
		}
		if scraper != nil {
			scraper.Close()
		}
		if graphDB != nil {
			graphDB.Close()
		}
	})

	Context("Server Dependencies", func() {
		It("should initialize all components successfully", func() {
			Expect(graphDB).NotTo(BeNil())
			Expect(searchEngine).NotTo(BeNil())
			Expect(scraper).NotTo(BeNil())

			// Test database connection
			err := graphDB.Ping(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("get_all_nodes Use Case Integration", func() {
		It("should work with empty database", func() {
			useCase := getallnodes.NewUseCase(graphDB)
			response, err := useCase.Execute(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
			Expect(response.TotalNodes).To(Equal(0))
			Expect(response.Nodes).To(HaveLen(0))
			Expect(response.ExecutionTime).To(BeNumerically(">", 0))
		})

		It("should retrieve nodes after population", func() {
			// Add test nodes
			node1 := object.NewURLNode("srv-test-1", "Server Test 1", "Description 1", "https://example.com/1")
			node2 := object.NewURLNode("srv-test-2", "Server Test 2", "Description 2", "https://example.com/2")

			err := graphDB.AddNode(ctx, node1)
			Expect(err).NotTo(HaveOccurred())
			err = graphDB.AddNode(ctx, node2)
			Expect(err).NotTo(HaveOccurred())

			useCase := getallnodes.NewUseCase(graphDB)
			response, err := useCase.Execute(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
			Expect(response.TotalNodes).To(Equal(2))
			Expect(response.Nodes).To(HaveLen(2))

			nodeIds := make([]string, len(response.Nodes))
			for i, node := range response.Nodes {
				nodeIds[i] = node.ID
			}
			Expect(nodeIds).To(ContainElements("srv-test-1", "srv-test-2"))
		})

		It("should handle large datasets efficiently", func() {
			// Add many nodes
			for i := 0; i < 100; i++ {
				node := object.NewURLNode(
					fmt.Sprintf("perf-%d", i),
					fmt.Sprintf("Performance Test %d", i),
					fmt.Sprintf("Description %d", i),
					fmt.Sprintf("https://example.com/perf-%d", i),
				)
				err := graphDB.AddNode(ctx, node)
				Expect(err).NotTo(HaveOccurred())
			}

			start := time.Now()
			useCase := getallnodes.NewUseCase(graphDB)
			response, err := useCase.Execute(ctx)
			duration := time.Since(start)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal(getallnodes.StatusSuccess))
			Expect(response.TotalNodes).To(Equal(100))
			Expect(duration).To(BeNumerically("<", 2*time.Second))
		})
	})

	Context("get_node_by_id Use Case Integration", func() {
		It("should return not found for non-existent node", func() {
			useCase := getnodebyid.NewUseCase(graphDB)
			request := &getnodebyid.GetNodeByIdRequest{
				NodeID: "non-existent-node",
			}

			response, err := useCase.Execute(ctx, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal(getnodebyid.StatusNotFound))
			Expect(response.Error).To(Equal("node not found"))
			Expect(response.NodeData).To(BeNil())
		})

		It("should retrieve existing node with all data", func() {
			// Add test node with rich data
			testNode := object.NewURLNodeWithDetails(
				"detailed-test",
				"Detailed Test Page",
				"A comprehensive test description",
				"https://example.com/detailed",
				"Detailed Test Title",
				"Rich content for testing node retrieval with full details",
			)
			testNode.SetProperty("category", "integration")
			testNode.SetProperty("priority", "high")
			testNode.SetProperty("tags", []string{"test", "integration", "server"})

			err := graphDB.AddNode(ctx, testNode)
			Expect(err).NotTo(HaveOccurred())

			useCase := getnodebyid.NewUseCase(graphDB)
			request := &getnodebyid.GetNodeByIdRequest{
				NodeID: "detailed-test",
			}

			response, err := useCase.Execute(ctx, request)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal(getnodebyid.StatusSuccess))
			Expect(response.NodeData).NotTo(BeNil())

			nodeData := response.NodeData
			Expect(nodeData["mgdsId"]).To(Equal("detailed-test"))
			Expect(nodeData["displayName"]).To(Equal("Detailed Test Page"))
			Expect(nodeData["description"]).To(Equal("A comprehensive test description"))
			Expect(nodeData["type"]).To(Equal(object.URLNodeType))
			Expect(nodeData["url"]).To(Equal("https://example.com/detailed"))
			Expect(nodeData["title"]).To(Equal("Detailed Test Title"))
			Expect(nodeData["content"]).To(ContainSubstring("Rich content"))

			properties, ok := nodeData["properties"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(properties["category"]).To(Equal("integration"))
			Expect(properties["priority"]).To(Equal("high"))
		})

		It("should handle invalid requests", func() {
			useCase := getnodebyid.NewUseCase(graphDB)

			// Test nil request
			response, err := useCase.Execute(ctx, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
			Expect(response.Error).To(Equal("invalid node ID"))

			// Test empty ID
			request := &getnodebyid.GetNodeByIdRequest{NodeID: ""}
			response, err = useCase.Execute(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal(getnodebyid.StatusFailed))
		})
	})

	Context("web_search Use Case Integration", func() {
		It("should validate queries properly", func() {
			factory := websearch.NewFactory(searchEngine, scraper, graphDB)
			useCase := factory.CreateUseCase()

			// Test empty query
			request := &websearch.SearchRequest{
				Query:      "",
				MaxResults: 5,
			}

			response, err := useCase.Execute(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal(websearch.StatusFailed))
			Expect(response.Error).To(Equal("Query cannot be empty"))

			// Test whitespace query
			request.Query = "   "
			response, err = useCase.Execute(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal(websearch.StatusFailed))
		})

		It("should handle MaxResults validation", func() {
			factory := websearch.NewFactory(searchEngine, scraper, graphDB)
			useCase := factory.CreateUseCase()

			// Test with 0 max results (should default to 10)
			request := &websearch.SearchRequest{
				Query:      "test query",
				MaxResults: 0,
			}

			response, err := useCase.Execute(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Query).To(Equal("test query"))
			// Response will likely fail due to no SearchXNG, but query should be processed
		})

		It("should handle search service unavailable", func() {
			factory := websearch.NewFactory(searchEngine, scraper, graphDB)
			useCase := factory.CreateUseCase()

			request := &websearch.SearchRequest{
				Query:      "unavailable service test",
				MaxResults: 3,
			}

			response, err := useCase.Execute(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			// SearchXNG might be unavailable (expected) or available (unexpected but valid)
			Expect(response.Status).To(BeElementOf([]string{
				websearch.StatusFailed,
				websearch.StatusSuccess,
			}))
			if response.Status == websearch.StatusFailed {
				Expect(response.Error).To(ContainSubstring("Search and index failed"))
			}
			Expect(response.Query).To(Equal("unavailable service test"))
			Expect(response.ExecutionTime).To(BeNumerically(">", 0))
		})
	})

	Context("Cross-Use Case Integration", func() {
		It("should maintain data consistency across use cases", func() {
			// Add nodes through direct database access
			node1 := object.NewURLNode("cross-1", "Cross Test 1", "Description 1", "https://example.com/cross-1")
			node2 := object.NewURLNode("cross-2", "Cross Test 2", "Description 2", "https://example.com/cross-2")

			err := graphDB.AddNode(ctx, node1)
			Expect(err).NotTo(HaveOccurred())
			err = graphDB.AddNode(ctx, node2)
			Expect(err).NotTo(HaveOccurred())

			// Verify through get_all_nodes
			getAllUseCase := getallnodes.NewUseCase(graphDB)
			allResponse, err := getAllUseCase.Execute(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(allResponse.TotalNodes).To(Equal(2))

			// Verify through get_node_by_id for each node
			getByIdUseCase := getnodebyid.NewUseCase(graphDB)

			request1 := &getnodebyid.GetNodeByIdRequest{NodeID: "cross-1"}
			response1, err := getByIdUseCase.Execute(ctx, request1)
			Expect(err).NotTo(HaveOccurred())
			Expect(response1.Status).To(Equal(getnodebyid.StatusSuccess))
			Expect(response1.NodeData["displayName"]).To(Equal("Cross Test 1"))

			request2 := &getnodebyid.GetNodeByIdRequest{NodeID: "cross-2"}
			response2, err := getByIdUseCase.Execute(ctx, request2)
			Expect(err).NotTo(HaveOccurred())
			Expect(response2.Status).To(Equal(getnodebyid.StatusSuccess))
			Expect(response2.NodeData["displayName"]).To(Equal("Cross Test 2"))
		})

		It("should handle concurrent access properly", func() {
			// Add initial nodes
			for i := 0; i < 10; i++ {
				node := object.NewURLNode(
					fmt.Sprintf("concurrent-%d", i),
					fmt.Sprintf("Concurrent Test %d", i),
					fmt.Sprintf("Description %d", i),
					fmt.Sprintf("https://example.com/concurrent-%d", i),
				)
				err := graphDB.AddNode(ctx, node)
				Expect(err).NotTo(HaveOccurred())
			}

			// Test concurrent reads
			done := make(chan bool, 5)
			for i := 0; i < 5; i++ {
				go func() {
					defer GinkgoRecover()
					useCase := getallnodes.NewUseCase(graphDB)
					response, err := useCase.Execute(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(response.TotalNodes).To(Equal(10))
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 5; i++ {
				<-done
			}
		})
	})

	Context("Error Recovery", func() {
		It("should handle database connection issues gracefully", func() {
			// Close database
			graphDB.Close()

			// Test all use cases handle this gracefully
			getAllUseCase := getallnodes.NewUseCase(graphDB)
			response1, err := getAllUseCase.Execute(ctx)
			Expect(err).NotTo(HaveOccurred())
			// May succeed or fail depending on implementation, but shouldn't panic

			getByIdUseCase := getnodebyid.NewUseCase(graphDB)
			request := &getnodebyid.GetNodeByIdRequest{NodeID: "test"}
			response2, err := getByIdUseCase.Execute(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			// Should handle closed DB gracefully

			factory := websearch.NewFactory(searchEngine, scraper, graphDB)
			webUseCase := factory.CreateUseCase()
			webRequest := &websearch.SearchRequest{Query: "test", MaxResults: 1}
			response3, err := webUseCase.Execute(ctx, webRequest)
			Expect(err).NotTo(HaveOccurred())
			// Should handle closed DB in the chain

			// All should return valid responses, not panic
			Expect(response1).NotTo(BeNil())
			Expect(response2).NotTo(BeNil())
			Expect(response3).NotTo(BeNil())
		})
	})
})