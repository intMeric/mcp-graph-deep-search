package search_document

import (
	"context"

	"mgds/internal/pkg/node"
)

type SearchDocumentUseCase interface {
	Execute(ctx context.Context, node *node.Node) (interface{}, error)
}
