package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/redis/go-redis/v9"
)


type RedisCacheRepository struct {
	logger logger.Logger
	client *redis.Client
}


func NewRedisCacheRepository(logger logger.Logger, client *redis.Client) CacheRepository {
	return &RedisCacheRepository{
		logger: logger,
		client: client,
	}
}


func (r *RedisCacheRepository) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("Failed to get from cache", err)
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		r.logger.Error("Failed to unmarshal cache value", err)
		return nil, err
	}
	return result, nil
}


func (r *RedisCacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.Error("Failed to marshal cache value", err)
		return err
	}
	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		r.logger.Error("Failed to set cache value", err)
		return err
	}
	return nil
}

						
func (r *RedisCacheRepository) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.logger.Error("Failed to delete cache value", err)
		return err
	}
	return nil
}
