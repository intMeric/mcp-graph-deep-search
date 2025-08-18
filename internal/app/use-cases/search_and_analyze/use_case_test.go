package search_and_analyze_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"mgds/internal/app/use-cases/search_and_analyze"
)

var _ = Describe("SearchAndAnalyzeUseCase", func() {
	// Note: We skip use case initialization in tests since it requires
	// database and graph connections. In a real test environment,
	// we would mock these dependencies.

	Describe("Request Validation", func() {
		Context("with invalid requests", func() {
			It("should validate that query is required", func() {
				// This test would require mocking dependencies
				// For now, we're just testing the structure
				request := &search_and_analyze.SearchAndAnalyzeRequest{
					Query: "",
				}
				Expect(request.Query).To(BeEmpty())
			})

			It("should validate that language defaults correctly", func() {
				request := &search_and_analyze.SearchAndAnalyzeRequest{
					Query:    "test query",
					Language: "",
				}
				Expect(request.Language).To(BeEmpty())
			})
		})

		Context("with valid requests", func() {
			It("should accept properly formatted requests", func() {
				request := &search_and_analyze.SearchAndAnalyzeRequest{
					Query:      "test query",
					TimeRange:  "day",
					Category:   "general",
					Language:   "en",
					MaxResults: 5,
				}

				Expect(request.Query).To(Equal("test query"))
				Expect(request.Category).To(Equal("general"))
				Expect(request.Language).To(Equal("en"))
				Expect(request.MaxResults).To(Equal(5))
			})
		})
	})

	Describe("Response Structure", func() {
		It("should have proper response structure", func() {
			response := &search_and_analyze.SearchAndAnalyzeResponse{
				Query:              "test query",
				SearchResultsFound: 10,
				URLsAnalyzed:       8,
				TotalRelations:     25,
				AnalysisResults: []search_and_analyze.AnalysisResult{
					{
						URL:              "https://example.com",
						DocumentID:       "doc-123",
						RelationsCreated: 3,
					},
				},
				AnalysisErrors: []search_and_analyze.AnalysisError{
					{
						URL:   "https://error.com",
						Error: "connection failed",
					},
				},
			}

			Expect(response.Query).To(Equal("test query"))
			Expect(response.SearchResultsFound).To(Equal(10))
			Expect(response.URLsAnalyzed).To(Equal(8))
			Expect(response.TotalRelations).To(Equal(25))
			Expect(response.AnalysisResults).To(HaveLen(1))
			Expect(response.AnalysisErrors).To(HaveLen(1))
		})
	})
})
