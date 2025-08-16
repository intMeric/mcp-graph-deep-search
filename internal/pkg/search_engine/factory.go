package search_engine

import (
	"fmt"

	"mgds/internal/pkg/configuration"
)

func NewSearchEngine(engineType SearchEngineType) (SearchEngine, error) {
	switch engineType {
	case SearXNGType:
		config := configuration.NewSearXNGConfig()
		return NewSearXNG(&SearXNGConfig{
			Host:     config.Host,
			Language: config.Language,
		})
	default:
		return nil, fmt.Errorf("unsupported search engine type: %s", engineType)
	}
}
