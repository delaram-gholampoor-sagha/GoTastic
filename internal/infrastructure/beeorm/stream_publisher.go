package beeinfra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	beeorm "git.ice.global/packages/beeorm/v4"

	"github.com/delaram/GoTastic/internal/domain"
)

type StreamPublisher struct {
	engine *beeorm.Engine
	stream string
}

func NewStreamPublisher(engine *beeorm.Engine, stream string) *StreamPublisher {
	return &StreamPublisher{engine: engine, stream: stream}
}

func (p *StreamPublisher) PublishTodoItem(ctx context.Context, todo *domain.TodoItem) (err error) {
	// NOTE: FileID is optional; if you use *string in domain, handle nil accordingly.
	var dueStr string
	if todo.DueDate != nil {
		dueStr = todo.DueDate.UTC().Format(time.RFC3339Nano)
	}
	var fileStr string
	if todo.FileID != nil {
		fileStr = *todo.FileID
	}

	payload := struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		DueDate     string `json:"due_date,omitempty"`
		FileID      string `json:"file_id,omitempty"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}{
		ID:          todo.UUID, // public id
		Description: todo.Description,
		DueDate:     dueStr,
		FileID:      fileStr,
		CreatedAt:   todo.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:   todo.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.xaddOne(b)
}

// Optional bulk path (your use-case uses a type assertion for this)
func (p *StreamPublisher) PublishTodoItems(ctx context.Context, todos []*domain.TodoItem) (err error) {
	if len(todos) == 0 {
		return nil
	}
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("redis pipeline xadd failed: %v", rec)
		}
	}()

	pipe := p.engine.GetRedis().PipeLine() // uses the default registered Redis pool
	for _, todo := range todos {
		var dueStr string
		if todo.DueDate != nil {
			dueStr = todo.DueDate.UTC().Format(time.RFC3339Nano)
		}
		var fileStr string
		if todo.FileID != nil {
			fileStr = *todo.FileID
		}

		payload := struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			DueDate     string `json:"due_date,omitempty"`
			FileID      string `json:"file_id,omitempty"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		}{
			ID:          todo.UUID,
			Description: todo.Description,
			DueDate:     dueStr,
			FileID:      fileStr,
			CreatedAt:   todo.CreatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedAt:   todo.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}

		b, mErr := json.Marshal(payload)
		if mErr != nil {
			return mErr
		}
		_ = pipe.XAdd(p.stream, []string{"data", string(b)})
	}
	pipe.Exec() // panics on error under BeeORM
	return nil
}

// --- internals ---------------------------------------------------------------

func (p *StreamPublisher) xaddOne(b []byte) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("redis xadd failed: %v", rec)
		}
	}()
	pipe := p.engine.GetRedis().PipeLine()
	cmd := pipe.XAdd(p.stream, []string{"data", string(b)})
	pipe.Exec()      // will panic on failure; recover above turns it into error
	_ = cmd.Result() // touch the result; not strictly required
	return nil
}
