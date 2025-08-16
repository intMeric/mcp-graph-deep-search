package webpage_analysis

import (
	"context"

	"mgds/internal/app/object/webpage"
	"mgds/internal/app/services/text_analysis"
	"mgds/internal/pkg/scrapper"
)

type webpageAnalysisService struct {
	webScraper   scrapper.WebScraper
	textAnalyzer text_analysis.TextAnalysisService
}

func NewWebpageAnalysisService(webScraper scrapper.WebScraper, textAnalyzer text_analysis.TextAnalysisService) WebpageAnalysisService {
	return &webpageAnalysisService{
		webScraper:   webScraper,
		textAnalyzer: textAnalyzer,
	}
}

func (s *webpageAnalysisService) AnalyzeWebpage(ctx context.Context, urlStr string, options *scrapper.ScrapingOptions) (webpage.WebpageInterface, error) {
	scrapedData, err := s.webScraper.Scrape(ctx, urlStr, options)
	if err != nil {
		return nil, err
	}

	textAnalysis, err := s.textAnalyzer.AnalyzeText(ctx, scrapedData.Text)
	if err != nil {
		return nil, err
	}

	return webpage.Build(scrapedData, textAnalysis, urlStr), nil
}

func (s *webpageAnalysisService) Close() error {
	if s.webScraper != nil {
		if err := s.webScraper.Close(); err != nil {
			return err
		}
	}
	if s.textAnalyzer != nil {
		if err := s.textAnalyzer.Close(); err != nil {
			return err
		}
	}
	return nil
}
