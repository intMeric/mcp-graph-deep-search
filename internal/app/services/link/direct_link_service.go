package link

import (
	"context"

	"mgds/internal/pkg/graph"
	"mgds/internal/pkg/node"
)

type directLinkService struct {
	graph graph.Graph
}

func NewDirectLinkService(graph graph.Graph) LinkService {
	return &directLinkService{
		graph: graph,
	}
}

func (s *directLinkService) CreateLink(ctx context.Context, sourceNode *node.Node, targetNode *node.Node, linkType string) (*LinkResponse, error) {
	response := &LinkResponse{
		Success: true,
	}

	sourceExists, err := s.graph.NodeExists(ctx, sourceNode.ID)
	if err != nil {
		return nil, err
	}
	if !sourceExists {
		if err := s.graph.CreateNode(ctx, sourceNode); err != nil {
			return nil, err
		}
		response.SourceCreated = true
	}

	targetExists, err := s.graph.NodeExists(ctx, targetNode.ID)
	if err != nil {
		return nil, err
	}
	if !targetExists {
		if err := s.graph.CreateNode(ctx, targetNode); err != nil {
			return nil, err
		}
		response.TargetCreated = true
	}

	relation, err := s.graph.CreateLink(ctx, sourceNode, targetNode, linkType)
	if err != nil {
		return nil, err
	}

	response.Message = "Link created successfully"
	response.Relation = relation

	return response, nil
}

func (s *directLinkService) Close() error {
	if s.graph != nil {
		return s.graph.Close(context.Background())
	}
	return nil
}
