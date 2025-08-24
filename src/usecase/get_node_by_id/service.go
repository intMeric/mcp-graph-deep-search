package getnodebyid

import (
	"context"
	"strings"

	"mgds/src/pkg/graph"
)

// Service interface for the get node by ID use case
type Service interface {
	Execute(ctx context.Context, request *GetNodeByIdRequest) (*GetNodeByIdResponse, error)
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

// Execute retrieves a node by ID and deserializes it
func (s *service) Execute(ctx context.Context, request *GetNodeByIdRequest) (*GetNodeByIdResponse, error) {
	if request == nil || strings.TrimSpace(request.NodeID) == "" {
		return &GetNodeByIdResponse{
			Status: StatusFailed,
			Error:  "invalid node ID",
		}, nil
	}

	nodeExists, err := s.graphDB.NodeExists(ctx, request.NodeID)
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

	node, err := s.graphDB.GetNode(ctx, request.NodeID)
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