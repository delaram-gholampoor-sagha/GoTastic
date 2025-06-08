package repository

import (
	"context"
	"io"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
)

// Common errors
var (
	ErrNotFound = NewError("not found")
)

// Error represents a repository error
type Error struct {
	message string
}

// NewError creates a new repository error
func NewError(message string) error {
	return &Error{message: message}
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.message
}

// TodoRepository defines the interface for todo storage operations
type TodoRepository interface {
	Create(ctx context.Context, todo *domain.TodoItem) error
	GetByID(ctx context.Context, id string) (*domain.TodoItem, error)
	List(ctx context.Context) ([]*domain.TodoItem, error)
	Update(ctx context.Context, todo *domain.TodoItem) error
	Delete(ctx context.Context, id string) error
}

// FileRepository defines the interface for file storage operations
type FileRepository interface {
	Upload(ctx context.Context, reader io.Reader, filename string) (string, error)
	Download(ctx context.Context, id string) (io.ReadCloser, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}
