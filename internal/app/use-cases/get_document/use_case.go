package get_document

import (
	"context"
	"fmt"
	"strings"

	"mgds/internal/app/use-cases/analyze_webpage"
	"mgds/internal/pkg/database"
	"mgds/internal/pkg/node"
)

type getDocumentUseCase struct {
	database              database.Database
	analyzeWebpageUseCase analyze_webpage.AnalyzeWebpageUseCase
}

func NewGetDocumentUseCase(
	db database.Database,
	analyzeWebUC analyze_webpage.AnalyzeWebpageUseCase,
) GetDocumentUseCase {
	return &getDocumentUseCase{
		database:              db,
		analyzeWebpageUseCase: analyzeWebUC,
	}
}

func (uc *getDocumentUseCase) Execute(ctx context.Context, req *GetDocumentRequest) (*GetDocumentResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.Node == nil {
		return nil, fmt.Errorf("node cannot be nil")
	}

	if strings.TrimSpace(req.Node.ID) == "" {
		return nil, fmt.Errorf("node ID cannot be empty")
	}

	response := &GetDocumentResponse{
		Node: req.Node,
	}

	// Check if document exists
	if req.Node.IsDocumentAvailable {
		// Try to retrieve existing document directly from database
		var document map[string]interface{}
		err := uc.database.FindByID(ctx, req.Node.Type, req.Node.ID, &document)
		if err == nil {
			response.Document = document
			response.Status = "found"
			response.Action = "retrieved_existing_document"
			response.Message = "Document found and retrieved from database"
			return response, nil
		}
		
		// Document should be available but retrieval failed
		response.Status = "error"
		response.Action = "document_retrieval_failed"
		response.Message = fmt.Sprintf("Failed to retrieve document: %v", err)
		return response, nil
	}

	// No document exists yet
	if !req.AutoAnalyze {
		response.Status = "no_document"
		response.Action = "analysis_required"
		response.Message = "No document found. Set autoAnalyze=true or use analyze_webpage tool first."
		return response, nil
	}

	// Check if we can auto-analyze (need URL from node)
	url := extractURLFromNode(req.Node)
	if url == "" {
		response.Status = "no_document"
		response.Action = "analysis_impossible"
		response.Message = "No document found and cannot auto-analyze: no URL available in node. Please analyze the webpage first using analyze_webpage tool."
		return response, nil
	}

	// Perform automatic analysis
	analyzeReq := &analyze_webpage.AnalyzeWebpageRequest{
		URL: url,
	}

	analyzeResp, err := uc.analyzeWebpageUseCase.Execute(ctx, analyzeReq)
	if err != nil {
		response.Status = "error"
		response.Action = "auto_analysis_failed"
		response.Message = fmt.Sprintf("Auto-analysis failed: %v", err)
		return response, nil
	}

	// Analysis successful - now retrieve the document
	// After analysis, the document should be available in the database using node type as collection
	updatedNode := &node.Node{
		ID:                  req.Node.ID,
		Type:                req.Node.Type,
		DisplayName:         req.Node.DisplayName,
		IsDocumentAvailable: true,
	}

	var document map[string]interface{}
	err = uc.database.FindByID(ctx, updatedNode.Type, updatedNode.ID, &document)
	if err != nil {
		response.Status = "partial_success"
		response.Action = "analyzed_but_retrieval_failed"
		response.Message = "Webpage analyzed successfully but document retrieval failed"
		response.AnalysisResult = &AnalysisResult{
			DocumentID:       analyzeResp.DocumentID,
			CreatedRelations: analyzeResp.RelationsCreated,
			Errors:           extractErrorMessages(analyzeResp.RelationErrors),
		}
		return response, nil
	}

	// Complete success
	response.Document = document
	response.Status = "analyzed_and_retrieved"
	response.Action = "auto_analysis_successful"
	response.Message = "Webpage analyzed and document retrieved successfully"
	response.AnalysisResult = &AnalysisResult{
		DocumentID:       analyzeResp.DocumentID,
		CreatedRelations: analyzeResp.RelationsCreated,
		Errors:           extractErrorMessages(analyzeResp.RelationErrors),
	}

	// Update the node in response with new location
	response.Node = updatedNode

	return response, nil
}

func extractURLFromNode(n *node.Node) string {
	// For URL type nodes, try different strategies to extract URL
	if n.Type == "URL" {
		// Strategy 1: Use DisplayName if it looks like a URL
		if strings.HasPrefix(n.DisplayName, "http://") || strings.HasPrefix(n.DisplayName, "https://") {
			return n.DisplayName
		}

		// Strategy 2: Try to construct URL from ID if it contains domain info
		if strings.Contains(n.ID, "_") {
			// URL nodes typically have IDs like "domain.com_/path"
			parts := strings.SplitN(n.ID, "_", 2)
			if len(parts) == 2 {
				domain := parts[0]
				path := parts[1]
				// Assume HTTPS by default
				return fmt.Sprintf("https://%s%s", domain, path)
			}
		}
	}

	return ""
}

func extractErrorMessages(relationErrors []analyze_webpage.RelationError) []string {
	var errors []string
	for _, relErr := range relationErrors {
		errors = append(errors, fmt.Sprintf("%s relation error: %s", relErr.Type, relErr.Error))
	}
	return errors
}