package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"filestoringservice/internal/domain/file"
)

// FileRepository implements the repository.FileRepository interface with PostgreSQL
type FileRepository struct {
	db *sql.DB
}

// NewFileRepository creates a new PostgreSQL file repository
func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{
		db: db,
	}
}

// Store saves a file to the database
func (r *FileRepository) Store(ctx context.Context, file *file.File) error {
	query := `
		INSERT INTO files (id, name, hash, size, content_type, location, uploaded_at, updated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		file.ID,
		file.Name,
		file.Hash,
		file.Size,
		file.ContentType,
		file.Location,
		file.UploadedAt,
		file.UpdatedAt,
		file.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store file: %w", err)
	}

	return nil
}

// findBy implements universal find logic.
func (r *FileRepository) findBy(ctx context.Context, key string, value any) (*file.File, error) {
	query := fmt.Sprintf(`
		SELECT id, name, hash, size, content_type, location, uploaded_at, updated_at, created_at
		FROM files
		WHERE %s = $1
	`, key)

	row := r.db.QueryRowContext(ctx, query, value)

	var f file.File
	var uploadedAt, updatedAt, createdAt time.Time

	err := row.Scan(
		&f.ID,
		&f.Name,
		&f.Hash,
		&f.Size,
		&f.ContentType,
		&f.Location,
		&uploadedAt,
		&updatedAt,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	f.UploadedAt = uploadedAt
	f.UpdatedAt = updatedAt
	f.CreatedAt = createdAt

	return &f, nil
}

func (r *FileRepository) FindByID(ctx context.Context, id string) (*file.File, error) {
	return r.findBy(ctx, "id", id)
}

func (r *FileRepository) FindByHash(ctx context.Context, hash string) (*file.File, error) {
	return r.findBy(ctx, "hash", hash)
}

// FindAll retrieves all files from the database
func (r *FileRepository) FindAll(ctx context.Context) ([]*file.File, error) {
	query := `
		SELECT id, name, hash, size, content_type, location, uploaded_at, updated_at, created_at
		FROM files
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all files: %w", err)
	}
	defer rows.Close()

	var files []*file.File
	for rows.Next() {
		var f file.File
		var uploadedAt, updatedAt, createdAt time.Time

		err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.Hash,
			&f.Size,
			&f.ContentType,
			&f.Location,
			&uploadedAt,
			&updatedAt,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		f.UploadedAt = uploadedAt
		f.UpdatedAt = updatedAt
		f.CreatedAt = createdAt

		files = append(files, &f)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over files: %w", err)
	}

	return files, nil
}
