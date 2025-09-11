package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/latolukasz/beeorm"
)

type RedisCacheRepository struct {
	logger logger.Logger
	cache  *beeorm.RedisCache
}

func NewRedisCacheRepository(logger logger.Logger, cache *beeorm.RedisCache) CacheRepository {
	return &RedisCacheRepository{logger: logger, cache: cache}
}

func (r *RedisCacheRepository) Get(ctx context.Context, key string) (interface{}, error) {
	val, has := r.cache.Get(key)
	if !has {
		return nil, nil
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
	ttlSeconds := int(expiration / time.Second)
	r.cache.Set(key, string(data), ttlSeconds)
	return nil
}

func (r *RedisCacheRepository) Delete(ctx context.Context, key string) error {
	r.cache.Del(key)
	return nil
}
