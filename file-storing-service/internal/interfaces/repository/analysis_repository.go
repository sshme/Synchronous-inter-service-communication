package repository

import (
	"context"

	"filestoringservice/internal/domain/file"
)

// FileRepository defines the interface for file persistence operations
type FileRepository interface {
	// Store saves a file to the repository
	Store(ctx context.Context, file *file.File) error

	// FindByID retrieves a file by its ID
	FindByID(ctx context.Context, id string) (*file.File, error)

	// FindByHash retrieves a file by its content hash
	FindByHash(ctx context.Context, hash string) (*file.File, error)

	// FindAll retrieves all files from the repository
	FindAll(ctx context.Context) ([]*file.File, error)
}
