package http

import (
	"io"
	"net/http"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/internal/usecase"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


type Handler struct {
	logger      logger.Logger
	todoUseCase *usecase.TodoUseCase
	fileUseCase *usecase.FileUseCase
}


func NewHandler(logger logger.Logger, todoUseCase *usecase.TodoUseCase, fileUseCase *usecase.FileUseCase) *Handler {
	return &Handler{
		logger:      logger,
		todoUseCase: todoUseCase,
		fileUseCase: fileUseCase,
	}
}


func (h *Handler) RegisterRoutes(r *gin.Engine) {

	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := r.Group("/api/v1")
	{
		todos := api.Group("/todos")
		{
			todos.GET("/", h.ListTodoItems)
			todos.POST("/", h.CreateTodoItem)
			todos.GET("/:id", h.GetTodoItem)
			todos.PUT("/:id", h.UpdateTodoItem)
			todos.DELETE("/:id", h.DeleteTodoItem)
		}
		files := api.Group("/files")
		{
			files.POST("/", h.UploadFile)
			files.GET("/:id", h.DownloadFile)
			files.DELETE("/:id", h.DeleteFile)
		}
	}
}


func (h *Handler) ListTodoItems(c *gin.Context) {
	todos, err := h.todoUseCase.ListTodoItems(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list todo items", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list todo items",
			"details": err.Error(),
		})
		return
	}


	response := make([]gin.H, len(todos))
	for i, todo := range todos {
		response[i] = gin.H{
			"id":          todo.ID,
			"description": todo.Description,
			"due_date":    todo.DueDate,
			"file_id":     todo.FileID,
			"created_at":  todo.CreatedAt,
			"updated_at":  todo.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"todos": response,
	})
}


func (h *Handler) CreateTodoItem(c *gin.Context) {
	var req struct {
		Description string    `json:"description" binding:"required"`
		DueDate     time.Time `json:"due_date" binding:"required"`
		FileID      string    `json:"file_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to decode request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}


	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Description is required",
		})
		return
	}


	if req.DueDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Due date is required",
		})
		return
	}

	todo, err := h.todoUseCase.CreateTodoItem(c.Request.Context(), req.Description, req.DueDate, req.FileID)
	if err != nil {
		h.logger.Error("Failed to create todo item", err)
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create todo item",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          todo.ID,
		"description": todo.Description,
		"due_date":    todo.DueDate,
		"file_id":     todo.FileID,
		"created_at":  todo.CreatedAt,
		"updated_at":  todo.UpdatedAt,
	})
}


func (h *Handler) GetTodoItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}
	todo, err := h.todoUseCase.GetTodoItem(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get todo item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get todo item"})
		return
	}
	if todo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo item not found"})
		return
	}
	c.JSON(http.StatusOK, todo)
}


func (h *Handler) UpdateTodoItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID format"})
		return
	}

	var todo domain.TodoItem
	if err := c.ShouldBindJSON(&todo); err != nil {
		h.logger.Error("Failed to decode request body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	todo.ID = parsedID
	if err := h.todoUseCase.UpdateTodoItem(c.Request.Context(), &todo); err != nil {
		h.logger.Error("Failed to update todo item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo item"})
		return
	}
	c.Status(http.StatusNoContent)
}


func (h *Handler) DeleteTodoItem(c *gin.Context) {
	id := c.Param("id")
	if err := h.todoUseCase.DeleteTodoItem(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete todo item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo item"})
		return
	}
	c.Status(http.StatusNoContent)
}


func (h *Handler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get file from request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()
	fileID, err := h.fileUseCase.UploadFile(c.Request.Context(), file, header.Filename)
	if err != nil {
		h.logger.Error("Failed to upload file", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"file_id": fileID})
}


func (h *Handler) DownloadFile(c *gin.Context) {
	id := c.Param("id")
	reader, err := h.fileUseCase.DownloadFile(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to download file", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		return
	}
	defer reader.Close()
	io.Copy(c.Writer, reader)
}

		
func (h *Handler) DeleteFile(c *gin.Context) {
	id := c.Param("id")
	if err := h.fileUseCase.DeleteFile(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete file", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}
	c.Status(http.StatusNoContent)
}
