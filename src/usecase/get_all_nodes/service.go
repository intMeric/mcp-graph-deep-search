package getallnodes

import (
	"context"
	"time"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/node"
)

// Service interface for the get all nodes use case
type Service interface {
	Execute(ctx context.Context) (*GetAllNodesResponse, error)
}

// service implementation of the use case
type service struct {
	graphDB graph.Interface
}

// NewService creates a new instance of the use case service
func NewService(graphDB graph.Interface) Service {
	return &service{
		graphDB: graphDB,
	}
}

// Execute retrieves all nodes from the graph
func (s *service) Execute(ctx context.Context) (*GetAllNodesResponse, error) {
	startTime := time.Now()

	nodes, err := s.graphDB.GetAllNodes(ctx)
	if err != nil {
		return &GetAllNodesResponse{
			Status:        StatusFailed,
			Nodes:         nil,
			TotalNodes:    0,
			ExecutionTime: time.Since(startTime),
			Error:         err.Error(),
		}, nil
	}

	// Convert nodes to NodeSummary
	var nodeSummaries []NodeSummary
	for _, node := range nodes {
		summary := s.convertToNodeSummary(node)
		nodeSummaries = append(nodeSummaries, summary)
	}

	return &GetAllNodesResponse{
		Status:        StatusSuccess,
		Nodes:         nodeSummaries,
		TotalNodes:    len(nodeSummaries),
		ExecutionTime: time.Since(startTime),
	}, nil
}

// convertToNodeSummary converts a node.Interface to NodeSummary
func (s *service) convertToNodeSummary(nodeInterface node.Interface) NodeSummary {
	summary := NodeSummary{
		ID:          nodeInterface.GetMgdsId(),
		Type:        nodeInterface.GetType(),
		Description: nodeInterface.GetDescription(),
		DisplayName: nodeInterface.GetDisplayName(),
	}

	return summary
}
