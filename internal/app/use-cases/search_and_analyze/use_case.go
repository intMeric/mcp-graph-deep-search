package search_and_analyze

import (
	"context"
	"fmt"
	"log"

	"mgds/internal/app/services/link"
	"mgds/internal/app/services/text_analysis"
	"mgds/internal/app/services/webpage_analysis"
	"mgds/internal/app/use-cases/analyze_webpage"
	"mgds/internal/constant"
	"mgds/internal/pkg/database"
	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/keyword"
	"mgds/internal/pkg/scrapper"
	"mgds/internal/pkg/search_engine"
)

type searchAndAnalyzeUseCase struct {
	analyzeWebpageUseCase analyze_webpage.AnalyzeWebpageUseCase
}

func NewSearchAndAnalyzeUseCase() (SearchAndAnalyzeUseCase, error) {
	// Initialize database
	db, err := database.NewMongoDatabase(constant.WebPageLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize graph
	graphDB, err := graph.NewGraph(graph.Neo4jGraphType)
	if err != nil {
		db.Close(context.Background())
		return nil, fmt.Errorf("failed to initialize graph: %w", err)
	}

	// Initialize services
	webScraper := scrapper.NewWebScraper()

	keywordExtractor, err := keyword.NewExtractor()
	if err != nil {
		db.Close(context.Background())
		graphDB.Close(context.Background())
		return nil, fmt.Errorf("failed to initialize keyword extractor: %w", err)
	}

	textAnalyzer := text_analysis.NewTextAnalysisService(keywordExtractor)
	webpageAnalyzer := webpage_analysis.NewWebpageAnalysisService(webScraper, textAnalyzer)
	linkService := link.NewDirectLinkService(graphDB)

	// Initialize analyze webpage use case
	analyzeUseCase := analyze_webpage.NewAnalyzeWebpageUseCase(
		webpageAnalyzer,
		linkService,
		db,
	)

	return &searchAndAnalyzeUseCase{
		analyzeWebpageUseCase: analyzeUseCase,
	}, nil
}

func (uc *searchAndAnalyzeUseCase) Execute(ctx context.Context, request *SearchAndAnalyzeRequest) (*SearchAndAnalyzeResponse, error) {
	if err := uc.validateRequest(request); err != nil {
		return nil, err
	}

	// Initialize search engine
	searchEngine, err := search_engine.NewSearchEngine(search_engine.SearXNGType)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize search engine: %w", err)
	}

	// Configure search engine
	searchEngine = searchEngine.SetCategory(request.Category).SetLanguage(request.Language)
	if request.TimeRange != "" {
		searchEngine = searchEngine.SetTimeRange(request.TimeRange)
	}

	// Perform search
	log.Printf("Searching for: %s", request.Query)
	searchResponse, err := searchEngine.Search(ctx, request.Query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	log.Printf("Found %d search results", len(searchResponse.Results))

	// Limit results if specified
	results := searchResponse.Results
	if request.MaxResults > 0 && len(results) > request.MaxResults {
		results = results[:request.MaxResults]
		log.Printf("Analyzing first %d results", request.MaxResults)
	}

	// Initialize response
	response := &SearchAndAnalyzeResponse{
		Query:              request.Query,
		SearchResultsFound: len(searchResponse.Results),
		URLsAnalyzed:       0,
		TotalRelations:     0,
		AnalysisResults:    make([]AnalysisResult, 0),
		AnalysisErrors:     make([]AnalysisError, 0),
	}

	// Analyze each search result
	for i, result := range results {
		log.Printf("[%d/%d] Analyzing: %s", i+1, len(results), result.URL)

		analyzeRequest := &analyze_webpage.AnalyzeWebpageRequest{
			URL: result.URL,
		}

		analyzeResponse, err := uc.analyzeWebpageUseCase.Execute(ctx, analyzeRequest)
		if err != nil {
			log.Printf("Failed to analyze %s: %v", result.URL, err)
			response.AnalysisErrors = append(response.AnalysisErrors, AnalysisError{
				URL:   result.URL,
				Error: err.Error(),
			})
			continue
		}

		response.URLsAnalyzed++
		response.TotalRelations += analyzeResponse.RelationsCreated

		response.AnalysisResults = append(response.AnalysisResults, AnalysisResult{
			URL:               result.URL,
			DocumentID:        analyzeResponse.DocumentID,
			ExtractedKeywords: analyzeResponse.ExtractedKeywords,
			RelationsCreated:  analyzeResponse.RelationsCreated,
		})

		log.Printf("✓ Analyzed %s: Document ID: %s, Keywords: %d, Relations: %d",
			result.URL,
			analyzeResponse.DocumentID,
			len(analyzeResponse.ExtractedKeywords),
			analyzeResponse.RelationsCreated,
		)

		if len(analyzeResponse.RelationErrors) > 0 {
			log.Printf("  Relation errors: %d", len(analyzeResponse.RelationErrors))
			for _, relErr := range analyzeResponse.RelationErrors {
				log.Printf("    %s: %s", relErr.Type, relErr.Error)
			}
		}
	}

	return response, nil
}

func (uc *searchAndAnalyzeUseCase) validateRequest(request *SearchAndAnalyzeRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.Query == "" {
		return fmt.Errorf("query is required")
	}

	if request.Category == "" {
		request.Category = "general"
	}

	if request.Language == "" {
		request.Language = "en"
	}

	if request.MaxResults <= 0 {
		request.MaxResults = 10
	}

	return nil
}
