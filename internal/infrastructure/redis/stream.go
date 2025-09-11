package redis

import (
	"context"
	"encoding/json"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/latolukasz/beeorm"
)

const (
	todoStreamKey = "todo:stream"
)

type StreamPublisher struct {
	cache  *beeorm.RedisCache
	logger logger.Logger
}

func NewStreamPublisher(cache *beeorm.RedisCache, logger logger.Logger) *StreamPublisher {
	return &StreamPublisher{
		cache:  cache,
		logger: logger,
	}
}

func (p *StreamPublisher) PublishTodoItem(ctx context.Context, todo *domain.TodoItem) error {
	_ = ctx

	payload, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	pl := p.cache.PipeLine()
	idCmd := pl.XAdd(todoStreamKey, []string{"data", string(payload)})
	pl.Exec()

	_ = idCmd.Result()
	return nil
}
