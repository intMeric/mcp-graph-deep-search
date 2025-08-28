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
- NEVER write empty if blocks or placeholder code that does nothing - delete it completely

## Testing

- TEST-DRIVEN DEVELOPMENT IS NON-NEGOTIABLE.
- Run all tests: `go test -v ./...`
- Run specific package tests: `go test -v ./src/pkg/graph`
- Run tests with coverage: `go test -v -cover ./...`
- Run tests for specific test: `go test -v ./src/pkg/object -run TestURLNode`
- Tests use Ginkgo BDD framework with Gomega assertions
- Run manual testing: `go run testing/main.go` (tests SearchXNG integration)
- Start SearchXNG instance: See searXNG/ directory for configuration (expects localhost:8888)

### Test Example Structure

All tests must follow the Ginkgo BDD pattern with Gomega assertions:

```go
package mypackage_test

import (
    "context"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "mgds/src/pkg/mypackage"
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
- Build specific package: `go build ./src/pkg/graph`
- Build MCP server: `go build -o build/mcp-gds cmd/src-mcp/main.go`
- Check for Go formatting issues: `go fmt ./...`
- Check for common Go issues: `go vet ./...`
- Run manual testing: `go run testing/main.go`

## Module Management

- Module name: `mgds`
- Go version: 1.23.0

## Application Architecture

The application is a graph-based knowledge mapping system designed for LLM deep search capabilities via MCP (Model Context Protocol). The system combines web search, scraping, and graph storage to create a knowledge network accessible by LLMs.

### Core Components

- **MCP Server**: Exposes graph operations via MCP tools (`get_all_nodes`, `get_node_by_id`, `web_search`)
- **Graph Database Layer**: Uses Cayley graph database with memory/LevelDB backends for persistent storage
- **Node System**: Interface-based design with URLNode as the primary node type representing web pages  
- **Search Engine Layer**: Integration with SearchXNG for web search capabilities
- **Scraping Engine**: Web scraping functionality using Colly with anti-detection measures
- **Web Search Service**: High-level service that orchestrates search, scraping, and graph indexing
- **Use Cases**: Clean architecture with dedicated use cases for graph operations

### Package Structure

- `cmd/src-mcp/`: MCP server implementation exposing graph operations to LLMs
- `src/pkg/graph/`: Graph database interface and Cayley implementation
  - `Interface`: Defines graph operations (CRUD, traversal, queries)  
  - `CayleyGraph`: Cayley-based implementation with memory/LevelDB storage
  - `Relation`: Represents directed relationships between nodes
- `src/pkg/node/`: Generic node interface for graph entities
- `src/pkg/object/`: Concrete node implementations
  - `URLNode`: Web page nodes with URL, title, content, scraping status
- `src/pkg/search_engine/`: SearchXNG integration
  - `Interface`: Search engine abstraction layer
  - `SearXNG`: SearchXNG client implementation with categories, time ranges, filtering
- `src/pkg/scrapper/`: Web scraping functionality
  - `Interface`: Scraping abstraction layer
  - `Colly`: Colly-based scraper with anti-detection, rate limiting, robots.txt support
- `src/service/web_seach/`: High-level web search and indexing service
  - `Interface`: Service interface for search and indexing operations
  - Implementation orchestrates search, scraping, and graph storage
- `src/usecase/`: Clean architecture use cases
  - `get_all_nodes/`: Retrieve all nodes from graph (interface: `UseCase`, implementation: `useCase`)
  - `get_node_by_id/`: Retrieve specific node by ID (interface: `UseCase`, implementation: `useCase`)
  - `web_search/`: Execute web search and indexing workflow (interface: `UseCase`, implementation: `useCase`)
- `searXNG/`: SearchXNG configuration for local search instance
- `testing/`: Manual testing utilities demonstrating SearchXNG integration

### Key Design Patterns

- **Clean Architecture**: Separation of use cases, services, and infrastructure layers
- **Interface-driven architecture**: All major components expose interfaces for testability and flexibility
- **MCP Integration**: Server exposes tools for LLM interaction via Model Context Protocol
- **Context-aware operations**: All operations use context.Context for cancellation and timeouts
- **Configuration-driven**: Extensive configuration support for search engines, scrapers, and services
- **Generic property system**: Nodes support arbitrary key-value properties for extensibility
- **Anti-detection measures**: Scrapers include user agent rotation, delays, and robots.txt compliance

### External Dependencies

- **SearchXNG**: Local search aggregator instance (configured to run on localhost:8888)
  - Configuration file: `searXNG/settings.yml`
  - Supports multiple search engines (Google, Bing, DuckDuckGo, etc.)
  - Categories: general, news, images, videos, etc.
  - Time ranges: short, medium, long
- **Cayley Graph Database**: Graph database with memory/LevelDB backends
- **Colly**: Web scraping framework with anti-detection capabilities
- **MCP-Golang**: Model Context Protocol implementation for LLM integration

### SearchXNG Setup

The system requires a local SearchXNG instance running on `http://localhost:8888`. The configuration file in `searXNG/settings.yml` provides:
- Server configuration (port 8888, bind address 127.0.0.1)
- Search engine integrations
- Request timeout and rate limiting settings
- Categories and filters configuration

### MCP Tools Available

1. **get_all_nodes**: Retrieves all nodes from the graph database
2. **get_node_by_id**: Retrieves a specific node by its ID
3. **web_search**: Searches the web, scrapes results, and indexes them in the graph

### Database Operations

The graph supports standard CRUD operations, relationship management, graph traversal (connected nodes, pathfinding), and node queries (by type, properties). All operations are context-aware and interface-based.

### Environment Configuration

- Copy `.env.template` to `.env` for environment configuration
- Configure SearchXNG endpoint via `SEARXNG_HOST` (default: http://localhost:8888)
- Neo4j configuration available but not currently used (system uses Cayley)