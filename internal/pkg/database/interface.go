package database

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("document not found")
var ErrInvalidID = errors.New("invalid document ID")

type Database interface {
	Insert(ctx context.Context, collection, id string, document any) error
	FindByID(ctx context.Context, collection, id string, dest any) error
	Update(ctx context.Context, collection, id string, update any) error
	Delete(ctx context.Context, collection, id string) error
	Close(ctx context.Context) error
}
