package configuration

import (
	"time"

	"mgds/internal/pkg/env"
)

// MongoDB configuration constants
const (
	DefaultMongoDBURI        = "mongodb://localhost:27017"
	DefaultMongoDBDatabase   = "mmm-osint"
	DefaultMongoDBTimeout    = 10 * time.Second
	DefaultMongoDBCollection = "documents"
)

// MongoDBConfig represents MongoDB connection configuration
type MongoDBConfig struct {
	URI               string
	Database          string
	Username          string
	Password          string
	Timeout           time.Duration
	DefaultCollection string
}

// NewMongoDBConfig creates a new MongoDB configuration from environment variables
func NewMongoDBConfig(defaultCollection string) *MongoDBConfig {
	return &MongoDBConfig{
		URI:               env.GetOrDefault("MONGODB_URI", DefaultMongoDBURI),
		Database:          env.GetOrDefault("MONGODB_DATABASE", DefaultMongoDBDatabase),
		Username:          env.GetOrDefault("MONGODB_USERNAME", "admin"),
		Password:          env.GetOrDefault("MONGODB_PASSWORD", "admin"),
		DefaultCollection: defaultCollection,
		Timeout:           DefaultMongoDBTimeout,
	}
}
