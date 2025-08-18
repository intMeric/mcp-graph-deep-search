package get_document_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/use-cases/analyze_webpage"
	"mgds/internal/app/use-cases/get_document"
	"mgds/internal/pkg/node"
)

// Mock database
type mockDatabase struct {
	shouldError bool
	document    map[string]interface{}
}

func (m *mockDatabase) FindByID(ctx context.Context, collection, id string, dest interface{}) error {
	if m.shouldError {
		return fmt.Errorf("database error")
	}
	if collection == "" {
		return fmt.Errorf("collection cannot be empty")
	}
	if result, ok := dest.(*map[string]interface{}); ok {
		*result = m.document
	}
	return nil
}

func (m *mockDatabase) Insert(ctx context.Context, id string, document interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockDatabase) Update(ctx context.Context, collection, id string, update interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *mockDatabase) Delete(ctx context.Context, collection, id string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockDatabase) Close(ctx context.Context) error {
	return nil
}

func (m *mockDatabase) GetLocation() string {
	return "mock_location"
}

// Mock analyze webpage use case
type mockAnalyzeWebpageUseCase struct {
	shouldError bool
	response    *analyze_webpage.AnalyzeWebpageResponse
}

func (m *mockAnalyzeWebpageUseCase) Execute(ctx context.Context, req *analyze_webpage.AnalyzeWebpageRequest) (*analyze_webpage.AnalyzeWebpageResponse, error) {
	if m.shouldError {
		return nil, fmt.Errorf("analyze webpage failed")
	}
	return m.response, nil
}

var _ = Describe("SmartDocumentRetrievalUseCase", func() {
	var (
		useCase               get_document.GetDocumentUseCase
		mockDB                *mockDatabase
		mockAnalyzeWeb        *mockAnalyzeWebpageUseCase
		ctx                   context.Context
		testNode              *node.Node
		testNodeWithLocation  *node.Node
		testNodeWithoutURL    *node.Node
	)

	BeforeEach(func() {
		mockDB = &mockDatabase{
			document: map[string]interface{}{
				"id":      "test-doc",
				"content": "test content",
				"title":   "Test Document",
			},
		}

		mockAnalyzeWeb = &mockAnalyzeWebpageUseCase{
			response: &analyze_webpage.AnalyzeWebpageResponse{
				URL:               "https://example.com/test",
				DocumentID:        "new-doc-id",
				ExtractedKeywords: []string{"test", "keyword"},
				RelationsCreated:  2,
				RelationErrors:    []analyze_webpage.RelationError{},
			},
		}

		useCase = get_document.NewGetDocumentUseCase(mockDB, mockAnalyzeWeb)
		ctx = context.Background()

		testNode = &node.Node{
			ID:          "example.com_/test",
			Type:        "URL",
			DisplayName: "https://example.com/test",
			Location:    "",
		}

		testNodeWithLocation = &node.Node{
			ID:          "example.com_/test",
			Type:        "URL",
			DisplayName: "Test Page",
			Location:    "webpage",
		}

		testNodeWithoutURL = &node.Node{
			ID:          "no-url-node",
			Type:        "URL",
			DisplayName: "No URL Node",
			Location:    "",
		}
	})

	Describe("Execute", func() {
		Context("with nil request", func() {
			It("should return error", func() {
				response, err := useCase.Execute(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request cannot be nil"))
				Expect(response).To(BeNil())
			})
		})

		Context("with nil node", func() {
			It("should return error", func() {
				req := &get_document.GetDocumentRequest{
					Node: nil,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("node cannot be nil"))
				Expect(response).To(BeNil())
			})
		})

		Context("with empty node ID", func() {
			It("should return error", func() {
				req := &get_document.GetDocumentRequest{
					Node: &node.Node{ID: ""},
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("node ID cannot be empty"))
				Expect(response).To(BeNil())
			})
		})

		Context("with existing document", func() {
			It("should retrieve existing document successfully", func() {
				req := &get_document.GetDocumentRequest{
					Node:        testNodeWithLocation,
					AutoAnalyze: false,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal("found"))
				Expect(response.Action).To(Equal("retrieved_existing_document"))
				Expect(response.Document).To(Equal(mockDB.document))
				Expect(response.Message).To(ContainSubstring("Document found and retrieved"))
			})

			It("should handle database error", func() {
				mockDB.shouldError = true
				req := &get_document.GetDocumentRequest{
					Node:        testNodeWithLocation,
					AutoAnalyze: false,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal("error"))
				Expect(response.Action).To(Equal("document_retrieval_failed"))
				Expect(response.Message).To(ContainSubstring("Failed to retrieve document"))
			})
		})

		Context("with no document and autoAnalyze=false", func() {
			It("should indicate analysis is required", func() {
				req := &get_document.GetDocumentRequest{
					Node:        testNode,
					AutoAnalyze: false,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal("no_document"))
				Expect(response.Action).To(Equal("analysis_required"))
				Expect(response.Message).To(ContainSubstring("webpage analysis required"))
			})
		})

		Context("with no document, autoAnalyze=true but no URL", func() {
			It("should indicate analysis is impossible", func() {
				req := &get_document.GetDocumentRequest{
					Node:        testNodeWithoutURL,
					AutoAnalyze: true,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal("no_document"))
				Expect(response.Action).To(Equal("analysis_impossible"))
				Expect(response.Message).To(ContainSubstring("no URL available"))
			})
		})

		Context("with autoAnalyze=true and URL available", func() {
			It("should perform auto-analysis and retrieve document", func() {
				req := &get_document.GetDocumentRequest{
					Node:        testNode,
					AutoAnalyze: true,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal("analyzed_and_retrieved"))
				Expect(response.Action).To(Equal("auto_analysis_successful"))
				Expect(response.Document).To(Equal(mockDB.document))
				Expect(response.AnalysisResult).NotTo(BeNil())
				Expect(response.AnalysisResult.DocumentID).To(Equal("new-doc-id"))
				Expect(response.AnalysisResult.ExtractedKeywords).To(ContainElement("test"))
			})

			It("should handle analysis failure", func() {
				mockAnalyzeWeb.shouldError = true
				req := &get_document.GetDocumentRequest{
					Node:        testNode,
					AutoAnalyze: true,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal("error"))
				Expect(response.Action).To(Equal("auto_analysis_failed"))
				Expect(response.Message).To(ContainSubstring("Auto-analysis failed"))
			})

			It("should handle analysis success but document retrieval failure", func() {
				mockDB.shouldError = true
				req := &get_document.GetDocumentRequest{
					Node:        testNode,
					AutoAnalyze: true,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
				Expect(response.Status).To(Equal("partial_success"))
				Expect(response.Action).To(Equal("analyzed_but_retrieval_failed"))
				Expect(response.AnalysisResult).NotTo(BeNil())
				Expect(response.Document).To(BeNil())
			})
		})

		Context("URL extraction from node", func() {
			It("should extract URL from DisplayName when it starts with http", func() {
				req := &get_document.GetDocumentRequest{
					Node:        testNode, // DisplayName: "https://example.com/test"
					AutoAnalyze: true,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal("analyzed_and_retrieved"))
			})

			It("should extract URL from ID when it contains domain info", func() {
				nodeWithDomainID := &node.Node{
					ID:          "github.com_/user/repo",
					Type:        "URL",
					DisplayName: "GitHub Repo",
					Location:    "",
				}

				req := &get_document.GetDocumentRequest{
					Node:        nodeWithDomainID,
					AutoAnalyze: true,
				}

				response, err := useCase.Execute(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(response.Status).To(Equal("analyzed_and_retrieved"))
			})
		})
	})
})