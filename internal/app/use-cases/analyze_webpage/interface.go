package analyze_webpage

import (
	"context"
)

type AnalyzeWebpageRequest struct {
	URL string `json:"url"`
}

type AnalyzeWebpageResponse struct {
	URL               string          `json:"url"`
	DocumentID        string          `json:"document_id"`
	ExtractedPII      map[string]any  `json:"extracted_pii"`
	ExtractedKeywords []string        `json:"extracted_keywords"`
	RelationsCreated  int             `json:"relations_created"`
	RelationErrors    []RelationError `json:"relation_errors,omitempty"`
}

type RelationError struct {
	Type       string `json:"type"`        // "PII" ou "LINK"
	Source     string `json:"source"`      // ID du noeud source
	Target     string `json:"target"`      // ID du noeud target
	TargetName string `json:"target_name"` // Nom descriptif du target
	Error      string `json:"error"`       // Message d'erreur
}

type AnalyzeWebpageUseCase interface {
	Execute(ctx context.Context, request *AnalyzeWebpageRequest) (*AnalyzeWebpageResponse, error)
}
