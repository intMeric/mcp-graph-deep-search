package queue

import (
	"context"
	"time"
)

type Queue[T any] interface {
	PublishMessage(msg T) error
	ConsumeMessages(func(T) error) error
	Close() error
}

type RequestMessage struct {
	ID      string `json:"id"`
	ReplyTo string `json:"reply_to"`
	Data    any    `json:"data"`
}

type ResponseMessage struct {
	ID    string `json:"id"`
	Data  any    `json:"data"`
	Error string `json:"error,omitempty"`
}

type RequestResponseQueue[T any, R any] interface {
	Queue[RequestMessage]

	SendAndWait(ctx context.Context, data T, timeout time.Duration) (R, error)
	ConsumeWithReply(handler func(T) (R, error)) error
}
