package database

import (
	"mgds/internal/pkg/configuration"
)

// NewMongoDatabase creates a new MongoDB database with a specific collection
func NewMongoDatabase(collection string) (Database, error) {
	config := configuration.NewMongoDBConfig(collection)
	return NewMongoDB(&MongoConfig{
		URI:               config.URI,
		Database:          config.Database,
		Username:          config.Username,
		Password:          config.Password,
		DefaultCollection: config.DefaultCollection,
		Timeout:           config.Timeout,
	})
}
