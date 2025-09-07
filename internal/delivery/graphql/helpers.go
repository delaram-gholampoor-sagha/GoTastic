package graphql

import (
	"github.com/delaram/GoTastic/internal/delivery/graphql/model"
	"github.com/delaram/GoTastic/internal/domain"
)

func toModelTodoPtr(t *domain.TodoItem) *model.Todo {
	if t == nil {
		return nil
	}
	id := t.ID.String()
	var filePtr *string
	if t.FileID != "" {
		f := t.FileID
		filePtr = &f
	}
	return &model.Todo{
		ID:          id,
		Description: t.Description,
		DueDate:     t.DueDate, // time.Time
		FileID:      filePtr,
		CreatedAt:   t.CreatedAt, // time.Time
		UpdatedAt:   t.UpdatedAt, // time.Time
	}
}

func toModelTodosPtr(in []*domain.TodoItem) []*model.Todo {
	out := make([]*model.Todo, 0, len(in))
	for _, it := range in {
		out = append(out, toModelTodoPtr(it))
	}
	return out
}
