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
	MaxFileSize = 10 << 20

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


type FileUseCase struct {
	logger   logger.Logger
	fileRepo repository.FileRepository
}


func NewFileUseCase(logger logger.Logger, fileRepo repository.FileRepository) *FileUseCase {
	return &FileUseCase{
		logger:   logger,
		fileRepo: fileRepo,
	}
}


func (u *FileUseCase) UploadFile(ctx context.Context, reader io.Reader, filename string) (string, error) {

	ext := filepath.Ext(filename)
	if !isAllowedFileType(ext) {
		return "", ErrInvalidFileType
	}


	content, err := io.ReadAll(reader)
	if err != nil {
		u.logger.Error("Failed to read file content", err)
		return "", err
	}
	if len(content) > MaxFileSize {
		return "", ErrFileTooLarge
	}


	contentReader := bytes.NewReader(content)

	fileID, err := u.fileRepo.Upload(ctx, contentReader, filename)
	if err != nil {
		u.logger.Error("Failed to upload file", err)
		return "", err
	}
	return fileID, nil
}


func (u *FileUseCase) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {

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


func (u *FileUseCase) DeleteFile(ctx context.Context, fileID string) error {

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

		
func (u *FileUseCase) FileExists(ctx context.Context, fileID string) (bool, error) {
	exists, err := u.fileRepo.Exists(ctx, fileID)
	if err != nil {
		u.logger.Error("Failed to check file existence", err)
		return false, err
	}
	return exists, nil
}
