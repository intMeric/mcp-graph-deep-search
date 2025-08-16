#!/bin/bash

set -e

echo "🔍 Graph Deep Search (GDS) Launcher"
echo "=================================="

# Check if .env file exists in current directory (build)
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found in current directory"
    echo "Please create a .env file with the required environment variables"
    exit 1
fi

# Load environment variables from .env file
echo "📄 Loading environment variables from .env..."
export $(grep -v '^#' .env | grep -v '^$' | xargs)

# Print all environment variables from .env (excluding comments and empty lines)
echo "🔑 Environment variables loaded:"
while IFS= read -r line; do
    # Skip comments and empty lines
    if [[ ! "$line" =~ ^[[:space:]]*# ]] && [[ ! "$line" =~ ^[[:space:]]*$ ]]; then
        key=$(echo "$line" | cut -d'=' -f1)
        echo "  ✓ $key"
    fi
done < .env

echo ""

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker Compose not found. Please install Docker and Docker Compose"
    exit 1
fi

# Start infrastructure services using docker-compose from current directory
echo "🐳 Starting infrastructure services (Neo4j, MongoDB, Redis, SearXNG)..."
if command -v docker-compose &> /dev/null; then
    docker-compose up -d
else
    docker compose up -d
fi

# Wait a moment for services to start
echo "⏳ Waiting for services to initialize..."
sleep 5

# Check if the MCP server binary exists
if [ ! -f "mcp-gds" ]; then
    echo "❌ Error: MCP server binary not found at mcp-gds"
    echo "Please run: go build -o build/mcp-gds cmd/mcp/main.go"
    exit 1
fi

# Make sure the binary is executable
chmod +x mcp-gds

# Start the MCP Deep Search server
echo "🚀 Starting MCP Graph Deep Search server..."
echo "Press Ctrl+C to stop the server"
echo ""

./mcp-gds