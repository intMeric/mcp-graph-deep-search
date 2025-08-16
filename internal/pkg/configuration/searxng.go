package configuration

import (
	"mgds/internal/pkg/env"
)

// SearXNG configuration constants
const (
	DefaultSearXNGHost     = "http://localhost:8888"
	DefaultSearXNGLanguage = "en"
	DefaultSearXNGCategory = "general"
)

// SearXNGConfig represents SearXNG search engine configuration
type SearXNGConfig struct {
	Host     string
	Language string
	Category string
}

// NewSearXNGConfig creates a new SearXNG configuration from environment variables
func NewSearXNGConfig() *SearXNGConfig {
	return &SearXNGConfig{
		Host:     env.GetOrDefault("SEARXNG_HOST", DefaultSearXNGHost),
		Language: env.GetOrDefault("SEARXNG_LANGUAGE", DefaultSearXNGLanguage),
		Category: env.GetOrDefault("SEARXNG_CATEGORY", DefaultSearXNGCategory),
	}
}
