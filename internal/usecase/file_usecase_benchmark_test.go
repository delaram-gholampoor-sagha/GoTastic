package usecase

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/stretchr/testify/mock"
)

func BenchmarkUploadFile(b *testing.B) {

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


	b.ResetTimer()


	for i := 0; i < b.N; i++ {
		_, err := uc.UploadFile(context.Background(), reader, filename)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDownloadFile(b *testing.B) {
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockRepo := new(MockFileRepository)
	useCase := NewFileUseCase(log, mockRepo)


	mockRepo.On("Exists", mock.Anything, "test-file-id").Return(true, nil)
	mockRepo.On("Download", mock.Anything, "test-file-id").Return(io.NopCloser(bytes.NewReader([]byte("test content"))), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, err := useCase.DownloadFile(ctx, "test-file-id")
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, reader)
		reader.Close()
	}
}

func BenchmarkDeleteFile(b *testing.B) {
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})
	mockRepo := new(MockFileRepository)
	useCase := NewFileUseCase(log, mockRepo)

			
	mockRepo.On("Exists", mock.Anything, "test-file-id").Return(true, nil)
	mockRepo.On("Delete", mock.Anything, "test-file-id").Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := useCase.DeleteFile(ctx, "test-file-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}
