package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mgds/internal/pkg/env"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue[T any] struct {
	client    *redis.Client
	queueName string
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewRedisQueue[T any](uri string, password string, queueName QueueName) (*RedisQueue[T], error) {
	client := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithCancel(context.Background())

	if err := client.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisQueue[T]{
		client:    client,
		queueName: string(queueName),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (r *RedisQueue[T]) PublishMessage(msg T) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	return r.client.LPush(r.ctx, r.queueName, data).Err()
}

func (r *RedisQueue[T]) ConsumeMessages(handler func(T) error) error {
	for {
		if r.ctx.Err() != nil {
			return r.ctx.Err()
		}

		result, err := r.client.BRPop(r.ctx, 1*time.Second, r.queueName).Result()

		if err != nil {
			if err == redis.Nil {
				continue
			}
			if err == context.Canceled {
				return nil
			}
			log.Printf("Error getting message from queue (%s): %v", r.queueName, err)
			continue
		}

		payload := result[1]

		var msg T
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if err := handler(msg); err != nil {
			log.Printf("Error processing message for queue (%s - worker : %s): %v", r.queueName, env.GetHostName(), err)
			// r.client.LPush(r.ctx, r.queueName, payload)
			continue
		}
	}
}

func (r *RedisQueue[T]) Close() error {
	r.cancel()
	return r.client.Close()
}
