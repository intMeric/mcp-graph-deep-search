# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

YOUR MOTTO: "Everything should be made as simple as possible, but not simpler. Nothing is more simple than greatness; indeed, to be simple is to be great" **Albert Einstein**

FOLLOW K.I.S.S principle !

You're not just a technician, you're a real software engineer. You can challenge and ask me questions as an equal !

## Development

- If you don't have all the information, ASK !
- If you don't know, ASK !
- No TODOs in the code, no unused functions !
- Comments must be in ENGLISH !
- For each package that is intended to be used by others, always create interfaces. Make sure they are as SIMPLE as possible.

## Testing

- TEST-DRIVEN DEVELOPMENT IS NON-NEGOTIABLE.
- Run all tests: `go test -v ./...`
- Run specific package tests: `go test -v ./internal/pkg/cache`
- Run tests with coverage: `go test -v -cover ./...`
- Run tests for specific file: `go test -v ./internal/pkg/cache -run TestLRUCache`
- Tests use Ginkgo BDD framework with Gomega assertions

### Test Example Structure

All tests must follow the Ginkgo BDD pattern with Gomega assertions:

```go
package mypackage_test

import (
    "context"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "mgds/internal/pkg/mypackage"
)

var _ = Describe("MyComponent", func() {
    var (
        component mypackage.Interface
        ctx       context.Context
    )

    BeforeEach(func() {
        component = mypackage.New()
        ctx = context.Background()
    })

    AfterEach(func() {
        if component != nil {
            component.Close()
        }
    })

    Describe("MethodName", func() {
        Context("with valid input", func() {
            It("should return expected result", func() {
                result, err := component.MethodName(ctx, "input")

                Expect(err).NotTo(HaveOccurred())
                Expect(result).NotTo(BeEmpty())
                Expect(result).To(ContainSubstring("expected"))
            })
        })

        Context("with invalid input", func() {
            It("should handle errors gracefully", func() {
                result, err := component.MethodName(ctx, "")

                Expect(err).To(HaveOccurred())
                Expect(result).To(BeEmpty())
            })
        })
    })
})
```

## Building

- Build all packages: `go build ./...`
- Build MCP server: `go build -o build/mcp-gds cmd/mcp/main.go`
- Check for Go formatting issues: `go fmt ./...`
- Check for common Go issues: `go vet ./...`

## Module Management

- Module name: `mgds`
- Go version: 1.23.0

## MCP Server

- Main executable: `cmd/mcp/main.go` - MCP (Model Context Protocol) server
- Build and run: `go run cmd/mcp/main.go`
- For Claude Desktop: Use `build/mcp-gds` binary (automatically loads .env from multiple locations)
- Required services: Start dependencies with `docker-compose up -d` (Neo4j, MongoDB, Redis, SearXNG)

## Infrastructure Dependencies

**Required Services (via Docker Compose)**
- **Neo4j**: Graph database for storing nodes and relationships (ports: 7474, 7687)
- **MongoDB**: Document database for storing scraped content and analysis results (port: 27017)
- **Redis**: Cache and message queue backend (port: 6379)  
- **SearXNG**: Privacy-respecting web search engine (port: 8888)
- **Mongo Express**: MongoDB web admin interface (port: 8081)

**Environment Variables**
The server requires environment variables for database connections. Create a `.env` file with:
- `NEO4J_USERNAME`, `NEO4J_PASSWORD`: Neo4j authentication
- `MONGODB_USERNAME`, `MONGODB_PASSWORD`: MongoDB authentication  
- `REDIS_PASSWORD`: Redis authentication
- `MONGO_EXPRESS_USERNAME`, `MONGO_EXPRESS_PASSWORD`: Mongo Express web UI

**Start Infrastructure**: `docker-compose up -d`
**Stop Infrastructure**: `docker-compose down`

## Available MCP Tools

The server provides the following tools for Claude Desktop integration:

### Graph Exploration Tools
- `get_graph_overview`: Get knowledge graph statistics and node types
- `get_nodes_by_type`: Retrieve paginated nodes of a specific type
- `get_node_relations`: Get all relations for a specific node
- `get_connected_nodes`: Find nodes connected to a specific node

### Content Analysis Tools
- `search_and_analyze`: Perform deep web search with knowledge graph construction
- `analyze_webpage`: Analyze specific webpage content and build graph relationships

### Document Management Tools
- `get_document`: Intelligently retrieve documents with automatic analysis fallback

### Graph Maintenance Tools
- `delete_node`: Delete specific nodes from the knowledge graph
- `delete_node_cascade`: Delete nodes and all descendants with depth limit
- `delete_relation`: Delete specific relationships between nodes
- `preview_deletion`: Preview deletion operations before execution

## Architecture

WHEN THERE IS A CHANGE IN THE ARCHITECTURE, UPDATE THE CLAUDE.MD FILE.

This is a Go-based MCP server providing deep search capabilities through an LLM-explorable graph system. The application enables comprehensive information discovery by building knowledge graphs from web searches, content analysis, and document storage that Large Language Models can explore to provide detailed, contextual responses.

### Core Components

**Web Scraping (`internal/pkg/scrapper/`)**

- Main interface: `WebScraper` with implementations using Colly framework
- Supports comprehensive data extraction: HTML, text, links, images, forms, scripts, meta tags
- Configurable scraping options: timeouts, user agents, rate limiting, selective extraction
- Factory pattern for scraper instantiation
- BDD-style tests using Ginkgo and Gomega

**Caching System (`internal/pkg/cache/`)**

- **Note**: Cache system is documented but not currently implemented in the codebase

**Queue System (`internal/pkg/queue/`)**

- **Note**: Queue system is documented but not currently implemented in the codebase


**Graph Database (`internal/pkg/graph/`)**

- Interface for graph database operations with Neo4j implementation
- Typed nodes (URL, User) with validation for displayName and ID fields
- Separate concerns: Neo4j stores relationships, MongoDB stores data
- Methods: CreateNode, CreateRelation, GetNode, NodeExists, Close
- Connection pooling and automatic reconnection handling
- Factory pattern for graph instantiation with environment variable configuration

**HTML Serialization (`internal/pkg/serializer/`)**

- Interface for parsing and serializing HTML documents with goquery integration
- Document structure with title, body, head, links, images, meta tags, and plain text
- Configurable extraction options for links, images, meta tags, and text content
- Support for both string and io.Reader/Writer operations
- Factory pattern for serializer instantiation

**Database Abstraction (`internal/pkg/database/`)**

- Generic database interface supporting Insert, FindByID, Update, Delete operations
- MongoDB implementation with proper error handling (ErrNotFound, ErrInvalidID)
- Context-aware operations for timeout and cancellation support
- Factory pattern for database instantiation

**Search Engine (`internal/pkg/search_engine/`)**

- Interface for web search functionality with SearXNG implementation
- Configurable search options: time range, category, language, result limits
- Returns structured search results with URLs, titles, content
- Factory pattern for search engine instantiation with environment variable configuration

### Key Design Patterns

- **Interface-driven design**: All major components define interfaces first
- **Factory pattern**: Used for component instantiation (web scrapers, databases)
- **Context propagation**: All operations support context for cancellation/timeouts

### Dependencies

Key external dependencies:

- `github.com/metoro-io/mcp-golang`: MCP (Model Context Protocol) server implementation
- `github.com/gocolly/colly/v2`: Web scraping framework
- `github.com/onsi/ginkgo/v2` + `github.com/onsi/gomega`: BDD testing framework
- `github.com/neo4j/neo4j-go-driver/v5`: Neo4j database driver for graph operations
- `github.com/PuerkitoBio/goquery`: HTML document traversal and manipulation for serialization
- `go.mongodb.org/mongo-driver`: MongoDB driver for database operations
- `github.com/joho/godotenv`: Environment variable loading from .env files

### Testing Strategy

- BDD-style tests using Ginkgo's Describe/Context/It structure
- Gomega assertions for readable test expectations
- Mock HTTP servers for web scraping tests
- Test coverage includes timeout handling, error scenarios, and configuration options

### Application Services

**Graph Explorer Service (`internal/app/services/graph_explorer/`)**

- Service enabling LLM exploration of knowledge graphs with overview, node retrieval, and relationship queries
- Supports paginated node retrieval by type and comprehensive relationship exploration for deep search
- Integration with graph database for statistical analysis and intelligent connectivity mapping
- Provides LLMs with tools to understand and navigate complex information structures

**Link Service (`internal/app/services/link/`)**

- Service for creating relationships between nodes in the graph database
- Direct and queued implementation patterns for scalable link creation
- Request validation ensuring source and target nodes are different and valid
- Integration with graph database for relationship management

**Graph Pruner Service (`internal/app/services/graph_pruner/`)**

- Service for systematic cleanup and maintenance of graph database
- Provides node and relationship deletion capabilities with cascade operations
- Supports bulk deletion operations for efficient graph pruning
- Returns detailed statistics on pruning operations (deleted nodes and relationships)
- Integration with Neo4j for transactional deletion operations


**Note**: The webpage analysis service has been removed to simplify the architecture. The `analyze_webpage` use case now directly uses the web scraper and webpage.Build() for maximum simplicity.

### Use Cases (MCP Tools)

**Explore Graph Use Case (`internal/app/use-cases/explore_graph/`)**

- Provides MCP tools enabling LLM exploration of knowledge graphs: get_graph_overview, get_nodes_by_type, get_node_relations, get_connected_nodes
- Allows LLMs to understand graph structure, navigate relationships, and discover connected information
- Graph statistics, node type analysis, and relationship mapping for intelligent traversal
- Paginated results with configurable limits for efficient exploration of large knowledge graphs

**Search and Analyze Use Case (`internal/app/use-cases/search_and_analyze/`)**

- Provides search_and_analyze MCP tool for deep search with knowledge graph construction
- Integrates SearXNG search with webpage analysis, content extraction, and relationship mapping
- Builds explorable graph nodes and connections for LLM-driven discovery
- Configurable search parameters: time range, category, language, result limits
- Returns comprehensive analysis including found documents, extracted data, and graph relationships for LLM exploration

**Analyze Webpage Use Case (`internal/app/use-cases/analyze_webpage/`)**

- Comprehensive webpage analysis workflow combining scraping, content analysis, and graph storage
- Request validation and response structuring with detailed error reporting
- Integration of web scraping and graph relationship creation
- Handles relation creation errors gracefully with detailed error context
- Returns document ID, extracted data, and relation statistics

**Get Document Use Case (`internal/app/use-cases/get_document/`)**

- Intelligent document retrieval with automatic analysis fallback for better LLM guidance
- Provides get_document MCP tool that tries document retrieval first, then auto-analyzes if needed
- Extracts URLs from node structure (DisplayName or ID patterns) for automatic webpage analysis
- Returns comprehensive status information (found, analyzed_and_retrieved, no_document, error)
- Improves LLM experience by providing a single, intelligent tool for document access
- Configurable auto-analysis behavior with detailed action and status reporting

**Prune Graph Use Case (`internal/app/use-cases/prune_graph/`)**

- Provides prune_graph MCP tool for cleaning up knowledge graph by removing nodes and relationships
- Uses graph pruner service to systematically delete nodes and their associated relationships
- Supports bulk operations for efficient graph maintenance and cleanup
- Returns statistics on deleted nodes and relationships for transparency

### Directory Structure

- `cmd/mcp/`: MCP server executable with tool registration and MCP protocol handling
- `internal/app/`: Application-specific code
  - `object/webpage/`: Domain objects for webpage entities  
  - `services/`: High-level business services
    - `graph_explorer/`: Graph database exploration and analysis
    - `link/`: Graph relationship creation with validation
    - `graph_pruner/`: Graph cleanup and maintenance service
  - `use-cases/`: MCP tool implementations
    - `explore_graph/`: Graph exploration tools (overview, nodes by type, relations, connected nodes)
    - `search_and_analyze/`: Web search and content analysis tool
    - `analyze_webpage/`: Webpage analysis workflow
    - `get_document/`: Intelligent document retrieval with automatic analysis fallback
    - `prune_graph/`: Graph cleanup and maintenance tools
- `internal/pkg/`: Reusable internal packages
  - `scrapper/`: Web scraping functionality with Colly integration
  - `node/`: Node type definitions for graph database entities
  - `graph/`: Graph database interfaces and Neo4j implementation with typed nodes
  - `serializer/`: HTML parsing and serialization with goquery integration
  - `database/`: Generic database abstraction with MongoDB implementation
  - `search_engine/`: Web search functionality with SearXNG integration
  - `configuration/`: Configuration objects for external services (MongoDB, Neo4j, SearXNG)
- `internal/constant/`: Application constants and definitions
- `searxng/`: SearXNG search engine configuration and data
- `build/`: Build artifacts and infrastructure configuration
- `docker/`: Docker-related files and configurations
- `docker-compose.yml`: Infrastructure services (Neo4j, MongoDB, Redis, SearXNG)
