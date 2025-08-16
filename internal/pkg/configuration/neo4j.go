package configuration

import (
	"mgds/internal/pkg/env"
)

// Neo4j configuration constants
const (
	DefaultNeo4jURI      = "bolt://localhost:7687"
	DefaultNeo4jUsername = "neo4j"
	DefaultNeo4jPassword = "isMyNeo4jPassword"
	DefaultNeo4jDatabase = "neo4j"
)

// Neo4jConfig represents Neo4j graph database configuration
type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

// NewNeo4jConfig creates a new Neo4j configuration from environment variables
func NewNeo4jConfig() *Neo4jConfig {
	return &Neo4jConfig{
		URI:      env.GetOrDefault("NEO4J_URI", DefaultNeo4jURI),
		Username: env.GetOrDefault("NEO4J_USERNAME", DefaultNeo4jUsername),
		Password: env.GetOrDefault("NEO4J_PASSWORD", DefaultNeo4jPassword),
		Database: env.GetOrDefault("NEO4J_DATABASE", DefaultNeo4jDatabase),
	}
}
