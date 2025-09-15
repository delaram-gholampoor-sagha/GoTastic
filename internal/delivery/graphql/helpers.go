package graphql

import (
	"time"

	"github.com/delaram/GoTastic/internal/delivery/graphql/model"
	"github.com/delaram/GoTastic/internal/domain"
)

func toModelTodoPtr(t *domain.TodoItem) *model.Todo {
	if t == nil {
		return nil
	}

	var due time.Time
	if t.DueDate != nil {
		due = *t.DueDate
	}

	var filePtr *string
	if t.FileID != nil && *t.FileID != "" {
		filePtr = t.FileID
	}

	return &model.Todo{
		ID:          t.UUID,
		Description: t.Description,
		DueDate:     due,
		FileID:      filePtr,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toModelTodosPtr(in []*domain.TodoItem) []*model.Todo {
	out := make([]*model.Todo, 0, len(in))
	for _, it := range in {
		out = append(out, toModelTodoPtr(it))
	}
	return out
}
