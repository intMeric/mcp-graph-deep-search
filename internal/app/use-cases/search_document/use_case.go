package search_document

import (
	"context"
	"fmt"
	"strings"

	"mgds/internal/pkg/database"
	"mgds/internal/pkg/node"
)

type searchDocumentUseCase struct {
	database database.Database
}

func NewSearchDocumentUseCase(db database.Database) SearchDocumentUseCase {
	return &searchDocumentUseCase{
		database: db,
	}
}

func (uc *searchDocumentUseCase) Execute(ctx context.Context, node *node.Node) (interface{}, error) {
	if node == nil {
		return nil, fmt.Errorf("node cannot be nil")
	}

	if strings.TrimSpace(node.Location) == "" {
		return nil, fmt.Errorf("node location cannot be empty")
	}

	if strings.TrimSpace(node.ID) == "" {
		return nil, fmt.Errorf("node ID cannot be empty")
	}

	var result map[string]interface{}
	err := uc.database.FindByID(ctx, node.Location, node.ID, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
