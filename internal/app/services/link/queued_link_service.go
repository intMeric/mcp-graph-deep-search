package link

import (
	"context"

	"mgds/internal/pkg/node"
	"mgds/internal/pkg/queue"
)

type queuedLinkService struct {
	queue queue.Queue[*QueuedLinkRequest]
}

func NewQueuedLinkService(queue queue.Queue[*QueuedLinkRequest]) LinkService {
	return &queuedLinkService{
		queue: queue,
	}
}

func (s *queuedLinkService) CreateLink(ctx context.Context, sourceNode *node.Node, targetNode *node.Node, linkType string) (*LinkResponse, error) {
	request := &QueuedLinkRequest{
		SourceNode: sourceNode,
		TargetNode: targetNode,
		LinkType:   linkType,
	}

	if err := s.queue.PublishMessage(request); err != nil {
		return nil, err
	}

	return &LinkResponse{
		Success: true,
		Message: "Link request queued successfully",
	}, nil
}

func (s *queuedLinkService) Close() error {
	if s.queue != nil {
		return s.queue.Close()
	}
	return nil
}
