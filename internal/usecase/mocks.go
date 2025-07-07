package usecase

import (
	"context"
	"io"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockTodoRepository struct {
	mock.Mock
}

func (m *MockTodoRepository) Create(ctx context.Context, todo *domain.TodoItem) error {
	args := m.Called(ctx, todo)
	return args.Error(0)
}

func (m *MockTodoRepository) GetByID(ctx context.Context, id string) (*domain.TodoItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TodoItem), args.Error(1)
}

func (m *MockTodoRepository) List(ctx context.Context) ([]*domain.TodoItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.TodoItem), args.Error(1)
}

func (m *MockTodoRepository) Update(ctx context.Context, todo *domain.TodoItem) error {
	args := m.Called(ctx, todo)
	return args.Error(0)
}

func (m *MockTodoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	args := m.Called(ctx, file, filename)
	return args.String(0), args.Error(1)
}

func (m *MockFileRepository) Download(ctx context.Context, fileID string) (io.ReadCloser, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockFileRepository) Delete(ctx context.Context, fileID string) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockFileRepository) Exists(ctx context.Context, fileID string) (bool, error) {
	args := m.Called(ctx, fileID)
	return args.Bool(0), args.Error(1)
}

type MockStreamPublisher struct {
	mock.Mock
}

func (m *MockStreamPublisher) PublishTodoItem(ctx context.Context, todo *domain.TodoItem) error {
	args := m.Called(ctx, todo)
	return args.Error(0)
}
		
func (m *MockStreamPublisher) PublishTodoItems(ctx context.Context, todos []*domain.TodoItem) error {
	args := m.Called(ctx, todos)
	return args.Error(0)
}
