package search_and_analyze

import (
	"context"
)

type SearchAndAnalyzeRequest struct {
	Query      string `json:"query"`
	TimeRange  string `json:"time_range,omitempty"`
	Category   string `json:"category"`
	Language   string `json:"language"`
	MaxResults int    `json:"max_results"`
}

type SearchAndAnalyzeResponse struct {
	Query              string           `json:"query"`
	SearchResultsFound int              `json:"search_results_found"`
	URLsAnalyzed       int              `json:"urls_analyzed"`
	TotalRelations     int              `json:"total_relations"`
	AnalysisResults    []AnalysisResult `json:"analysis_results,omitempty"`
	AnalysisErrors     []AnalysisError  `json:"analysis_errors,omitempty"`
}

type AnalysisResult struct {
	URL              string `json:"url"`
	DocumentID       string `json:"document_id"`
	RelationsCreated int    `json:"relations_created"`
}

type AnalysisError struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type SearchAndAnalyzeUseCase interface {
	Execute(ctx context.Context, request *SearchAndAnalyzeRequest) (*SearchAndAnalyzeResponse, error)
}
