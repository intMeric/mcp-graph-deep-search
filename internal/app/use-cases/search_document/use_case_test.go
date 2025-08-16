package search_document_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/use-cases/search_document"
	"mgds/internal/pkg/database"
	"mgds/internal/pkg/node"
)

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	documents map[string]map[string]interface{} // collection -> id -> document
	shouldErr bool
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		documents: make(map[string]map[string]interface{}),
	}
}

func (m *MockDatabase) AddDocument(collection, id string, doc interface{}) {
	if m.documents[collection] == nil {
		m.documents[collection] = make(map[string]interface{})
	}
	m.documents[collection][id] = doc
}

func (m *MockDatabase) Insert(ctx context.Context, id string, document interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *MockDatabase) FindByID(ctx context.Context, collection, id string, dest interface{}) error {
	if m.shouldErr {
		return fmt.Errorf("database error")
	}

	if m.documents[collection] == nil {
		return database.ErrNotFound
	}

	doc, exists := m.documents[collection][id]
	if !exists {
		return database.ErrNotFound
	}

	// Simple mock assignment - in real tests this would be more sophisticated
	if destMap, ok := dest.(*map[string]interface{}); ok {
		*destMap = doc.(map[string]interface{})
	}

	return nil
}

func (m *MockDatabase) Update(ctx context.Context, collection, id string, update interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *MockDatabase) Delete(ctx context.Context, collection, id string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockDatabase) Close(ctx context.Context) error {
	return nil
}

func (m *MockDatabase) GetLocation() string {
	return "mock_location"
}

var _ = Describe("SearchDocumentUseCase", func() {
	var (
		useCase  search_document.SearchDocumentUseCase
		mockDB   *MockDatabase
		ctx      context.Context
		testNode *node.Node
		testDoc  map[string]interface{}
	)

	BeforeEach(func() {
		mockDB = NewMockDatabase()
		useCase = search_document.NewSearchDocumentUseCase(mockDB)
		ctx = context.Background()

		testNode = &node.Node{
			ID:          "test-id",
			Type:        "webpage",
			DisplayName: "Test Page",
			Location:    "test_collection",
		}

		testDoc = map[string]interface{}{
			"title":   "Test Document",
			"url":     "https://example.com",
			"content": "This is test content",
		}
	})

	AfterEach(func() {
		if mockDB != nil {
			mockDB.Close(ctx)
		}
	})

	Describe("Execute", func() {
		Context("with valid node having location", func() {
			It("should find and return document", func() {
				// Arrange
				mockDB.AddDocument("test_collection", "test-id", testDoc)

				// Act
				result, err := useCase.Execute(ctx, testNode)

				// Assert
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				resultMap := result.(map[string]interface{})
				Expect(resultMap["title"]).To(Equal("Test Document"))
				Expect(resultMap["url"]).To(Equal("https://example.com"))
			})
		})

		Context("with node without location", func() {
			It("should return error", func() {
				// Arrange
				testNode.Location = ""

				// Act
				result, err := useCase.Execute(ctx, testNode)

				// Assert
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("location"))
				Expect(result).To(BeNil())
			})

			It("should return error for nil node", func() {
				// Act
				result, err := useCase.Execute(ctx, nil)

				// Assert
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("node"))
				Expect(result).To(BeNil())
			})
		})

		Context("with node not found in database", func() {
			It("should return ErrNotFound", func() {
				// Arrange - no document added to mock DB

				// Act
				result, err := useCase.Execute(ctx, testNode)

				// Assert
				Expect(err).To(Equal(database.ErrNotFound))
				Expect(result).To(BeNil())
			})
		})

		Context("with database errors", func() {
			It("should handle database connection errors", func() {
				// Arrange
				mockDB.shouldErr = true

				// Act
				result, err := useCase.Execute(ctx, testNode)

				// Assert
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database error"))
				Expect(result).To(BeNil())
			})
		})
	})
})
