package search_and_analyze

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"mgds/internal/app/services/link"
	"mgds/internal/app/use-cases/analyze_webpage"
	"mgds/internal/constant"
	"mgds/internal/pkg/database"
	"mgds/internal/pkg/graph"
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
	linkService := link.NewDirectLinkService(graphDB)

	// Initialize analyze webpage use case
	analyzeUseCase := analyze_webpage.NewAnalyzeWebpageUseCase(
		webScraper,
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

	// Initialize response with thread-safe access
	response := &SearchAndAnalyzeResponse{
		Query:              request.Query,
		SearchResultsFound: len(searchResponse.Results),
		URLsAnalyzed:       0,
		TotalRelations:     0,
		AnalysisResults:    make([]AnalysisResult, 0),
		AnalysisErrors:     make([]AnalysisError, 0),
	}

	// Analyze results in parallel with worker pool
	response, err = uc.analyzeResultsParallel(ctx, results, response)
	if err != nil {
		return nil, fmt.Errorf("parallel analysis failed: %w", err)
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

type workItem struct {
	index  int
	result search_engine.Result
}

type workResult struct {
	index    int
	success  bool
	analysis *AnalysisResult
	error    *AnalysisError
}

func (uc *searchAndAnalyzeUseCase) analyzeResultsParallel(ctx context.Context, results []search_engine.Result, response *SearchAndAnalyzeResponse) (*SearchAndAnalyzeResponse, error) {
	const maxWorkers = 4
	const analysisTimeout = 2 * time.Minute

	// Create context with timeout for the entire analysis
	analyzeCtx, cancel := context.WithTimeout(ctx, analysisTimeout)
	defer cancel()

	numResults := len(results)
	if numResults == 0 {
		return response, nil
	}

	// Calculate optimal number of workers
	numWorkers := maxWorkers
	if numResults < maxWorkers {
		numWorkers = numResults
	}

	// Create channels
	workChan := make(chan workItem, numResults)
	resultChan := make(chan workResult, numResults)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go uc.analysisWorker(analyzeCtx, &wg, workChan, resultChan)
	}

	// Send work items
	for i, result := range results {
		workChan <- workItem{index: i, result: result}
	}
	close(workChan)

	// Start goroutine to close result channel when all workers finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	analysisResults := make([]AnalysisResult, 0, numResults)
	analysisErrors := make([]AnalysisError, 0)
	urlsAnalyzed := 0
	totalRelations := 0

	for workRes := range resultChan {
		if workRes.success {
			analysisResults = append(analysisResults, *workRes.analysis)
			urlsAnalyzed++
			totalRelations += workRes.analysis.RelationsCreated

			log.Printf("[%d/%d] ✓ Analyzed %s: Document ID: %s, Relations: %d",
				workRes.index+1, numResults,
				workRes.analysis.URL,
				workRes.analysis.DocumentID,
				workRes.analysis.RelationsCreated,
			)
		} else {
			analysisErrors = append(analysisErrors, *workRes.error)
			log.Printf("[%d/%d] ✗ Failed to analyze %s: %v",
				workRes.index+1, numResults,
				workRes.error.URL,
				workRes.error.Error,
			)
		}
	}

	// Update response
	response.AnalysisResults = analysisResults
	response.AnalysisErrors = analysisErrors
	response.URLsAnalyzed = urlsAnalyzed
	response.TotalRelations = totalRelations

	return response, nil
}

func (uc *searchAndAnalyzeUseCase) analysisWorker(ctx context.Context, wg *sync.WaitGroup, workChan <-chan workItem, resultChan chan<- workResult) {
	defer wg.Done()

	for work := range workChan {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			resultChan <- workResult{
				index:   work.index,
				success: false,
				error: &AnalysisError{
					URL:   work.result.URL,
					Error: "analysis cancelled: " + ctx.Err().Error(),
				},
			}
			return
		default:
		}

		log.Printf("[%d] Starting analysis: %s", work.index+1, work.result.URL)

		analyzeRequest := &analyze_webpage.AnalyzeWebpageRequest{
			URL: work.result.URL,
		}

		// Create per-URL timeout context
		urlCtx, urlCancel := context.WithTimeout(ctx, 15*time.Second)
		
		analyzeResponse, err := uc.analyzeWebpageUseCase.Execute(urlCtx, analyzeRequest)
		urlCancel()

		if err != nil {
			resultChan <- workResult{
				index:   work.index,
				success: false,
				error: &AnalysisError{
					URL:   work.result.URL,
					Error: err.Error(),
				},
			}
		} else {
			resultChan <- workResult{
				index:   work.index,
				success: true,
				analysis: &AnalysisResult{
					URL:              work.result.URL,
					DocumentID:       analyzeResponse.DocumentID,
					RelationsCreated: analyzeResponse.RelationsCreated,
				},
			}
		}
	}
}
