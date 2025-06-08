package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// BenchmarkCreateTodoItem benchmarks the CreateTodoItem operation
func BenchmarkCreateTodoItem(b *testing.B) {
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	useCase := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	// Cleanup
	defer func() {
		mockTodoRepo.AssertExpectations(b)
		mockFileRepo.AssertExpectations(b)
		mockCacheRepo.AssertExpectations(b)
		mockStreamPublisher.AssertExpectations(b)
	}()

	// Set up mock expectations
	mockFileRepo.On("Exists", mock.Anything, "test-file-id").Return(true, nil)
	mockTodoRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todos").Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := useCase.CreateTodoItem(ctx, "test description", time.Now(), "test-file-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetTodoItem benchmarks the GetTodoItem operation
func BenchmarkGetTodoItem(b *testing.B) {
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	useCase := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	// Cleanup
	defer func() {
		mockTodoRepo.AssertExpectations(b)
		mockFileRepo.AssertExpectations(b)
		mockCacheRepo.AssertExpectations(b)
		mockStreamPublisher.AssertExpectations(b)
	}()

	todoID := uuid.New()
	expectedTodo := &domain.TodoItem{
		ID:          todoID,
		Description: "Test todo",
		DueDate:     time.Now().Add(24 * time.Hour),
	}

	mockCacheRepo.On("Get", mock.Anything, "todo:"+todoID.String()).Return(nil, errors.New("cache miss"))
	mockTodoRepo.On("GetByID", mock.Anything, todoID.String()).Return(expectedTodo, nil)
	mockCacheRepo.On("Set", mock.Anything, "todo:"+todoID.String(), expectedTodo, time.Hour).Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, expectedTodo).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := useCase.GetTodoItem(ctx, todoID.String())
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListTodoItems benchmarks the ListTodoItems operation
func BenchmarkListTodoItems(b *testing.B) {
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	useCase := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	// Cleanup
	defer func() {
		mockTodoRepo.AssertExpectations(b)
		mockFileRepo.AssertExpectations(b)
		mockCacheRepo.AssertExpectations(b)
		mockStreamPublisher.AssertExpectations(b)
	}()

	expectedTodos := []*domain.TodoItem{
		{
			ID:          uuid.New(),
			Description: "Todo 1",
			DueDate:     time.Now().Add(24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Description: "Todo 2",
			DueDate:     time.Now().Add(48 * time.Hour),
		},
	}

	mockCacheRepo.On("Get", mock.Anything, "todos").Return(nil, errors.New("cache miss"))
	mockTodoRepo.On("List", mock.Anything).Return(expectedTodos, nil)
	mockCacheRepo.On("Set", mock.Anything, "todos", expectedTodos, time.Hour).Return(nil)
	mockStreamPublisher.On("PublishTodoItems", mock.Anything, expectedTodos).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := useCase.ListTodoItems(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpdateTodoItem benchmarks the UpdateTodoItem operation
func BenchmarkUpdateTodoItem(b *testing.B) {
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	useCase := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	todo := &domain.TodoItem{
		ID:          uuid.New(),
		Description: "Updated todo",
		DueDate:     time.Now().Add(24 * time.Hour),
		FileID:      "updated-file",
	}

	mockTodoRepo.On("Update", mock.Anything, todo).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, mock.MatchedBy(func(key string) bool {
		return key == "todos" || (len(key) > 5 && key[:5] == "todo:")
	})).Return(nil)
	mockFileRepo.On("Exists", mock.Anything, "updated-file").Return(true, nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, todo).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = useCase.UpdateTodoItem(context.Background(), todo)
	}
}
