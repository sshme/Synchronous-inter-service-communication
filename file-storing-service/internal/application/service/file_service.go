package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"filestoringservice/internal/domain/file"
	"filestoringservice/internal/infrastructure/storage/s3"
	"filestoringservice/internal/interfaces/hash"
	"filestoringservice/internal/interfaces/repository"
)

// FileService handles file-related business logic
type FileService struct {
	fileRepository repository.FileRepository
	fileStorage    *s3.FileStorage
	hasher         hash.Hasher
}

// NewFileService creates a new file service
func NewFileService(repository repository.FileRepository, storage *s3.FileStorage, hasher hash.Hasher) *FileService {
	return &FileService{
		fileRepository: repository,
		fileStorage:    storage,
		hasher:         hasher,
	}
}

// UploadFile handles file upload, stores metadata in DB and actual file in S3
func (s *FileService) UploadFile(ctx context.Context, name, contentType string, size int64, fileData io.Reader) (*file.File, error) {
	fileModel, err := file.NewFile(name, contentType, size)
	if err != nil {
		return nil, err
	}

	tempFile, err := os.CreateTemp("", "upload-*"+filepath.Ext(name))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Printf("failed to remove temp file: %w", err)
		}
	}(tempFile.Name())

	defer func(tempFile *os.File) {
		err := tempFile.Close()
		if err != nil {
			log.Printf("failed to close temp file: %w", err)
		}
	}(tempFile)

	_, err = io.Copy(tempFile, fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	fileHash, err := s.hasher.ComputeHashFromFile(ctx, tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to compute file fileHash: %w", err)
	}

	if err := fileModel.SetHash(fileHash); err != nil {
		return nil, fmt.Errorf("failed to set file fileHash: %w", err)
	}

	existingFile, err := s.fileRepository.FindByHash(ctx, fileHash)
	if err == nil && existingFile != nil {
		log.Printf("file with hash %s already exists", existingFile.Hash)
		return existingFile, nil
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	fileInfo, err := s.fileStorage.Upload(ctx, tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	fileModel.ID = fileInfo.ID
	fileModel.Location = fileInfo.Location

	err = s.fileRepository.Store(ctx, fileModel)
	if err != nil {
		return nil, fmt.Errorf("failed to store file metadata: %w", err)
	}

	return fileModel, nil
}

// GetFileByID retrieves a file by its ID
func (s *FileService) GetFileByID(ctx context.Context, id string) (*file.File, error) {
	return s.fileRepository.FindByID(ctx, id)
}

// GetAllFiles retrieves all files
func (s *FileService) GetAllFiles(ctx context.Context) ([]*file.File, error) {
	return s.fileRepository.FindAll(ctx)
}

// DownloadFile retrieves a file's content from storage
func (s *FileService) DownloadFile(ctx context.Context, id string) (io.ReadCloser, *file.File, error) {
	fileModel, err := s.fileRepository.FindByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	if fileModel == nil {
		return nil, nil, fmt.Errorf("file not found")
	}

	fileReader, err := s.fileStorage.Download(ctx, fileModel.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file from storage: %w", err)
	}

	return fileReader, fileModel, nil
}
