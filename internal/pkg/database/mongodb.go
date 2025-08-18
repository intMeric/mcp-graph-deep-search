package database

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	URI               string
	Database          string
	Username          string
	Password          string
	Timeout           time.Duration
	DefaultCollection string
}

type mongoDatabase struct {
	client            *mongo.Client
	database          *mongo.Database
	timeout           time.Duration
	defaultCollection string
}

func NewMongoDB(config *MongoConfig) (Database, error) {
	if config == nil {
		return nil, fmt.Errorf("MongoDB config cannot be nil")
	}

	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	if config.DefaultCollection == "" {
		config.DefaultCollection = "documents"
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)

	if config.Username != "" && config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &mongoDatabase{
		client:            client,
		database:          client.Database(config.Database),
		timeout:           config.Timeout,
		defaultCollection: config.DefaultCollection,
	}, nil
}

// Implement the Database interface methods
func (m *mongoDatabase) Insert(ctx context.Context, collection, id string, document any) error {
	coll := m.database.Collection(collection)

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	objectID := m.stringToObjectID(id)

	docWithID := map[string]any{
		"_id": objectID,
	}

	if docMap, ok := document.(map[string]any); ok {
		for k, v := range docMap {
			if k != "_id" {
				docWithID[k] = v
			}
		}
	} else {
		docWithID["data"] = document
	}

	_, err := coll.InsertOne(ctx, docWithID)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

func (m *mongoDatabase) stringToObjectID(id string) primitive.ObjectID {
	hash := sha256.Sum256([]byte(id))
	var objectIDBytes [12]byte
	copy(objectIDBytes[:], hash[:12])
	return primitive.ObjectID(objectIDBytes)
}

func (m *mongoDatabase) FindByID(ctx context.Context, collection, id string, dest any) error {
	objectID := m.stringToObjectID(id)

	coll := m.database.Collection(collection)

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	err := coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(dest)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		return fmt.Errorf("failed to find document: %w", err)
	}

	return nil
}

func (m *mongoDatabase) Update(ctx context.Context, collection, id string, update any) error {
	objectID := m.stringToObjectID(id)

	coll := m.database.Collection(collection)

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	result, err := coll.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (m *mongoDatabase) Delete(ctx context.Context, collection, id string) error {
	objectID := m.stringToObjectID(id)

	coll := m.database.Collection(collection)

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	result, err := coll.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (m *mongoDatabase) Close(ctx context.Context) error {
	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}
