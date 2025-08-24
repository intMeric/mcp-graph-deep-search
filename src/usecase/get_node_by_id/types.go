package getnodebyid

// GetNodeByIdRequest represents the request to get a node by ID
type GetNodeByIdRequest struct {
	NodeID string `json:"nodeId"`
}

// GetNodeByIdResponse represents the response from getting a node by ID
type GetNodeByIdResponse struct {
	Status   ResponseStatus `json:"status"`
	NodeData map[string]any `json:"nodeData,omitempty"`
	Error    string         `json:"error,omitempty"`
}

// ResponseStatus represents the status of the operation
type ResponseStatus string

const (
	StatusSuccess  ResponseStatus = "success"
	StatusFailed   ResponseStatus = "failed"
	StatusNotFound ResponseStatus = "not_found"
)