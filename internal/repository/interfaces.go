package repository

import (
	"context"
	"io"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
)


var (
	ErrNotFound = NewError("not found")
)


type Error struct {
	message string
}


func NewError(message string) error {
	return &Error{message: message}
}


func (e *Error) Error() string {
	return e.message
}


type TodoRepository interface {
	Create(ctx context.Context, todo *domain.TodoItem) error
	GetByID(ctx context.Context, id string) (*domain.TodoItem, error)
	List(ctx context.Context) ([]*domain.TodoItem, error)
	Update(ctx context.Context, todo *domain.TodoItem) error
	Delete(ctx context.Context, id string) error
}


type FileRepository interface {
	Upload(ctx context.Context, reader io.Reader, filename string) (string, error)
	Download(ctx context.Context, id string) (io.ReadCloser, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}

	
type CacheRepository interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}
