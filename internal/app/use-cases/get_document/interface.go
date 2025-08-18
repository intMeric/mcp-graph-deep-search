package get_document

import (
	"context"
	
	"mgds/internal/pkg/node"
)

type GetDocumentUseCase interface {
	Execute(ctx context.Context, req *GetDocumentRequest) (*GetDocumentResponse, error)
}

type GetDocumentRequest struct {
	Node        *node.Node `json:"node"`
	AutoAnalyze bool       `json:"autoAnalyze"`
}

type GetDocumentResponse struct {
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