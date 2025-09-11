package redis

import (
	"context"
	"encoding/json"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	todoStreamKey = "todo:stream"
)


type StreamPublisher struct {
	client *redis.Client
	logger logger.Logger
}


func NewStreamPublisher(client *redis.Client, logger logger.Logger) *StreamPublisher {
	return &StreamPublisher{
		client: client,
		logger: logger,
	}
}


func (p *StreamPublisher) PublishTodoItem(ctx context.Context, todo *domain.TodoItem) error {
	data, err := json.Marshal(todo)
	if err != nil {
		return err
	}

					
	_, err = p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: todoStreamKey,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Result()

	if err != nil {
		p.logger.Error("Failed to publish todo item to stream", err)
		return err
	}

	p.logger.Info("Successfully published todo item to stream")
	return nil
}
