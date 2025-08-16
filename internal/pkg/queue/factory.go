package queue

import "mgds/internal/pkg/env"

var REDIS_URI = env.GetOrDefault("REDIS_URI", "localhost:6379")
var REDIS_PASSWORD = env.GetOrDefault("REDIS_PASSWORD", "")

func Create[T any](queueName QueueName) (Queue[T], error) {
	return NewRedisQueue[T](REDIS_URI, REDIS_PASSWORD, queueName)
}

func CreateRequestResponse[T any, R any](queueName QueueName) (RequestResponseQueue[T, R], error) {
	return NewRedisRequestResponseQueue[T, R](REDIS_URI, REDIS_PASSWORD, queueName)
}
