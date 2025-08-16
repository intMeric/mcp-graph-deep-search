package link

import (
	"mgds/internal/pkg/node"
)

type QueuedLinkRequest struct {
	SourceNode *node.Node `json:"source_node"`
	TargetNode *node.Node `json:"target_node"`
	LinkType   string     `json:"link_type"`
}
