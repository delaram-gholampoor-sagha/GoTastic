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
	TxStarter
	CreateTx(ctx context.Context, tx Tx, todo *domain.TodoItem) error
	Create(ctx context.Context, todo *domain.TodoItem) error
	GetByID(ctx context.Context, id string) (*domain.TodoItem, error)
	List(ctx context.Context) ([]*domain.TodoItem, error)
	ListPaged(ctx context.Context, f domain.TodoFilter, s domain.TodoSort, limit, offset int) ([]*domain.TodoItem, int64, error)
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

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TxStarter interface {
	BeginTx(ctx context.Context) (Tx, error)
}

type StreamPublisher interface {
	PublishTodoItem(ctx context.Context, todo *domain.TodoItem) error
}
