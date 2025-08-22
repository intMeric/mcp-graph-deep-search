package websearch

import (
	"time"
)

type SearchRequest struct {
	Query      string `json:"query" validate:"required"`
	MaxResults int    `json:"maxResults" validate:"min=1,max=100"`
}

type SearchResponse struct {
	Query         string        `json:"query"`
	Status        string        `json:"status"` // "success", "failed"
	Nodes         []NodeSummary `json:"nodes"`
	ExecutionTime time.Duration `json:"executionTime"`
	Error         string        `json:"error,omitempty"`
}

type NodeSummary struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
)
