package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisRequestResponseQueue[T any, R any] struct {
	*RedisQueue[RequestMessage]
}

func NewRedisRequestResponseQueue[T any, R any](uri string, password string, queueName QueueName) (*RedisRequestResponseQueue[T, R], error) {
	baseQueue, err := NewRedisQueue[RequestMessage](uri, password, queueName)
	if err != nil {
		return nil, err
	}

	return &RedisRequestResponseQueue[T, R]{
		RedisQueue: baseQueue,
	}, nil
}

func (r *RedisRequestResponseQueue[T, R]) SendAndWait(ctx context.Context, data T, timeout time.Duration) (R, error) {
	var result R

	requestID := uuid.New().String()
	replyQueue := r.queueName + "_reply_" + requestID

	defer r.client.Del(ctx, replyQueue)

	request := RequestMessage{
		ID:      requestID,
		ReplyTo: replyQueue,
		Data:    data,
	}

	if err := r.PublishMessage(request); err != nil {
		return result, fmt.Errorf("failed to publish request: %v", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	response, err := r.client.BRPop(timeoutCtx, timeout, replyQueue).Result()
	if err != nil {
		if err == redis.Nil {
			return result, fmt.Errorf("timeout waiting for response")
		}
		return result, fmt.Errorf("failed to receive response: %v", err)
	}

	var responseMsg ResponseMessage
	if err := json.Unmarshal([]byte(response[1]), &responseMsg); err != nil {
		return result, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if responseMsg.Error != "" {
		return result, fmt.Errorf("remote error: %s", responseMsg.Error)
	}

	responseData, err := json.Marshal(responseMsg.Data)
	if err != nil {
		return result, fmt.Errorf("failed to marshal response data: %v", err)
	}

	if err := json.Unmarshal(responseData, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal response data: %v", err)
	}

	return result, nil
}

func (r *RedisRequestResponseQueue[T, R]) ConsumeWithReply(handler func(T) (R, error)) error {
	return r.ConsumeMessages(func(req RequestMessage) error {
		var requestData T

		reqDataBytes, err := json.Marshal(req.Data)
		if err != nil {
			log.Printf("Error marshaling request data: %v", err)
			return r.sendErrorResponse(req, fmt.Sprintf("failed to marshal request data: %v", err))
		}

		if err := json.Unmarshal(reqDataBytes, &requestData); err != nil {
			log.Printf("Error unmarshaling request data: %v", err)
			return r.sendErrorResponse(req, fmt.Sprintf("failed to unmarshal request data: %v", err))
		}

		response, err := handler(requestData)
		if err != nil {
			log.Printf("Error processing request %s: %v", req.ID, err)
			return r.sendErrorResponse(req, err.Error())
		}

		return r.sendSuccessResponse(req, response)
	})
}

func (r *RedisRequestResponseQueue[T, R]) sendErrorResponse(req RequestMessage, errorMsg string) error {
	response := ResponseMessage{
		ID:    req.ID,
		Error: errorMsg,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling error response: %v", err)
		return err
	}

	return r.client.LPush(r.ctx, req.ReplyTo, responseData).Err()
}

func (r *RedisRequestResponseQueue[T, R]) sendSuccessResponse(req RequestMessage, data R) error {
	response := ResponseMessage{
		ID:   req.ID,
		Data: data,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling success response: %v", err)
		return err
	}

	return r.client.LPush(r.ctx, req.ReplyTo, responseData).Err()
}
