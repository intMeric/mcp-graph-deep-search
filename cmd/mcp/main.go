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

	"mgds/internal/app/services/graph_explorer"
	"mgds/internal/app/services/graph_pruner"
	"mgds/internal/app/services/link"
	"mgds/internal/app/use-cases/analyze_webpage"
	"mgds/internal/app/use-cases/get_document"
	"mgds/internal/app/use-cases/graph/explore_graph"
	"mgds/internal/app/use-cases/graph/prune_graph"
	"mgds/internal/app/use-cases/search_and_analyze"
	"mgds/internal/pkg/configuration"
	"mgds/internal/pkg/database"
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
	"mgds/internal/pkg/scrapper"
)

// Global variables for use cases - initialized once
var exploreGraphUseCase explore_graph.ExploreGraphUseCase
var searchAndAnalyzeUseCase search_and_analyze.SearchAndAnalyzeUseCase
var analyzeWebpageUseCase analyze_webpage.AnalyzeWebpageUseCase
var pruneGraphUseCase prune_graph.PruneGraphUseCase
var getDocumentUseCase get_document.GetDocumentUseCase

func main() {
	// Add debug logging to stderr so it doesn't interfere with stdio transport
	log.SetOutput(os.Stderr)
	log.Printf("Starting MCP Deep Search server...")

	ctx := context.Background()

	// Load environment variables from .env file if it exists
	envFile := ".env"
	if len(os.Args) > 1 {
		envFile = os.Args[1]
	}
	log.Printf("Loading environment variables from: %s", envFile)
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}

	// Initialize graph database connection
	log.Printf("Initializing graph database connection...")
	graphDB, err := graph.NewGraph(graph.Neo4jGraphType)
	if err != nil {
		log.Printf("Failed to initialize graph database: %v", err)
		// Continue anyway to allow MCP server to start - tools will return errors
	} else {
		defer func() {
			if err := graphDB.Close(ctx); err != nil {
				log.Printf("Error closing graph database: %v", err)
			}
		}()
		log.Printf("Graph database initialized successfully")
	}

	// Initialize services and use cases
	log.Printf("Initializing services and use cases...")
	if graphDB != nil {
		graphExplorerService := graph_explorer.NewGraphExplorerService(graphDB)
		exploreGraphUseCase = explore_graph.NewExploreGraphUseCase(graphExplorerService)
	} else {
		log.Printf("Warning: Graph database not available, graph exploration tools will return errors")
	}

	// Initialize search and analyze use case
	log.Printf("Initializing search and analyze use case...")
	var err2 error
	searchAndAnalyzeUseCase, err2 = search_and_analyze.NewSearchAndAnalyzeUseCase()
	if err2 != nil {
		log.Printf("Warning: Failed to initialize search and analyze use case: %v", err2)
		log.Printf("Search and analyze tools will return errors")
	} else {
		log.Printf("Search and analyze use case initialized successfully")
	}

	// Initialize database for document operations
	log.Printf("Initializing database...")
	db, err3 := database.NewMongoDatabase("mgds") // Database name, collections determined by node type
	if err3 != nil {
		log.Printf("Warning: Failed to initialize database: %v", err3)
		log.Printf("Document-related tools will return errors")
	} else {
		log.Printf("Database initialized successfully")
	}

	// Initialize analyze webpage use case
	log.Printf("Initializing analyze webpage use case...")
	if graphDB != nil && db != nil {
		webScraper := scrapper.NewWebScraper()
		linkService := link.NewDirectLinkService(graphDB)

		analyzeWebpageUseCase = analyze_webpage.NewAnalyzeWebpageUseCase(
			webScraper,
			linkService,
			db,
		)
		log.Printf("Analyze webpage use case initialized successfully")
	} else {
		log.Printf("Warning: Dependencies not available for analyze webpage use case")
		log.Printf("Analyze webpage tools will return errors")
	}

	// Initialize prune graph use case
	log.Printf("Initializing prune graph use case...")
	if graphDB != nil && db != nil {
		prunerService := graph_pruner.NewGraphPrunerService(graphDB, db)
		pruneGraphUseCase = prune_graph.NewPruneGraphUseCase(prunerService)
		log.Printf("Prune graph use case initialized successfully")
	} else {
		log.Printf("Warning: Dependencies not available for prune graph use case")
		log.Printf("Prune graph tools will return errors")
	}

	// Initialize get document use case
	log.Printf("Initializing get document use case...")
	if db != nil && analyzeWebpageUseCase != nil {
		getDocumentUseCase = get_document.NewGetDocumentUseCase(
			db,
			analyzeWebpageUseCase,
		)
		log.Printf("Get document use case initialized successfully")
	} else {
		log.Printf("Warning: Dependencies not available for get document use case")
		log.Printf("Get document tools will return errors")
	}

	// Create channel to keep server alive (based on official example)
	done := make(chan struct{})

	// Create MCP server with stdio transport
	log.Printf("Creating MCP server...")
	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())

	// Register tools
	log.Printf("Registering tools...")
	if err := registerGraphExplorationTools(server); err != nil {
		log.Fatalf("Failed to register graph exploration tools: %v", err)
	}

	if err := registerSearchAndAnalyzeTools(server); err != nil {
		log.Fatalf("Failed to register search and analyze tools: %v", err)
	}

	if err := registerAnalyzeWebpageTools(server); err != nil {
		log.Fatalf("Failed to register analyze webpage tools: %v", err)
	}

	if err := registerPruneGraphTools(server); err != nil {
		log.Fatalf("Failed to register prune graph tools: %v", err)
	}

	if err := registerGetDocumentTools(server); err != nil {
		log.Fatalf("Failed to register get document tools: %v", err)
	}

	log.Printf("MCP server ready, starting to serve...")

	// Start the server (this should run in background according to the official example)
	err = server.Serve()
	if err != nil {
		log.Fatalf("MCP server error: %v", err)
	}

	log.Printf("Server started successfully, waiting...")
	// Keep the server alive (this is the key part from the official example)
	<-done
}

// Define argument types at package level to avoid reflection issues
type EmptyArgs struct{}

type GetNodesByTypeArgs struct {
	NodeType string `json:"nodeType" jsonschema:"required,description=Type of nodes to retrieve"`
	Offset   int    `json:"offset" jsonschema:"description=Offset for pagination (default: 0),minimum=0"`
	Limit    int    `json:"limit" jsonschema:"description=Number of nodes to return (default: 10),minimum=1,maximum=100"`
}

type GetNodeRelationsArgs struct {
	NodeID string `json:"nodeId" jsonschema:"required,description=ID of the node to get relations for"`
}

type GetConnectedNodesArgs struct {
	NodeID string `json:"nodeId" jsonschema:"required,description=ID of the node to get connected nodes for"`
}

// Search and analyze tool arguments
type SearchAndAnalyzeArgs struct {
	Query      string `json:"query" jsonschema:"required,description=Search query to execute"`
	TimeRange  string `json:"timeRange" jsonschema:"description=Time range for search (day, week, month, year)"`
	Category   string `json:"category" jsonschema:"description=Search category: general, news, rs (default: general)"`
	Language   string `json:"language" jsonschema:"description=Search language (default: en)"`
	MaxResults int    `json:"maxResults" jsonschema:"description=Maximum number of results to analyze (default: 10),minimum=1,maximum=50"`
}

// Analyze webpage tool arguments
type AnalyzeWebpageArgs struct {
	URL string `json:"url" jsonschema:"required,description=URL of the webpage to analyze"`
}

// Prune graph tool arguments
type DeleteNodeArgs struct {
	NodeID string `json:"nodeId" jsonschema:"required,description=ID of the node to delete"`
}

type DeleteNodeCascadeArgs struct {
	NodeID   string `json:"nodeId" jsonschema:"required,description=ID of the node to delete with all descendants"`
	MaxDepth int    `json:"maxDepth" jsonschema:"description=Maximum depth for cascade deletion (default: 5),minimum=1,maximum=20"`
}

type DeleteRelationArgs struct {
	SourceID     string `json:"sourceId" jsonschema:"required,description=ID of the source node"`
	TargetID     string `json:"targetId" jsonschema:"required,description=ID of the target node"`
	RelationType string `json:"relationType" jsonschema:"required,description=Type of relation to delete"`
}

type PreviewDeletionArgs struct {
	NodeID   string `json:"nodeId" jsonschema:"required,description=ID of the node to preview deletion for"`
	Cascade  bool   `json:"cascade" jsonschema:"description=Whether to include descendants in preview"`
	MaxDepth int    `json:"maxDepth" jsonschema:"description=Maximum depth for cascade preview (default: 5),minimum=1,maximum=20"`
}

// Get document tool arguments
type GetDocumentArgs struct {
	Node *node.Node `json:"node" jsonschema:"required,description=Node with ID to retrieve or analyze document for"`
}

func registerGraphExplorationTools(server *mcp_golang.Server) error {
	// Tool: Get Graph Overview
	err := server.RegisterTool("get_graph_overview", "Get knowledge graph overview with statistics and node types for LLM exploration planning",
		func(args EmptyArgs) (*mcp_golang.ToolResponse, error) {
			if exploreGraphUseCase == nil {
				return nil, fmt.Errorf("graph database not available")
			}

			ctx := context.Background()
			response, err := exploreGraphUseCase.GetGraphOverview(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get graph overview: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register get_graph_overview tool: %w", err)
	}

	// Tool: Get Nodes by Type
	err = server.RegisterTool("get_nodes_by_type", "Get paginated nodes of a specific type for knowledge graph exploration",
		func(args GetNodesByTypeArgs) (*mcp_golang.ToolResponse, error) {
			if exploreGraphUseCase == nil {
				return nil, fmt.Errorf("graph database not available")
			}

			ctx := context.Background()

			// Set defaults
			if args.Offset < 0 {
				args.Offset = 0
			}
			if args.Limit <= 0 {
				args.Limit = 10
			}
			if args.Limit > 100 {
				args.Limit = 100
			}

			req := &explore_graph.GetNodesByTypeRequest{
				NodeType: args.NodeType,
				Offset:   args.Offset,
				Limit:    args.Limit,
			}

			response, err := exploreGraphUseCase.GetNodesByType(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get nodes by type: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register get_nodes_by_type tool: %w", err)
	}

	// Tool: Get Node Relations
	err = server.RegisterTool("get_node_relations", "Get all relations for a specific node to enable deep graph traversal",
		func(args GetNodeRelationsArgs) (*mcp_golang.ToolResponse, error) {
			if exploreGraphUseCase == nil {
				return nil, fmt.Errorf("graph database not available")
			}

			ctx := context.Background()

			req := &explore_graph.GetNodeRelationsRequest{
				NodeID: args.NodeID,
			}

			response, err := exploreGraphUseCase.GetNodeRelations(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get node relations: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register get_node_relations tool: %w", err)
	}

	// Tool: Get Connected Nodes
	err = server.RegisterTool("get_connected_nodes", "Get all nodes connected to a specific node for comprehensive knowledge discovery",
		func(args GetConnectedNodesArgs) (*mcp_golang.ToolResponse, error) {
			if exploreGraphUseCase == nil {
				return nil, fmt.Errorf("graph database not available")
			}

			ctx := context.Background()

			req := &explore_graph.GetConnectedNodesRequest{
				NodeID: args.NodeID,
			}

			response, err := exploreGraphUseCase.GetConnectedNodes(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get connected nodes: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register get_connected_nodes tool: %w", err)
	}

	return nil
}

func registerSearchAndAnalyzeTools(server *mcp_golang.Server) error {
	// Tool: Search and Analyze
	err := server.RegisterTool("search_and_analyze", "Perform deep search using SearXNG, analyze content, and build explorable knowledge graph for LLM discovery",
		func(args SearchAndAnalyzeArgs) (*mcp_golang.ToolResponse, error) {
			if searchAndAnalyzeUseCase == nil {
				return nil, fmt.Errorf("search and analyze use case not available")
			}

			ctx := context.Background()

			// Get SearXNG configuration
			searxngConfig := configuration.NewSearXNGConfig()

			category := args.Category
			if category == "" {
				category = searxngConfig.Category
			}

			language := args.Language
			if language == "" {
				language = searxngConfig.Language
			}

			maxResults := args.MaxResults
			if maxResults <= 0 {
				maxResults = 10
			}
			if maxResults > 50 {
				maxResults = 50
			}

			req := &search_and_analyze.SearchAndAnalyzeRequest{
				Query:      args.Query,
				TimeRange:  args.TimeRange,
				Category:   category,
				Language:   language,
				MaxResults: maxResults,
			}

			response, err := searchAndAnalyzeUseCase.Execute(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to execute search and analyze: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register search_and_analyze tool: %w", err)
	}

	return nil
}

func registerAnalyzeWebpageTools(server *mcp_golang.Server) error {
	// Tool: Analyze Webpage
	err := server.RegisterTool("analyze_webpage", "Analyze a specific webpage for content, PII, keywords, and build knowledge graph relationships",
		func(args AnalyzeWebpageArgs) (*mcp_golang.ToolResponse, error) {
			if analyzeWebpageUseCase == nil {
				return nil, fmt.Errorf("analyze webpage use case not available")
			}

			ctx := context.Background()

			req := &analyze_webpage.AnalyzeWebpageRequest{
				URL: args.URL,
			}

			response, err := analyzeWebpageUseCase.Execute(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze webpage: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register analyze_webpage tool: %w", err)
	}

	return nil
}

func registerPruneGraphTools(server *mcp_golang.Server) error {
	// Tool: Delete Node
	err := server.RegisterTool("delete_node", "Delete a specific node from the knowledge graph, including all its relationships and associated documents",
		func(args DeleteNodeArgs) (*mcp_golang.ToolResponse, error) {
			if pruneGraphUseCase == nil {
				return nil, fmt.Errorf("prune graph use case not available")
			}

			ctx := context.Background()

			req := &prune_graph.DeleteNodeRequest{
				NodeID: args.NodeID,
			}

			response, err := pruneGraphUseCase.DeleteNode(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to delete node: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register delete_node tool: %w", err)
	}

	// Tool: Delete Node Cascade
	err = server.RegisterTool("delete_node_cascade", "Delete a node and all its descendants from the knowledge graph, with configurable depth limit",
		func(args DeleteNodeCascadeArgs) (*mcp_golang.ToolResponse, error) {
			if pruneGraphUseCase == nil {
				return nil, fmt.Errorf("prune graph use case not available")
			}

			ctx := context.Background()

			req := &prune_graph.DeleteNodeCascadeRequest{
				NodeID:   args.NodeID,
				MaxDepth: args.MaxDepth,
			}

			response, err := pruneGraphUseCase.DeleteNodeCascade(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to delete node cascade: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register delete_node_cascade tool: %w", err)
	}

	// Tool: Delete Relation
	err = server.RegisterTool("delete_relation", "Delete a specific relationship between two nodes in the knowledge graph",
		func(args DeleteRelationArgs) (*mcp_golang.ToolResponse, error) {
			if pruneGraphUseCase == nil {
				return nil, fmt.Errorf("prune graph use case not available")
			}

			ctx := context.Background()

			req := &prune_graph.DeleteRelationRequest{
				SourceID:     args.SourceID,
				TargetID:     args.TargetID,
				RelationType: args.RelationType,
			}

			response, err := pruneGraphUseCase.DeleteRelation(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to delete relation: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register delete_relation tool: %w", err)
	}

	// Tool: Preview Deletion
	err = server.RegisterTool("preview_deletion", "Preview what nodes and relationships would be affected by a deletion operation before executing it",
		func(args PreviewDeletionArgs) (*mcp_golang.ToolResponse, error) {
			if pruneGraphUseCase == nil {
				return nil, fmt.Errorf("prune graph use case not available")
			}

			ctx := context.Background()

			req := &prune_graph.PreviewDeletionRequest{
				NodeID:   args.NodeID,
				Cascade:  args.Cascade,
				MaxDepth: args.MaxDepth,
			}

			response, err := pruneGraphUseCase.PreviewDeletion(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to preview deletion: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register preview_deletion tool: %w", err)
	}

	return nil
}

func registerGetDocumentTools(server *mcp_golang.Server) error {
	// Tool: Get Document
	err := server.RegisterTool("get_document", "Intelligently retrieve existing document or automatically analyze webpage if no document exists. Perfect for LLMs needing document content with automatic fallback to analysis.",
		func(args GetDocumentArgs) (*mcp_golang.ToolResponse, error) {
			if getDocumentUseCase == nil {
				return nil, fmt.Errorf("get document use case not available")
			}

			ctx := context.Background()

			if args.Node == nil {
				return nil, fmt.Errorf("node is required")
			}

			req := &get_document.GetDocumentRequest{
				Node:        args.Node,
				AutoAnalyze: true,
			}

			response, err := getDocumentUseCase.Execute(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to execute get document: %w", err)
			}

			content, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp_golang.NewToolResponse(
				mcp_golang.NewTextContent(string(content)),
			), nil
		})
	if err != nil {
		return fmt.Errorf("failed to register get_document tool: %w", err)
	}

	return nil
}
