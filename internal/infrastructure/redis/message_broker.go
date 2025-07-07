package redis

import (
	"context"
	"encoding/json"

	"github.com/delaram/GoTastic/internal/repository"
	"github.com/redis/go-redis/v9"
)


type MessageBroker struct {
	client *redis.Client
}


func NewMessageBroker(client *redis.Client) repository.MessageBroker {
	return &MessageBroker{
		client: client,
	}
}


func (b *MessageBroker) Publish(ctx context.Context, stream string, message interface{}) error {

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

				
	_, err = b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"payload": payload,
		},
	}).Result()

	return err
}
