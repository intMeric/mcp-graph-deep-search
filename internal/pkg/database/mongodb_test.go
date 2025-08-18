package database_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/pkg/database"
)

type TestDocument struct {
	Name  string `bson:"name" json:"name"`
	Value int    `bson:"value" json:"value"`
}

var _ = Describe("Database", func() {
	var (
		db  database.Database
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("Factory", func() {
		Context("when creating a new database", func() {
			It("should create MongoDB with specific collection", func() {
				Skip("Requires MongoDB connection - integration test")
			})

			It("should require collection parameter", func() {
				Skip("Requires MongoDB connection - integration test")
			})
		})
	})

	Describe("MongoDB Operations", func() {
		Context("when MongoDB is available", func() {
			BeforeEach(func() {
				Skip("Integration tests require MongoDB connection")
				var err error
				db, err = database.NewMongoDatabase("test_collection")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				if db != nil {
					db.Close(ctx)
				}
			})

			Describe("Insert", func() {
				It("should insert a document successfully", func() {
					doc := &TestDocument{
						Name:  "test",
						Value: 42,
					}

					err := db.Insert(ctx, "test_collection", "test-id", doc)

					Expect(err).NotTo(HaveOccurred())
				})

				It("should handle context timeout", func() {
					timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
					defer cancel()

					doc := &TestDocument{Name: "test", Value: 42}
					err := db.Insert(timeoutCtx, "test_collection", "timeout-test", doc)

					Expect(err).To(HaveOccurred())
				})
			})

			Describe("FindByID", func() {
				var insertedID string

				BeforeEach(func() {
					doc := &TestDocument{Name: "find_test", Value: 100}
					insertedID = "find-test-id"
					err := db.Insert(ctx, "test_collection", insertedID, doc)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should find an existing document", func() {
					var result TestDocument
					err := db.FindByID(ctx, "test_collection", insertedID, &result)

					Expect(err).NotTo(HaveOccurred())
					Expect(result.Name).To(Equal("find_test"))
					Expect(result.Value).To(Equal(100))
				})

				It("should return ErrNotFound for non-existent document", func() {
					var result TestDocument
					err := db.FindByID(ctx, "test_collection", "non-existent-id", &result)

					Expect(err).To(Equal(database.ErrNotFound))
				})
			})

			Describe("Update", func() {
				var insertedID string

				BeforeEach(func() {
					doc := &TestDocument{Name: "update_test", Value: 200}
					insertedID = "update-test-id"
					err := db.Insert(ctx, "test_collection", insertedID, doc)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should update an existing document", func() {
					update := map[string]any{
						"name":  "updated_test",
						"value": 300,
					}

					err := db.Update(ctx, "test_collection", insertedID, update)
					Expect(err).NotTo(HaveOccurred())

					var result TestDocument
					err = db.FindByID(ctx, "test_collection", insertedID, &result)
					Expect(err).NotTo(HaveOccurred())
					Expect(result.Name).To(Equal("updated_test"))
					Expect(result.Value).To(Equal(300))
				})

				It("should return ErrNotFound for non-existent document", func() {
					update := map[string]any{"name": "test"}
					err := db.Update(ctx, "test_collection", "non-existent-id", update)

					Expect(err).To(Equal(database.ErrNotFound))
				})
			})

			Describe("Delete", func() {
				var insertedID string

				BeforeEach(func() {
					doc := &TestDocument{Name: "delete_test", Value: 400}
					insertedID = "delete-test-id"
					err := db.Insert(ctx, "test_collection", insertedID, doc)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should delete an existing document", func() {
					err := db.Delete(ctx, "test_collection", insertedID)
					Expect(err).NotTo(HaveOccurred())

					var result TestDocument
					err = db.FindByID(ctx, "test_collection", insertedID, &result)
					Expect(err).To(Equal(database.ErrNotFound))
				})

				It("should return ErrNotFound for non-existent document", func() {
					err := db.Delete(ctx, "test_collection", "non-existent-id")

					Expect(err).To(Equal(database.ErrNotFound))
				})
			})

			Describe("Close", func() {
				It("should close the database connection", func() {
					err := db.Close(ctx)

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("Configuration", func() {
		Context("when validating MongoConfig", func() {
			It("should reject nil config", func() {
				_, err := database.NewMongoDB(nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("config cannot be nil"))
			})

			It("should set default timeout when not specified", func() {
				config := &database.MongoConfig{
					URI:      "mongodb://localhost:27017",
					Database: "test",
				}

				Skip("Requires MongoDB connection")
				db, err := database.NewMongoDB(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(db).NotTo(BeNil())
			})
		})
	})
})
