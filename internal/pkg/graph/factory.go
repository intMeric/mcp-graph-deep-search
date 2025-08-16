package graph

import (
	"fmt"
	"mgds/internal/pkg/configuration"
)

type GraphType string

const (
	Neo4jGraphType GraphType = "neo4j"
)

func NewGraph(graphType GraphType) (Graph, error) {
	switch graphType {
	case Neo4jGraphType:
		config := configuration.NewNeo4jConfig()
		return NewNeo4jGraph(&Neo4jConfig{
			URI:      config.URI,
			Username: config.Username,
			Password: config.Password,
			Database: config.Database,
		})
	default:
		return nil, fmt.Errorf("unsupported graph type: %s", graphType)
	}
}
