package domain

import "context"

// StreamPublisher defines the interface for publishing messages to a stream
type StreamPublisher interface {
	// PublishTodoItem publishes a todo item to the stream
	PublishTodoItem(ctx context.Context, todo *TodoItem) error
}
