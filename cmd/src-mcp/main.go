package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/scrapper"
	"mgds/src/pkg/search_engine"
	"mgds/src/usecase/get_all_nodes"
	"mgds/src/usecase/get_node_by_id"
	"mgds/src/usecase/web_search"

	// Import Cayley memory store
	_ "github.com/cayleygraph/cayley/graph/memstore"
)

type GetAllNodesArgs struct{}

type GetNodeByIdArgs struct {
	NodeID string `json:"nodeId" jsonschema:"required,description=ID of the node to retrieve"`
}

type WebSearchArgs struct {
	Query      string `json:"query" jsonschema:"required,description=Search query to execute"`
	MaxResults int    `json:"maxResults" jsonschema:"description=Maximum number of results (default: 10),minimum=1,maximum=100"`
}

func main() {
	log.SetOutput(os.Stderr)
	log.Printf("Starting src-mcp server...")

	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	// Initialize graph database with in-memory store for testing
	config := &graph.Config{
		DatabasePath: ":memory:",
		DatabaseType: "cayley",
	}
	
	graphDB := graph.NewCayleyGraph(config)
	if graphDB != nil {
		ctx := context.Background()
		if err := graphDB.Connect(ctx); err != nil {
			log.Printf("Failed to connect to graph database: %v", err)
			graphDB = nil
		}
	}

	// Initialize search engine and scraper for web search
	searchConfig := &search_engine.Config{
		BaseURL: "http://localhost:8888",
	}
	searchEngine := search_engine.NewSearXNG(searchConfig)
	
	scraper := scrapper.NewCollyScrapper(nil)

	// Create MCP server
	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())

	// Register tools
	if err := registerGetAllNodesTools(server, graphDB); err != nil {
		log.Fatalf("Failed to register get_all_nodes tool: %v", err)
	}

	if err := registerGetNodeByIdTools(server, graphDB); err != nil {
		log.Fatalf("Failed to register get_node_by_id tool: %v", err)
	}

	if err := registerWebSearchTools(server, searchEngine, scraper, graphDB); err != nil {
		log.Fatalf("Failed to register web_search tool: %v", err)
	}

	log.Printf("src-mcp server ready")

	// Create channel to keep server alive
	done := make(chan struct{})

	// Start the server
	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
	}()

	// Keep the server alive
	<-done
}

func registerGetAllNodesTools(server *mcp_golang.Server, graphDB graph.Interface) error {
	return server.RegisterTool("get_all_nodes", "Retrieve all nodes from the graph database",
		func(args GetAllNodesArgs) (*mcp_golang.ToolResponse, error) {
			if graphDB == nil {
				return nil, fmt.Errorf("graph database not available")
			}

			ctx := context.Background()
			service := getallnodes.NewService(graphDB)

			response, err := service.Execute(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get all nodes: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
}

func registerGetNodeByIdTools(server *mcp_golang.Server, graphDB graph.Interface) error {
	return server.RegisterTool("get_node_by_id", "Retrieve a specific node by its ID",
		func(args GetNodeByIdArgs) (*mcp_golang.ToolResponse, error) {
			if graphDB == nil {
				return nil, fmt.Errorf("graph database not available")
			}

			ctx := context.Background()
			service := getnodebyid.NewService(graphDB)

			request := &getnodebyid.GetNodeByIdRequest{
				NodeID: args.NodeID,
			}

			response, err := service.Execute(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("failed to get node by id: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
}

func registerWebSearchTools(server *mcp_golang.Server, searchEngine search_engine.Interface, scraper scrapper.Interface, graphDB graph.Interface) error {
	return server.RegisterTool("web_search", "Search the web and index results in the graph",
		func(args WebSearchArgs) (*mcp_golang.ToolResponse, error) {
			if graphDB == nil {
				return nil, fmt.Errorf("graph database not available")
			}

			ctx := context.Background()
			factory := websearch.NewFactory(searchEngine, scraper, graphDB)
			service := factory.CreateService()

			maxResults := args.MaxResults
			if maxResults <= 0 {
				maxResults = 10
			}
			if maxResults > 100 {
				maxResults = 100
			}

			request := &websearch.SearchRequest{
				Query:      args.Query,
				MaxResults: maxResults,
			}

			response, err := service.Execute(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("failed to execute web search: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
}