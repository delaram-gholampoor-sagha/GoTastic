package usecase

import (
	"context"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/google/uuid"
)

// TodoUseCase handles todo business logic
type TodoUseCase struct {
	logger          logger.Logger
	todoRepo        repository.TodoRepository
	fileRepo        repository.FileRepository
	cacheRepo       repository.CacheRepository
	streamPublisher domain.StreamPublisher
}

// NewTodoUseCase creates a new todo usecase
func NewTodoUseCase(logger logger.Logger, todoRepo repository.TodoRepository, fileRepo repository.FileRepository, cacheRepo repository.CacheRepository, streamPublisher domain.StreamPublisher) *TodoUseCase {
	return &TodoUseCase{
		logger:          logger,
		todoRepo:        todoRepo,
		fileRepo:        fileRepo,
		cacheRepo:       cacheRepo,
		streamPublisher: streamPublisher,
	}
}

// CreateTodoItem creates a new todo item
func (u *TodoUseCase) CreateTodoItem(ctx context.Context, description string, dueDate time.Time, fileID string) (*domain.TodoItem, error) {
	// Validate file ID if provided
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

	// Create todo item
	todo := &domain.TodoItem{
		ID:          uuid.New(),
		Description: description,
		DueDate:     dueDate,
		FileID:      fileID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := u.todoRepo.Create(ctx, todo); err != nil {
		u.logger.Error("Failed to create todo", err)
		return nil, err
	}

	// Invalidate cache
	if err := u.cacheRepo.Delete(ctx, "todos"); err != nil {
		u.logger.Warn("Failed to invalidate cache", err)
	}

	// Publish to stream
	if err := u.streamPublisher.PublishTodoItem(ctx, todo); err != nil {
		u.logger.Warn("Failed to publish todo item to stream", err)
	}

	return todo, nil
}

// GetTodoItem retrieves a todo item by ID
func (u *TodoUseCase) GetTodoItem(ctx context.Context, id string) (*domain.TodoItem, error) {
	cacheKey := "todo:" + id
	cached, _ := u.cacheRepo.Get(ctx, cacheKey)
	if cached != nil {
		if todo, ok := cached.(*domain.TodoItem); ok {
			return todo, nil
		}
	}
	todo, err := u.todoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.cacheRepo.Set(ctx, cacheKey, todo, time.Hour)

	// Publish to stream
	if err := u.streamPublisher.PublishTodoItem(ctx, todo); err != nil {
		u.logger.Warn("Failed to publish todo item to stream", err)
	}

	return todo, nil
}

// ListTodoItems retrieves all todo items
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

	// Publish to stream (if method exists)
	if publisher, ok := u.streamPublisher.(interface {
		PublishTodoItems(ctx context.Context, todos []*domain.TodoItem) error
	}); ok {
		if err := publisher.PublishTodoItems(ctx, todos); err != nil {
			u.logger.Warn("Failed to publish todo items to stream", err)
		}
	}

	return todos, nil
}

// UpdateTodoItem updates a todo item
func (u *TodoUseCase) UpdateTodoItem(ctx context.Context, todo *domain.TodoItem) error {
	// Validate file ID if provided
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

	// Update timestamp
	todo.UpdatedAt = time.Now()

	// Update in database
	if err := u.todoRepo.Update(ctx, todo); err != nil {
		u.logger.Error("Failed to update todo", err)
		return err
	}

	// Invalidate cache
	if err := u.cacheRepo.Delete(ctx, "todo:"+todo.ID.String()); err != nil {
		u.logger.Warn("Failed to invalidate todo cache", err)
	}
	if err := u.cacheRepo.Delete(ctx, "todos"); err != nil {
		u.logger.Warn("Failed to invalidate todos cache", err)
	}

	// Publish to stream
	if err := u.streamPublisher.PublishTodoItem(ctx, todo); err != nil {
		u.logger.Warn("Failed to publish todo item to stream", err)
	}

	return nil
}

// DeleteTodoItem deletes a todo item
func (u *TodoUseCase) DeleteTodoItem(ctx context.Context, id string) error {
	// Delete from database
	if err := u.todoRepo.Delete(ctx, id); err != nil {
		u.logger.Error("Failed to delete todo", err)
		return err
	}

	// Invalidate cache
	if err := u.cacheRepo.Delete(ctx, "todo:"+id); err != nil {
		u.logger.Warn("Failed to invalidate todo cache", err)
	}
	if err := u.cacheRepo.Delete(ctx, "todos"); err != nil {
		u.logger.Warn("Failed to invalidate todos cache", err)
	}

	// Publish to stream
	todo := &domain.TodoItem{ID: uuid.MustParse(id)}
	if err := u.streamPublisher.PublishTodoItem(ctx, todo); err != nil {
		u.logger.Warn("Failed to publish todo item to stream", err)
	}

	return nil
}
