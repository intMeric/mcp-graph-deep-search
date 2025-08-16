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

- Generic `Cache` interface supporting Set/Get/Delete/Exists operations
- Implementations: LRU cache and Redis cache
- Context-aware operations with expiration support

**Queue System (`internal/pkg/queue/`)**

- Generic queue interface `Queue[T]` for type-safe message handling
- Request-Response pattern with `RequestResponseQueue[T, R]`
- Redis-based implementation for distributed queuing
- Support for both fire-and-forget and request-reply messaging patterns

**Keyword Extraction (`internal/pkg/keyword/`)**

- Simple interface for extracting keywords from text
- Prose library integration for natural language processing
- Configurable options: minimum word length, maximum keywords, stop word filtering
- Two extraction methods: simple string list or keywords with frequency scores

**PII Extraction (`internal/pkg/pii/`)**

- Simple interface for extracting Personally Identifiable Information from text
- Regex-based detection of emails, phones, credit cards, SSNs, IP addresses, IBANs
- Built on intMeric/pii-extractor library
- Returns structured results with entity types, values, counts, and contexts

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
- **Factory pattern**: Used for component instantiation (web scrapers, queues)
- **Generic types**: Queue system uses Go generics for type safety
- **Context propagation**: All operations support context for cancellation/timeouts

### Dependencies

Key external dependencies:

- `github.com/metoro-io/mcp-golang`: MCP (Model Context Protocol) server implementation
- `github.com/gocolly/colly/v2`: Web scraping framework
- `github.com/redis/go-redis/v9`: Redis client for caching and queuing
- `github.com/hashicorp/golang-lru/v2`: LRU cache implementation
- `github.com/jdkato/prose/v2`: Natural language processing for keyword extraction
- `github.com/intMeric/pii-extractor`: PII detection and extraction
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

**Text Analysis Service (`internal/app/services/text_analysis/`)**

- Composite service combining PII extraction and keyword extraction
- Single interface for analyzing text content with structured results
- Automatic HTML parsing and text extraction using serializer package
- Methods to check for presence of PII or keywords in results
- JSON serialization support for results
- Built on top of existing PII and keyword extraction packages

**Webpage Analysis Service (`internal/app/services/webpage_analysis/`)**

- High-level service for complete webpage analysis workflow
- Integrates web scraping, text analysis, and node generation
- Generates structured webpage and link nodes for graph database
- Consistent URL-based node ID generation using domain and path
- Returns comprehensive analysis results including scraped data, text analysis, and nodes

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

**Search Document Use Case (`internal/app/use-cases/search_document/`)**

- Provides search_document MCP tool for retrieving detailed content from knowledge graph nodes
- Searches MongoDB for documents using node ID and location, enabling deep content access
- Returns structured document data that LLMs can analyze for comprehensive responses

**Analyze Webpage Use Case (`internal/app/use-cases/analyze_webpage/`)**

- Comprehensive webpage analysis workflow combining scraping, content analysis, and graph storage
- Request validation and response structuring with detailed error reporting
- Integration of web scraping, PII extraction, keyword extraction, and graph relationship creation
- Handles relation creation errors gracefully with detailed error context
- Returns document ID, extracted data, and relation statistics

### Directory Structure

- `cmd/mcp/`: MCP server executable with tool registration and MCP protocol handling
- `internal/app/`: Application-specific code
  - `object/webpage/`: Domain objects for webpage entities  
  - `services/`: High-level business services
    - `graph_explorer/`: Graph database exploration and analysis
    - `link/`: Graph relationship creation with validation
    - `text_analysis/`: Composite text analysis combining PII and keywords with HTML parsing
    - `webpage_analysis/`: Complete webpage analysis service with node generation
  - `use-cases/`: MCP tool implementations
    - `explore_graph/`: Graph exploration tools (overview, nodes by type, relations, connected nodes)
    - `search_and_analyze/`: Web search and content analysis tool
    - `search_document/`: Document retrieval from database
    - `analyze_webpage/`: Webpage analysis workflow
- `internal/pkg/`: Reusable internal packages
  - `scrapper/`: Web scraping functionality with Colly integration
  - `cache/`: Generic caching interfaces and implementations  
  - `queue/`: Message queue interfaces and Redis implementations
  - `keyword/`: Keyword extraction from text using prose library
  - `pii/`: PII extraction using intMeric/pii-extractor library
  - `node/`: Node type definitions for graph database entities
  - `env/`: Environment configuration utilities
  - `graph/`: Graph database interfaces and Neo4j implementation with typed nodes
  - `serializer/`: HTML parsing and serialization with goquery integration
  - `database/`: Generic database abstraction with MongoDB implementation
  - `search_engine/`: Web search functionality with SearXNG integration
  - `configuration/`: Configuration objects for external services (MongoDB, Neo4j, SearXNG)
- `searxng/`: SearXNG search engine configuration and data
- `docker-compose.yml`: Infrastructure services (Neo4j, MongoDB, Redis, SearXNG)
