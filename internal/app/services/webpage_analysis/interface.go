package webpage_analysis

import (
	"context"

	"mgds/internal/app/object/webpage"
	"mgds/internal/pkg/scrapper"
)

type WebpageAnalysisService interface {
	AnalyzeWebpage(ctx context.Context, urlStr string, options *scrapper.ScrapingOptions) (webpage.WebpageInterface, error)
	Close() error
}
