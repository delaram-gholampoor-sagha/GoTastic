package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTodoItem(t *testing.T) {
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	uc := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	defer func() {
		mockTodoRepo.AssertExpectations(t)
		mockFileRepo.AssertExpectations(t)
		mockCacheRepo.AssertExpectations(t)
		mockStreamPublisher.AssertExpectations(t)
	}()

	description := "Test todo"
	dueDate := time.Now().Add(24 * time.Hour)
	fileID := "test-file-id"

	mockFileRepo.On("Exists", mock.Anything, fileID).Return(true, nil)
	mockTodoRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todos").Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, mock.MatchedBy(func(todo *domain.TodoItem) bool {
		return todo.Description == description && todo.DueDate.Equal(dueDate) && todo.FileID == fileID
	})).Return(nil)

	todo, err := uc.CreateTodoItem(context.Background(), description, dueDate, fileID)

	assert.NoError(t, err)
	assert.NotNil(t, todo)
	assert.Equal(t, description, todo.Description)
	assert.Equal(t, dueDate, todo.DueDate)
	assert.Equal(t, fileID, todo.FileID)
}

func TestGetTodoItem(t *testing.T) {
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	uc := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	id := uuid.New()
	expectedTodo := &domain.TodoItem{
		ID:          id,
		Description: "Test todo",
		DueDate:     time.Now().Add(24 * time.Hour),
		FileID:      "test-file-id",
	}

	mockCacheRepo.On("Get", mock.Anything, "todo:"+id.String()).Return(nil, nil)
	mockTodoRepo.On("GetByID", mock.Anything, id.String()).Return(expectedTodo, nil)
	mockCacheRepo.On("Set", mock.Anything, "todo:"+id.String(), expectedTodo, time.Hour).Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, mock.MatchedBy(func(todo *domain.TodoItem) bool {
		return todo.ID == id
	})).Return(nil)

	todo, err := uc.GetTodoItem(context.Background(), id.String())

	assert.NoError(t, err)
	assert.Equal(t, expectedTodo, todo)
	mockTodoRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockStreamPublisher.AssertExpectations(t)
}

func TestListTodoItems(t *testing.T) {
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	uc := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	expectedTodos := []*domain.TodoItem{
		{
			ID:          uuid.New(),
			Description: "Todo 1",
			DueDate:     time.Now().Add(24 * time.Hour),
			FileID:      "file-1",
		},
		{
			ID:          uuid.New(),
			Description: "Todo 2",
			DueDate:     time.Now().Add(48 * time.Hour),
			FileID:      "file-2",
		},
	}

	
	mockCacheRepo.On("Get", mock.Anything, "todos").Return(nil, nil)
	mockTodoRepo.On("List", mock.Anything).Return(expectedTodos, nil)
	mockCacheRepo.On("Set", mock.Anything, "todos", expectedTodos, time.Hour).Return(nil)
	mockStreamPublisher.On("PublishTodoItems", mock.Anything, mock.Anything).Return(nil)

	
	todos, err := uc.ListTodoItems(context.Background())

	
	assert.NoError(t, err)
	assert.Equal(t, expectedTodos, todos)
	mockTodoRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockStreamPublisher.AssertExpectations(t)
}

func TestUpdateTodoItem(t *testing.T) {
	
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	uc := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	
	todo := &domain.TodoItem{
		ID:          uuid.New(),
		Description: "Updated todo",
		DueDate:     time.Now().Add(24 * time.Hour),
		FileID:      "updated-file",
	}

	
	mockFileRepo.On("Exists", mock.Anything, todo.FileID).Return(true, nil)
	mockTodoRepo.On("Update", mock.Anything, todo).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todos").Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todo:"+todo.ID.String()).Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, mock.MatchedBy(func(updatedTodo *domain.TodoItem) bool {
		return updatedTodo.ID == todo.ID && updatedTodo.Description == todo.Description
	})).Return(nil)

	
	err := uc.UpdateTodoItem(context.Background(), todo)

	
	assert.NoError(t, err)
	mockFileRepo.AssertExpectations(t)
	mockTodoRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockStreamPublisher.AssertExpectations(t)
}

func TestDeleteTodoItem(t *testing.T) {
	
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(MockTodoRepository)
	mockFileRepo := new(MockFileRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockStreamPublisher := new(MockStreamPublisher)
	uc := NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)

	
	id := uuid.New().String()

	
	mockTodoRepo.On("Delete", mock.Anything, id).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todos").Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todo:"+id).Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, mock.MatchedBy(func(todo *domain.TodoItem) bool {
		return todo.ID.String() == id
	})).Return(nil)

			
	err := uc.DeleteTodoItem(context.Background(), id)

	
	assert.NoError(t, err)
	mockTodoRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockStreamPublisher.AssertExpectations(t)
}
