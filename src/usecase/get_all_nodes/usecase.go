package getallnodes

import (
	"context"
	"time"

	"mgds/src/pkg/graph"
	"mgds/src/pkg/node"
)

// UseCase interface for the get all nodes use case
type UseCase interface {
	Execute(ctx context.Context) (*GetAllNodesResponse, error)
}

// useCase implementation of the use case
type useCase struct {
	graphDB graph.Interface
}

// NewUseCase creates a new instance of the use case
func NewUseCase(graphDB graph.Interface) UseCase {
	return &useCase{
		graphDB: graphDB,
	}
}

// Execute retrieves all nodes from the graph
func (u *useCase) Execute(ctx context.Context) (*GetAllNodesResponse, error) {
	startTime := time.Now()

	nodes, err := u.graphDB.GetAllNodes(ctx)
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
		summary := u.convertToNodeSummary(node)
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
func (u *useCase) convertToNodeSummary(nodeInterface node.Interface) NodeSummary {
	summary := NodeSummary{
		ID:          nodeInterface.GetMgdsId(),
		Type:        nodeInterface.GetType(),
		Description: nodeInterface.GetDescription(),
		DisplayName: nodeInterface.GetDisplayName(),
	}

	return summary
}