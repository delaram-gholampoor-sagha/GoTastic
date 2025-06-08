package s3

import (
	"context"
	"io"
	"path/filepath"

	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/google/uuid"
)

// FileRepository implements repository.FileRepository for S3
type FileRepository struct {
	client     *s3.Client
	bucketName string
}

// NewFileRepository creates a new S3 file repository
func NewFileRepository(client *s3.Client, bucketName string) repository.FileRepository {
	return &FileRepository{
		client:     client,
		bucketName: bucketName,
	}
}

// Upload stores a file in S3
func (r *FileRepository) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	// Generate a unique file ID
	fileID := uuid.New().String() + filepath.Ext(filename)

	// Upload to S3
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(fileID),
		Body:   file,
	})

	if err != nil {
		return "", err
	}

	return fileID, nil
}

// Download retrieves a file from S3
func (r *FileRepository) Download(ctx context.Context, fileID string) (io.ReadCloser, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(fileID),
	})

	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

// Delete removes a file from S3
func (r *FileRepository) Delete(ctx context.Context, fileID string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(fileID),
	})

	return err
}

// Exists checks if a file exists in S3
func (r *FileRepository) Exists(ctx context.Context, fileID string) (bool, error) {
	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(fileID),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
