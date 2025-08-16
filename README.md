# MCP Graph Deep Search

A Model Context Protocol (MCP) server that provides deep search capabilities through an LLM-explorable graph system. This tool enables comprehensive information discovery by building knowledge graphs from web searches, content analysis, and document storage.

## Quick Start

### Prerequisites

- Go 1.23.0 or later
- Docker and Docker Compose
- Claude Desktop

### 1. Setup Infrastructure

Start the required services (Neo4j, MongoDB, SearXNG):

```bash
cd build/
docker-compose up -d
```

This will start:

- **Neo4j**: Graph database (ports 7474, 7687)
- **MongoDB**: Document storage (port 27017)
- **SearXNG**: Privacy-focused search engine (port 8888)

### 2. Configure Environment

Create a `.env` file in the `build/` directory with the following variables:

```env
# MongoDB Configuration
MONGODB_URI=mongodb://admin:admin@localhost:27017
MONGODB_USERNAME=admin
MONGODB_PASSWORD=admin

# Neo4j Configuration  
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=isMyNeo4jPassword

# SearXNG Configuration
SEARXNG_HOST=http://localhost:8888
SEARXNG_LANGUAGE=en
```

### 3. Build the MCP Server

```bash
go build -o build/mcp-gds cmd/mcp/main.go
```

### 4. Configure Claude Desktop

Add the following configuration to your Claude Desktop MCP servers configuration:

```json
{
  "mcpServers": {
    "graph-deep-search": {
      "command": "/home/vieu/Documents/Dev/mcp-graph-deep-search/build/mcp-gds",
      "args": ["/home/vieu/Documents/Dev/mcp-graph-deep-search/build/.env"],
      "env": {}
    }
  }
}
```

**Note**: Update the paths to match your actual installation directory.

## Available Tools

Once configured, Claude Desktop will have access to these MCP tools:

### `search_and_analyze`

Performs deep web search with knowledge graph construction:

- Searches the web using SearXNG
- Analyzes webpage content for PII and keywords
- Builds explorable graph nodes and relationships
- Parameters: query, time_range, category, language, num_results

### `analyze_webpage`

Analyzes a specific webpage:

- Scrapes webpage content
- Extracts text, links, images, and metadata
- Performs PII and keyword analysis
- Creates graph nodes and relationships
- Parameters: url

### `get_graph_overview`

Provides statistical overview of the knowledge graph:

- Total nodes and relationships count
- Node type distribution
- Graph connectivity metrics

### `get_nodes_by_type`

Retrieves nodes filtered by type:

- Supports pagination
- Available types: URL, User, etc.
- Parameters: node_type, page, limit

### `get_node_relations`

Explores relationships for a specific node:

- Shows incoming and outgoing relationships
- Relationship type analysis
- Parameters: node_id

### `get_connected_nodes`

Finds nodes connected to a specific node:

- Multi-hop traversal capabilities
- Relationship filtering
- Parameters: node_id, relationship_types, max_depth

### `search_document`

Retrieves detailed content from stored documents:

- Searches MongoDB for document content
- Full-text document access
- Parameters: node_id, location

## Usage Examples

### Basic Web Search and Analysis

```
Please search for "artificial intelligence trends 2024" and analyze the results
```

### Webpage Analysis

```
Analyze the webpage https://example.com for me
```

### Graph Exploration

```
Show me an overview of the knowledge graph, then explore connections from the most interesting node
```

### Document Retrieval

```
Get the full content of document with ID "url_example_com_ai_article"
```

## Infrastructure Management

### Start Services

```bash
cd build/
docker-compose up -d
```

### Stop Services

```bash
cd build/
docker-compose down
```

### View Service Status

```bash
cd build/
docker-compose ps
```

### Access Web Interfaces

- **Neo4j Browser**: http://localhost:7474
- **SearXNG**: http://localhost:8888

## Configuration

### Environment Variables

All configuration is handled through environment variables in the `.env` file:

- `NEO4J_URI`, `NEO4J_USERNAME`, `NEO4J_PASSWORD`: Neo4j connection and authentication
- `MONGODB_URI`, `MONGODB_USERNAME`, `MONGODB_PASSWORD`: MongoDB connection and authentication  
- `SEARXNG_HOST`, `SEARXNG_LANGUAGE`: SearXNG search engine configuration

### SearXNG Configuration

SearXNG settings can be customized by editing `build/searxng/config/settings.yml`.

## Troubleshooting

### Common Issues

**MCP Server Won't Start**

- Check that all required services are running: `docker-compose ps`
- Verify environment variables in `.env` file
- Check file permissions on the `mcp-gds` binary

**Database Connection Issues**

- Ensure Neo4j and MongoDB are healthy: `docker-compose logs`
- Verify credentials in `.env` file match Docker Compose environment

**Search Not Working**

- Check SearXNG is accessible: http://localhost:8888
- Review SearXNG logs: `docker-compose logs searxng`

### Logs and Debugging

- View MCP server logs in Claude Desktop's developer console
- Check Docker service logs: `docker-compose logs [service_name]`

## Development

### Testing

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -cover ./...

# Run specific package tests
go test -v ./internal/pkg/cache
```

### Building

```bash
# Build all packages
go build ./...

# Build MCP server
go build -o build/mcp-gds cmd/mcp/main.go
```

### Code Quality

```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...
```
