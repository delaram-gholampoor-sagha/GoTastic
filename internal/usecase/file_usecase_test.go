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

	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)


	fileContent := []byte("test file content")
	reader := bytes.NewReader(fileContent)
	filename := "test.txt"


	mockFileRepo.On("Upload", mock.Anything, mock.Anything, filename).Return("test-file-id", nil)


	fileID, err := uc.UploadFile(context.Background(), reader, filename)


	assert.NoError(t, err)
	assert.Equal(t, "test-file-id", fileID)

	mockFileRepo.AssertExpectations(t)
}

func TestUploadFileTooLarge(t *testing.T) {

	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)


	fileContent := make([]byte, MaxFileSize+1)
	reader := bytes.NewReader(fileContent)
	filename := "test.txt"


	_, err := uc.UploadFile(context.Background(), reader, filename)


	assert.Error(t, err)
	assert.Equal(t, ErrFileTooLarge, err)
	mockFileRepo.AssertNotCalled(t, "Upload")
}

func TestUploadFileInvalidType(t *testing.T) {

	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)


	fileContent := []byte("test file content")
	reader := bytes.NewReader(fileContent)
	filename := "test.exe"


	_, err := uc.UploadFile(context.Background(), reader, filename)


	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFileType, err)
	mockFileRepo.AssertNotCalled(t, "Upload")
}

func TestDownloadFile(t *testing.T) {

	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)

	
	fileID := "test-file-id"
	fileContent := []byte("test file content")
	reader := io.NopCloser(bytes.NewReader(fileContent))

	
	mockFileRepo.On("Exists", mock.Anything, fileID).Return(true, nil)
	mockFileRepo.On("Download", mock.Anything, fileID).Return(reader, nil)

	
	rc, err := uc.DownloadFile(context.Background(), fileID)

	
	assert.NoError(t, err)
	assert.NotNil(t, rc)

	
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	assert.Equal(t, fileContent, buf.Bytes())

	mockFileRepo.AssertExpectations(t)
}

func TestDeleteFile(t *testing.T) {
	
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockFileRepo := new(MockFileRepository)
	uc := NewFileUseCase(log, mockFileRepo)

	
	fileID := "test-file-id"

	
	mockFileRepo.On("Exists", mock.Anything, fileID).Return(true, nil)
	mockFileRepo.On("Delete", mock.Anything, fileID).Return(nil)

			
	err := uc.DeleteFile(context.Background(), fileID)

	
	assert.NoError(t, err)

	mockFileRepo.AssertExpectations(t)
}
