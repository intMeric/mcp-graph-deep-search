package graph

import (
	"fmt"
	"strings"

	"mgds/internal/pkg/node"
)

type Relation struct {
	Type     string `json:"type"`
	SourceID string `json:"sourceId"`
	TargetID string `json:"targetId"`
}

func (r *Relation) Validate() error {
	if strings.TrimSpace(r.Type) == "" {
		return fmt.Errorf("relation type cannot be empty")
	}

	if strings.TrimSpace(r.SourceID) == "" {
		return fmt.Errorf("sourceId cannot be empty")
	}

	if strings.TrimSpace(r.TargetID) == "" {
		return fmt.Errorf("targetId cannot be empty")
	}

	if r.SourceID == r.TargetID {
		return fmt.Errorf("sourceId and targetId cannot be the same")
	}

	return nil
}

type GraphStats struct {
	TotalNodes      int            `json:"totalNodes"`
	NodesByType     map[string]int `json:"nodesByType"`
	TotalRelations  int            `json:"totalRelations"`
	RelationsByType map[string]int `json:"relationsByType"`
}

type PaginatedNodesResult struct {
	Nodes  []*node.Node `json:"nodes"`
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
}

type NodeConnection struct {
	Node     *node.Node `json:"node"`
	Relation *Relation  `json:"relation"`
	IsSource bool       `json:"isSource"`
}
