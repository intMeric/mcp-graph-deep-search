package analyze_webpage

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"mgds/internal/app/object/webpage"
	"mgds/internal/app/services/link"
	"mgds/internal/constant"
	"mgds/internal/pkg/database"
	"mgds/internal/pkg/node"
	"mgds/internal/pkg/scrapper"
)

type analyzeWebpageUseCase struct {
	webScraper  scrapper.WebScraper
	linkService link.LinkService
	database    database.Database
}

func NewAnalyzeWebpageUseCase(
	webScraper scrapper.WebScraper,
	linkService link.LinkService,
	database database.Database,
) AnalyzeWebpageUseCase {
	return &analyzeWebpageUseCase{
		webScraper:  webScraper,
		linkService: linkService,
		database:    database,
	}
}

func (uc *analyzeWebpageUseCase) Execute(ctx context.Context, request *AnalyzeWebpageRequest) (*AnalyzeWebpageResponse, error) {
	if err := uc.validateRequest(request); err != nil {
		return nil, err
	}

	// Scrape the webpage directly
	scrapedData, err := uc.webScraper.Scrape(ctx, request.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape webpage: %w", err)
	}

	// Build webpage object focused on content extraction only
	webpageObj := webpage.Build(scrapedData, request.URL)

	documentID := webpageObj.GetID()

	webpageNode := webpageObj.ToNode()
	
	err = uc.storeDocument(ctx, webpageNode.Type, documentID, webpageObj)
	if err != nil {
		return nil, err
	}

	webpageNode.IsDocumentAvailable = true

	linkNodes := webpageObj.GetLinkNodes()

	linkRelationsCreated, linkRelationErrors := uc.createLinkRelations(ctx, webpageNode, linkNodes)

	totalRelationsCreated := linkRelationsCreated
	allRelationErrors := linkRelationErrors

	return &AnalyzeWebpageResponse{
		URL:              request.URL,
		DocumentID:       documentID,
		Title:            webpageObj.GetTitle(),
		Text:             webpageObj.GetText(),
		RelationsCreated: totalRelationsCreated,
		RelationErrors:   allRelationErrors,
	}, nil
}

func (uc *analyzeWebpageUseCase) validateRequest(request *AnalyzeWebpageRequest) error {
	if request == nil || request.URL == "" {
		return fmt.Errorf("invalid request: URL is required")
	}

	_, err := url.Parse(request.URL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	return nil
}

func (uc *analyzeWebpageUseCase) storeDocument(ctx context.Context, collection, documentID string, webpageObj webpage.WebpageInterface) error {
	document := webpageObj.ToDocument()

	err := uc.database.Insert(ctx, collection, documentID, document)
	if err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	return nil
}


func (uc *analyzeWebpageUseCase) createLinkRelations(ctx context.Context, sourceNode *node.Node, linkNodes []*node.Node) (int, []RelationError) {
	relationsCreated := 0
	var relationErrors []RelationError

	for _, targetNode := range linkNodes {
		if targetNode.ID == sourceNode.ID {
			continue
		}

		_, err := uc.linkService.CreateLink(ctx, sourceNode, targetNode, constant.NavigationLink)
		if err != nil {
			relationError := RelationError{
				Type:       "LINK",
				Source:     sourceNode.ID,
				Target:     targetNode.ID,
				TargetName: targetNode.DisplayName,
				Error:      err.Error(),
			}
			relationErrors = append(relationErrors, relationError)
			log.Printf("Failed to create link relation from %s to %s (%s): %v",
				sourceNode.ID, targetNode.ID, targetNode.DisplayName, err)
		} else {
			relationsCreated++
		}
	}

	return relationsCreated, relationErrors
}

