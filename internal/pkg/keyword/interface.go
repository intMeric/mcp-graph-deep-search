package keyword

import "context"

type Extractor interface {
	ExtractKeywords(ctx context.Context, text string, options *Options) ([]string, error)
	ExtractKeywordsWithScores(ctx context.Context, text string, options *Options) ([]Keyword, error)
	Close() error
}

type Keyword struct {
	Text      string  `json:"text"`
	Frequency int     `json:"frequency"`
	Score     float64 `json:"score"`
}

type Options struct {
	MinLength       int  `json:"min_length"`
	MaxKeywords     int  `json:"max_keywords"`
	RemoveStopWords bool `json:"remove_stop_words"`
}

func DefaultOptions() *Options {
	return &Options{
		MinLength:       3,
		MaxKeywords:     20,
		RemoveStopWords: true,
	}
}
