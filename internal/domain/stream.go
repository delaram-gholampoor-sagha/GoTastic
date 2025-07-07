package domain

import "context"


type StreamPublisher interface {
	PublishTodoItem(ctx context.Context, todo *TodoItem) error
}
