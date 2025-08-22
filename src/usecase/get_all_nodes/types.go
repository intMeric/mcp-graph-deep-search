package getallnodes

import "time"

// GetAllNodesResponse represents the response from getting all nodes
type GetAllNodesResponse struct {
	Status        ResponseStatus `json:"status"`
	Nodes         []NodeSummary  `json:"nodes"`
	TotalNodes    int            `json:"totalNodes"`
	ExecutionTime time.Duration  `json:"executionTime"`
	Error         string         `json:"error,omitempty"`
}

// NodeSummary represents a simplified view of a node
type NodeSummary struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	DisplayName string `json:"displayName,omitempty"`
}

// ResponseStatus represents the status of the operation
type ResponseStatus string

const (
	StatusSuccess ResponseStatus = "success"
	StatusFailed  ResponseStatus = "failed"
)
