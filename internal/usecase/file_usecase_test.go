package usecase

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadFile(t *testing.T) {
	// Setup
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)

	// Test data
	fileContent := []byte("test file content")
	reader := bytes.NewReader(fileContent)
	filename := "test.txt"

	// Expectations
	mockFileRepo.On("Upload", mock.Anything, mock.Anything, filename).Return("test-file-id", nil)

	// Execute
	fileID, err := uc.UploadFile(context.Background(), reader, filename)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test-file-id", fileID)

	mockFileRepo.AssertExpectations(t)
}

func TestUploadFileTooLarge(t *testing.T) {
	// Setup
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)

	// Test data
	fileContent := make([]byte, MaxFileSize+1)
	reader := bytes.NewReader(fileContent)
	filename := "test.txt"

	// Execute
	_, err := uc.UploadFile(context.Background(), reader, filename)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrFileTooLarge, err)
	mockFileRepo.AssertNotCalled(t, "Upload")
}

func TestUploadFileInvalidType(t *testing.T) {
	// Setup
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)

	// Test data
	fileContent := []byte("test file content")
	reader := bytes.NewReader(fileContent)
	filename := "test.exe"

	// Execute
	_, err := uc.UploadFile(context.Background(), reader, filename)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFileType, err)
	mockFileRepo.AssertNotCalled(t, "Upload")
}

func TestDownloadFile(t *testing.T) {
	// Setup
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)

	// Test data
	fileID := "test-file-id"
	fileContent := []byte("test file content")
	reader := io.NopCloser(bytes.NewReader(fileContent))

	// Expectations
	mockFileRepo.On("Exists", mock.Anything, fileID).Return(true, nil)
	mockFileRepo.On("Download", mock.Anything, fileID).Return(reader, nil)

	// Execute
	rc, err := uc.DownloadFile(context.Background(), fileID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, rc)

	// Read content
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	assert.Equal(t, fileContent, buf.Bytes())

	mockFileRepo.AssertExpectations(t)
}

func TestDeleteFile(t *testing.T) {
	// Setup
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)

	// Test data
	fileID := "test-file-id"

	// Expectations
	mockFileRepo.On("Exists", mock.Anything, fileID).Return(true, nil)
	mockFileRepo.On("Delete", mock.Anything, fileID).Return(nil)

	// Execute
	err := uc.DeleteFile(context.Background(), fileID)

	// Assert
	assert.NoError(t, err)

	mockFileRepo.AssertExpectations(t)
}
