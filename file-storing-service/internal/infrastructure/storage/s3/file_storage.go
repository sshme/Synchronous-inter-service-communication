package s3

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"

	"filestoringservice/internal/infrastructure/config"
)

// FileStorage handles file operations with S3.
type FileStorage struct {
	client   *s3.S3
	bucket   string
	endpoint string
}

// NewFileStorage creates a new S3 file storage.
func NewFileStorage(cfg *config.Config) (*FileStorage, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.S3Region),
		Credentials:      credentials.NewStaticCredentials(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Endpoint:         aws.String(cfg.S3Endpoint),
		S3ForcePathStyle: aws.Bool(cfg.S3ForcePathStyle),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	client := s3.New(sess)

	return &FileStorage{
		client:   client,
		bucket:   cfg.S3Bucket,
		endpoint: cfg.S3Endpoint,
	}, nil
}

// UploadedFileInfo represents the information about the uploaded file.
type UploadedFileInfo struct {
	ID       string
	Location string
}

// Upload uploads a file to S3 and returns the uploaded file information.
func (s *FileStorage) Upload(_ context.Context, fileData *os.File) (*UploadedFileInfo, error) {
	fileKey := uuid.New().String()

	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
		Body:   fileData,
		ACL:    aws.String(s3.ObjectCannedACLPrivate),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return &UploadedFileInfo{
		ID:       fileKey,
		Location: fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, fileKey),
	}, nil
}

// Download downloads a file from S3 and returns a reader for the file content.
func (s *FileStorage) Download(ctx context.Context, fileKey string) (io.ReadCloser, error) {
	result, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return result.Body, nil
}
