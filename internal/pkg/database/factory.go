package database

import (
	"mgds/internal/pkg/configuration"
)

// NewMongoDatabase creates a new MongoDB database connection
func NewMongoDatabase(databaseName string) (Database, error) {
	config := configuration.NewMongoDBConfig(databaseName)
	return NewMongoDB(&MongoConfig{
		URI:               config.URI,
		Database:          config.Database,
		Username:          config.Username,
		Password:          config.Password,
		DefaultCollection: config.DefaultCollection,
		Timeout:           config.Timeout,
	})
}
