package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/pkg/logger"
)

var (
	ErrFileTooLarge    = errors.New("file too large")
	ErrInvalidFileType = errors.New("invalid file type")
	ErrFileNotFound    = errors.New("file not found")
)

const (
	// MaxFileSize is the maximum allowed file size (10MB)
	MaxFileSize = 10 << 20

	// AllowedFileTypes is a list of allowed file extensions
	AllowedFileTypes = ".jpg,.jpeg,.png,.gif,.pdf,.txt,.doc,.docx"
)

func isAllowedFileType(ext string) bool {
	allowed := map[string]bool{
		".txt": true,
		".pdf": true,
		".jpg": true,
		".png": true,
	}
	return allowed[ext]
}

// FileUseCase handles file operations
type FileUseCase struct {
	logger   logger.Logger
	fileRepo repository.FileRepository
}

// NewFileUseCase creates a new file usecase
func NewFileUseCase(logger logger.Logger, fileRepo repository.FileRepository) *FileUseCase {
	return &FileUseCase{
		logger:   logger,
		fileRepo: fileRepo,
	}
}

// UploadFile uploads a file
func (u *FileUseCase) UploadFile(ctx context.Context, reader io.Reader, filename string) (string, error) {
	// Check file type
	ext := filepath.Ext(filename)
	if !isAllowedFileType(ext) {
		return "", ErrInvalidFileType
	}

	// Check file size
	content, err := io.ReadAll(reader)
	if err != nil {
		u.logger.Error("Failed to read file content", err)
		return "", err
	}
	if len(content) > MaxFileSize {
		return "", ErrFileTooLarge
	}

	// Create a new reader from the content
	contentReader := bytes.NewReader(content)

	fileID, err := u.fileRepo.Upload(ctx, contentReader, filename)
	if err != nil {
		u.logger.Error("Failed to upload file", err)
		return "", err
	}
	return fileID, nil
}

// DownloadFile downloads a file
func (u *FileUseCase) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	// Check if file exists
	exists, err := u.fileRepo.Exists(ctx, fileID)
	if err != nil {
		u.logger.Error("Failed to check file existence", err)
		return nil, err
	}
	if !exists {
		return nil, ErrFileNotFound
	}

	reader, err := u.fileRepo.Download(ctx, fileID)
	if err != nil {
		u.logger.Error("Failed to download file", err)
		return nil, err
	}
	return reader, nil
}

// DeleteFile deletes a file
func (u *FileUseCase) DeleteFile(ctx context.Context, fileID string) error {
	// Check if file exists
	exists, err := u.fileRepo.Exists(ctx, fileID)
	if err != nil {
		u.logger.Error("Failed to check file existence", err)
		return err
	}
	if !exists {
		return ErrFileNotFound
	}

	if err := u.fileRepo.Delete(ctx, fileID); err != nil {
		u.logger.Error("Failed to delete file", err)
		return err
	}
	return nil
}

// FileExists checks if a file exists
func (u *FileUseCase) FileExists(ctx context.Context, fileID string) (bool, error) {
	exists, err := u.fileRepo.Exists(ctx, fileID)
	if err != nil {
		u.logger.Error("Failed to check file existence", err)
		return false, err
	}
	return exists, nil
}
