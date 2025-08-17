package smart_document_retrieval

import (
	"context"
	
	"mgds/internal/pkg/node"
)

type SmartDocumentRetrievalUseCase interface {
	Execute(ctx context.Context, req *SmartDocumentRetrievalRequest) (*SmartDocumentRetrievalResponse, error)
}

type SmartDocumentRetrievalRequest struct {
	Node        *node.Node `json:"node"`
	AutoAnalyze bool       `json:"autoAnalyze"`
}

type SmartDocumentRetrievalResponse struct {
	Node           *node.Node             `json:"node"`
	Document       map[string]interface{} `json:"document,omitempty"`
	Status         string                 `json:"status"`
	Action         string                 `json:"action"`
	Message        string                 `json:"message"`
	AnalysisResult *AnalysisResult        `json:"analysisResult,omitempty"`
}

type AnalysisResult struct {
	DocumentID        string                 `json:"documentId"`
	ExtractedKeywords []string               `json:"extractedKeywords"`
	ExtractedPII      []string               `json:"extractedPII"`
	CreatedRelations  int                    `json:"createdRelations"`
	Errors            []string               `json:"errors,omitempty"`
}