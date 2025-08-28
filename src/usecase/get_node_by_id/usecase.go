package getnodebyid

import (
	"context"
	"strings"

	"mgds/src/pkg/graph"
)

// UseCase interface for the get node by ID use case
type UseCase interface {
	Execute(ctx context.Context, request *GetNodeByIdRequest) (*GetNodeByIdResponse, error)
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

// Execute retrieves a node by ID and deserializes it
func (u *useCase) Execute(ctx context.Context, request *GetNodeByIdRequest) (*GetNodeByIdResponse, error) {
	if request == nil || strings.TrimSpace(request.NodeID) == "" {
		return &GetNodeByIdResponse{
			Status: StatusFailed,
			Error:  "invalid node ID",
		}, nil
	}

	nodeExists, err := u.graphDB.NodeExists(ctx, request.NodeID)
	if err != nil {
		return &GetNodeByIdResponse{
			Status: StatusFailed,
			Error:  err.Error(),
		}, nil
	}

	if !nodeExists {
		return &GetNodeByIdResponse{
			Status: StatusNotFound,
			Error:  "node not found",
		}, nil
	}

	node, err := u.graphDB.GetNode(ctx, request.NodeID)
	if err != nil {
		return &GetNodeByIdResponse{
			Status: StatusFailed,
			Error:  err.Error(),
		}, nil
	}

	if node == nil {
		return &GetNodeByIdResponse{
			Status: StatusNotFound,
			Error:  "node not found",
		}, nil
	}

	nodeData, err := node.Serialize()
	if err != nil {
		return &GetNodeByIdResponse{
			Status: StatusFailed,
			Error:  err.Error(),
		}, nil
	}

	return &GetNodeByIdResponse{
		Status:   StatusSuccess,
		NodeData: nodeData,
	}, nil
}