package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"fileanalysisservice/internal/interfaces/repository"
)

// ShingleRepository implements the repository.ShingleRepository interface with PostgreSQL
type ShingleRepository struct {
	db *sql.DB
}

// NewShingleRepository creates a new PostgreSQL shingle repository
func NewShingleRepository(db *sql.DB) *ShingleRepository {
	return &ShingleRepository{
		db: db,
	}
}

// StoreShingles stores shingles for a file
func (r *ShingleRepository) StoreShingles(ctx context.Context, fileID string, shingles []repository.ShingleData) error {
	if len(shingles) == 0 {
		return nil
	}

	err := r.DeleteShingles(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete existing shingles: %w", err)
	}

	valueStrings := make([]string, 0, len(shingles))
	valueArgs := make([]interface{}, 0, len(shingles)*5)

	for i, shingle := range shingles {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
			i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		valueArgs = append(valueArgs, fileID, shingle.Hash, shingle.Text, shingle.StartPos, shingle.EndPos)
	}

	query := fmt.Sprintf(`
		INSERT INTO shingles (file_id, shingle_hash, shingle_text, position_start, position_end)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err = r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to store shingles: %w", err)
	}

	return nil
}

// FindMatchingShingles finds shingles that match the given hashes
func (r *ShingleRepository) FindMatchingShingles(ctx context.Context, hashes []string, excludeFileID string) ([]repository.ShingleMatch, error) {
	if len(hashes) == 0 {
		return []repository.ShingleMatch{}, nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(hashes))
	args := make([]interface{}, len(hashes)+1)

	for i, hash := range hashes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = hash
	}
	args[len(hashes)] = excludeFileID

	query := fmt.Sprintf(`
		SELECT file_id, shingle_hash, shingle_text, position_start, position_end
		FROM shingles
		WHERE shingle_hash IN (%s) AND file_id != $%d
		ORDER BY file_id, position_start
	`, strings.Join(placeholders, ","), len(hashes)+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query matching shingles: %w", err)
	}
	defer rows.Close()

	var matches []repository.ShingleMatch
	for rows.Next() {
		var match repository.ShingleMatch
		err := rows.Scan(
			&match.FileID,
			&match.ShingleHash,
			&match.ShingleText,
			&match.StartPos,
			&match.EndPos,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shingle match: %w", err)
		}
		matches = append(matches, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shingle matches: %w", err)
	}

	return matches, nil
}

// DeleteShingles removes all shingles for a file
func (r *ShingleRepository) DeleteShingles(ctx context.Context, fileID string) error {
	query := `DELETE FROM shingles WHERE file_id = $1`

	_, err := r.db.ExecContext(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete shingles: %w", err)
	}

	return nil
}
