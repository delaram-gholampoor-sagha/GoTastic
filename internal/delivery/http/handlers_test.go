package http

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/usecase"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestHandler() (*Handler, *usecase.MockTodoRepository, *usecase.MockFileRepository, *usecase.MockCacheRepository, *usecase.MockStreamPublisher) {
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockTodoRepo := new(usecase.MockTodoRepository)
	mockFileRepo := new(usecase.MockFileRepository)
	mockCacheRepo := new(usecase.MockCacheRepository)
	mockStreamPublisher := new(usecase.MockStreamPublisher)

	todoUseCase := usecase.NewTodoUseCase(log, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher)
	fileUseCase := usecase.NewFileUseCase(log, mockFileRepo)

	handler := NewHandler(log, todoUseCase, fileUseCase)
	return handler, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher
}

func TestHandleFileUpload(t *testing.T) {
	handler, _, mockFileRepo, _, _ := setupTestHandler()


	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)
	part.Write([]byte("test content"))
	writer.Close()


	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req


	mockFileRepo.On("Upload", mock.Anything, mock.Anything, "test.txt").Return("test-file-id", nil)


	handler.UploadFile(c)


	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "test-file-id", response["file_id"])
}

func TestHandleCreateTodo(t *testing.T) {
	handler, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher := setupTestHandler()

	reqBody := map[string]interface{}{
		"description": "Test todo",
		"due_date":    time.Now().Add(24 * time.Hour),
		"file_id":     "test-file-id",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/todo", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req


	mockFileRepo.On("Exists", mock.Anything, "test-file-id").Return(true, nil)
	mockTodoRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todos").Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, mock.AnythingOfType("*domain.TodoItem")).Return(nil)


	handler.CreateTodoItem(c)


	assert.Equal(t, http.StatusCreated, w.Code)
	var response domain.TodoItem
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, reqBody["description"], response.Description)
	assert.Equal(t, reqBody["file_id"], response.FileID)
}

func TestHandleGetTodo(t *testing.T) {
	handler, mockTodoRepo, _, mockCacheRepo, mockStreamPublisher := setupTestHandler()


	id := uuid.New().String()
	expectedTodo := &domain.TodoItem{
		ID:          uuid.MustParse(id),
		Description: "Test todo",
		DueDate:     time.Now().Add(24 * time.Hour),
		FileID:      "test-file-id",
	}


	req := httptest.NewRequest("GET", "/todo/"+id, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: id}}


	cacheKey := "todo:" + id
	mockCacheRepo.On("Get", mock.Anything, cacheKey).Return(nil, nil)
	mockTodoRepo.On("GetByID", mock.Anything, id).Return(expectedTodo, nil)
	mockCacheRepo.On("Set", mock.Anything, cacheKey, expectedTodo, time.Hour).Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, expectedTodo).Return(nil)


	handler.GetTodoItem(c)


	assert.Equal(t, http.StatusOK, w.Code)
	var response domain.TodoItem
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTodo.ID, response.ID)
	assert.Equal(t, expectedTodo.Description, response.Description)
	assert.Equal(t, expectedTodo.FileID, response.FileID)
}

func TestHandleListTodos(t *testing.T) {
	handler, mockTodoRepo, _, mockCacheRepo, mockStreamPublisher := setupTestHandler()


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


	req := httptest.NewRequest("GET", "/todo", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req


	mockCacheRepo.On("Get", mock.Anything, "todos").Return(nil, nil)
	mockTodoRepo.On("List", mock.Anything).Return(expectedTodos, nil)
	mockCacheRepo.On("Set", mock.Anything, "todos", expectedTodos, time.Hour).Return(nil)
	mockStreamPublisher.On("PublishTodoItems", mock.Anything, mock.AnythingOfType("[]*domain.TodoItem")).Return(nil)


	handler.ListTodoItems(c)


	assert.Equal(t, http.StatusOK, w.Code)
	var response struct {
		Todos []domain.TodoItem `json:"todos"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response.Todos, 2)
}

func TestHandleUpdateTodo(t *testing.T) {
	handler, mockTodoRepo, mockFileRepo, mockCacheRepo, mockStreamPublisher := setupTestHandler()

	id := uuid.New().String()
	todo := domain.TodoItem{
		ID:          uuid.MustParse(id),
		Description: "Updated todo",
		DueDate:     time.Now().Add(24 * time.Hour),
		FileID:      "updated-file",
	}

	body, _ := json.Marshal(todo)

	mockFileRepo.On("Exists", mock.Anything, "updated-file").Return(true, nil)
	mockTodoRepo.On("Update", mock.Anything, mock.MatchedBy(func(t *domain.TodoItem) bool {
		return t.ID == todo.ID && t.Description == todo.Description
	})).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todo:"+id).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todos").Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, mock.AnythingOfType("*domain.TodoItem")).Return(nil)

	r := gin.Default()
	r.PUT("/todo/:id", handler.UpdateTodoItem)

	req := httptest.NewRequest("PUT", "/todo/"+id, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandleDeleteTodo(t *testing.T) {
	handler, mockTodoRepo, _, mockCacheRepo, mockStreamPublisher := setupTestHandler()

	id := uuid.New().String()

	mockTodoRepo.On("Delete", mock.Anything, id).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todo:"+id).Return(nil)
	mockCacheRepo.On("Delete", mock.Anything, "todos").Return(nil)
	mockStreamPublisher.On("PublishTodoItem", mock.Anything, mock.AnythingOfType("*domain.TodoItem")).Return(nil)

	r := gin.Default()
	r.DELETE("/todo/:id", handler.DeleteTodoItem)

	req := httptest.NewRequest("DELETE", "/todo/"+id, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
