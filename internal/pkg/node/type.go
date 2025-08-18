package node

import (
	"fmt"
	"strings"
)

type Node struct {
	Type                 string `json:"type"`
	DisplayName          string `json:"displayName"`
	ID                   string `json:"id"`
	Score                int    `json:"score,omitempty"`
	SubType              string `json:"subType,omitempty"`
	IsDocumentAvailable  bool   `json:"isDocumentAvailable,omitempty"`
}

type NodeConvertible interface {
	ToNode() *Node
	GetId() string
}

func (n *Node) ToNode() *Node {
	return n
}

func (n *Node) GetId() string {
	return n.ID
}

func (n *Node) Validate() error {
	if strings.TrimSpace(n.Type) == "" {
		return fmt.Errorf("type cannot be empty")
	}

	if strings.TrimSpace(n.DisplayName) == "" {
		return fmt.Errorf("displayName cannot be empty")
	}

	if strings.TrimSpace(n.ID) == "" {
		return fmt.Errorf("id cannot be empty")
	}

	return nil
}
