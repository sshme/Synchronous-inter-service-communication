package repository

import (
	"context"

	"filestoringservice/internal/domain/file"
)

// FileRepository defines the interface for file persistence operations
type FileRepository interface {
	Store(ctx context.Context, file *file.File) error
	FindByID(ctx context.Context, id string) (*file.File, error)
	FindByHash(ctx context.Context, hash string) (*file.File, error)
	FindAll(ctx context.Context) ([]*file.File, error)
}
