package search_engine

import (
	"context"
	"errors"
)

var (
	ErrInvalidCategory   = errors.New("invalid category")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrUrlIsInvalid      = errors.New("invalid URL")
	ErrInvalidRecurrence = errors.New("invalid recurrence value")
)

type SearchEngine interface {
	Search(ctx context.Context, query string) (*SearchResponse, error)
	SetTimeRange(timeRange string) SearchEngine
	SetCategory(category string) SearchEngine
	SetLanguage(lang string) SearchEngine
}