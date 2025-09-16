package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	streamPublisher repository.StreamPublisher
	outboxRepo      repository.OutboxRepository
}

func NewTodoUseCase(logger logger.Logger,
	todoRepo repository.TodoRepository,
	fileRepo repository.FileRepository,
	cacheRepo repository.CacheRepository,
	streamPublisher repository.StreamPublisher,
	outboxRepo repository.OutboxRepository,
) *TodoUseCase {
	return &TodoUseCase{
		logger:          logger,
		todoRepo:        todoRepo,
		fileRepo:        fileRepo,
		cacheRepo:       cacheRepo,
		streamPublisher: streamPublisher,
		outboxRepo:      outboxRepo,
	}
}

func (u *TodoUseCase) CreateTodoItem(ctx context.Context, description string, dueDate time.Time, fileID string) (*domain.TodoItem, error) {
	u.logger.Debug("Starting CreateTodoItem with description: %s, dueDate: %v, fileID: %s", description, dueDate, fileID)
	var filePtr *string
	if fileID != "" {
		exists, err := u.fileRepo.Exists(ctx, fileID)
		if err != nil {
			u.logger.Error("Failed to check file existence", err)
			return nil, err
		}
		if !exists {
			return nil, repository.ErrNotFound
		}
		filePtr = &fileID
		u.logger.Debug("FileID exists, set to: %s", fileID)
	} else {
		u.logger.Debug("No fileID provided")
	}

	todo := &domain.TodoItem{
		UUID:        uuid.NewString(),
		Description: description,
		DueDate:     &dueDate,
		FileID:      filePtr,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	u.logger.Debug("Created TodoItem: %+v", todo)

	txStarter, ok := u.todoRepo.(repository.TxStarter)
	if !ok {
		u.logger.Error("todoRepo does not support transactions", nil)
		return nil, fmt.Errorf("todoRepo does not support transactions")
	}
	tx, err := txStarter.BeginTx(ctx)
	if err != nil {
		u.logger.Error("Failed to begin transaction", err)
		return nil, err
	}
	u.logger.Debug("Transaction begun")
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			u.logger.Warn("Rollback failed", err)
		} else {
			u.logger.Debug("Transaction rolled back (deferred)")
		}
	}()

	if err := u.todoRepo.CreateTx(ctx, tx, todo); err != nil {
		u.logger.Error("Failed to create todo", err)
		return nil, err
	}
	u.logger.Debug("Todo created successfully with ID: %d", todo.ID)

	payload, err := json.Marshal(todo)
	if err != nil {
		u.logger.Error("Failed to marshal todo for outbox payload", err)
		return nil, err
	}
	u.logger.Debug("Outbox payload marshaled: %s", string(payload))

	outboxMsg := repository.OutboxMessage{
		AggregateType: "todo",
		AggregateID:   todo.UUID,
		EventType:     "todo.created",
		Payload:       payload,
		Headers:       map[string]string{"source": "api", "schema": "v1"},
	}
	u.logger.Debug("Outbox message prepared: %+v", outboxMsg)

	if err := u.outboxRepo.Insert(ctx, tx, outboxMsg); err != nil {
		u.logger.Error("Failed to insert outbox message", err)
		return nil, err
	}
	u.logger.Debug("Outbox message inserted successfully")

	if err := tx.Commit(ctx); err != nil {
		u.logger.Error("Failed to commit transaction", err)
		return nil, err
	}
	u.logger.Debug("Transaction committed successfully")

	if err := u.cacheRepo.Delete(ctx, "todos"); err != nil {
		u.logger.Warn("Failed to invalidate cache", err)
	} else {
		u.logger.Debug("Cache invalidated successfully")
	}

	return todo, nil
}

// usecase/todo.go (add method)
func (u *TodoUseCase) ListTodoItemsPaged(ctx context.Context, f domain.TodoFilter, s domain.TodoSort, limit, offset int) ([]*domain.TodoItem, int64, error) {
	// enforce sane caps
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.todoRepo.ListPaged(ctx, f, s, limit, offset)
}

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

	if err := u.streamPublisher.PublishTodoItem(ctx, todo); err != nil {
		u.logger.Warn("Failed to publish todo item to stream", err)
	}

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
	if todo.FileID != nil && *todo.FileID != "" {
		exists, err := u.fileRepo.Exists(ctx, *todo.FileID)
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

	if err := u.cacheRepo.Delete(ctx, "todo:"+strconv.FormatUint(todo.ID, 10)); err != nil {
		u.logger.Warn("Failed to invalidate todo cache", err)
	}

	if err := u.streamPublisher.PublishTodoItem(ctx, todo); err != nil {
		u.logger.Warn("Failed to publish todo item to stream", err)
	}

	return nil
}

func (u *TodoUseCase) DeleteTodoItem(ctx context.Context, uuid string) error {
	todo, err := u.todoRepo.GetByID(ctx, uuid)
	if err != nil {
		u.logger.Error("Failed to get todo for deletion", err)
		return err
	}

	if err := u.todoRepo.Delete(ctx, uuid); err != nil {
		u.logger.Error("Failed to delete todo", err)
		return err
	}

	if err := u.cacheRepo.Delete(ctx, "todo:"+uuid); err != nil {
		u.logger.Warn("Failed to invalidate todo cache", err)
	}

	if err := u.streamPublisher.PublishTodoItem(ctx, todo); err != nil {
		u.logger.Warn("Failed to publish todo item to stream", err)
	}

	return nil
}
