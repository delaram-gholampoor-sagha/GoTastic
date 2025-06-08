package redis

import (
	"context"
	"encoding/json"

	"github.com/delaram/GoTastic/internal/repository"
	"github.com/redis/go-redis/v9"
)

// MessageBroker implements repository.MessageBroker for Redis
type MessageBroker struct {
	client *redis.Client
}

// NewMessageBroker creates a new Redis message broker
func NewMessageBroker(client *redis.Client) repository.MessageBroker {
	return &MessageBroker{
		client: client,
	}
}

// Publish sends a message to a Redis stream
func (b *MessageBroker) Publish(ctx context.Context, stream string, message interface{}) error {
	// Marshal message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Add message to stream
	_, err = b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"payload": payload,
		},
	}).Result()

	return err
}
