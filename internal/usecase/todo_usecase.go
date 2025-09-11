package usecase

import (
	"context"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/google/uuid"
)

type TodoUseCase struct {
	logger          logger.Logger
	todoRepo        repository.TodoRepository
	fileRepo        repository.FileRepository
	cacheRepo       repository.CacheRepository
	streamPublisher domain.StreamPublisher
}

func NewTodoUseCase(logger logger.Logger, todoRepo repository.TodoRepository, fileRepo repository.FileRepository, cacheRepo repository.CacheRepository, streamPublisher domain.StreamPublisher) *TodoUseCase {
	return &TodoUseCase{
		logger:          logger,
		todoRepo:        todoRepo,
		fileRepo:        fileRepo,
		cacheRepo:       cacheRepo,
		streamPublisher: streamPublisher,
	}
}

func (u *TodoUseCase) CreateTodoItem(ctx context.Context, description string, dueDate time.Time, fileID string) (*domain.TodoItem, error) {
	if fileID != "" {
		exists, err := u.fileRepo.Exists(ctx, fileID)
		if err != nil {
			u.logger.Error("Failed to check file existence", err)
			return nil, err
		}
		if !exists {
			return nil, repository.ErrNotFound
		}
	}

	todo := &domain.TodoItem{
		PublicID:    uuid.NewString(),
		Description: description,
		DueDate:     dueDate,
		FileID:      fileID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.todoRepo.Create(ctx, todo); err != nil {
		u.logger.Error("Failed to create todo", err)
		return nil, err
	}

	_ = u.cacheRepo.Delete(ctx, "todos")
	_ = u.streamPublisher.PublishTodoItem(ctx, todo)

	return todo, nil
}

func (u *TodoUseCase) GetTodoItem(ctx context.Context, id string) (*domain.TodoItem, error) {
	cacheKey := "todo:" + id
	if cached, _ := u.cacheRepo.Get(ctx, cacheKey); cached != nil {
		// NOTE: your cache stores JSON into interface{};
		// type-assert here will usually fail.
	}

	todo, err := u.todoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = u.cacheRepo.Set(ctx, cacheKey, todo, time.Hour)
	_ = u.streamPublisher.PublishTodoItem(ctx, todo)
	return todo, nil
}

func (u *TodoUseCase) ListTodoItems(ctx context.Context) ([]*domain.TodoItem, error) {
	cacheKey := "todos"
	cached, _ := u.cacheRepo.Get(ctx, cacheKey)
	if cached != nil {
		if todos, ok := cached.([]*domain.TodoItem); ok {
			return todos, nil
		}
	}
	todos, err := u.todoRepo.List(ctx)
	if err != nil {
		u.logger.Error("Failed to list todos", err)
		return nil, err
	}
	u.cacheRepo.Set(ctx, cacheKey, todos, time.Hour)

	if publisher, ok := u.streamPublisher.(interface {
		PublishTodoItems(ctx context.Context, todos []*domain.TodoItem) error
	}); ok {
		if err := publisher.PublishTodoItems(ctx, todos); err != nil {
			u.logger.Warn("Failed to publish todo items to stream", err)
		}
	}

	return todos, nil
}

func (u *TodoUseCase) UpdateTodoItem(ctx context.Context, todo *domain.TodoItem) error {
	if todo.FileID != "" {
		exists, err := u.fileRepo.Exists(ctx, todo.FileID)
		if err != nil {
			u.logger.Error("Failed to check file existence", err)
			return err
		}
		if !exists {
			return repository.ErrNotFound
		}
	}
	todo.UpdatedAt = time.Now()

	if err := u.todoRepo.Update(ctx, todo); err != nil {
		u.logger.Error("Failed to update todo", err)
		return err
	}

	_ = u.cacheRepo.Delete(ctx, "todo:"+todo.PublicID)
	_ = u.cacheRepo.Delete(ctx, "todos")
	_ = u.streamPublisher.PublishTodoItem(ctx, todo)
	return nil
}

func (u *TodoUseCase) DeleteTodoItem(ctx context.Context, id string) error {
	if err := u.todoRepo.Delete(ctx, id); err != nil {
		u.logger.Error("Failed to delete todo", err)
		return err
	}

	_ = u.cacheRepo.Delete(ctx, "todo:"+id)
	_ = u.cacheRepo.Delete(ctx, "todos")
	_ = u.streamPublisher.PublishTodoItem(ctx, &domain.TodoItem{PublicID: id})
	return nil
}
