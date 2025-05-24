package postgres

import (
	"database/sql"
	"fmt"

	"fileanalysisservice/internal/infrastructure/config"

	_ "github.com/lib/pq"
)

// NewDB creates a new database connection
func NewDB(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS analysis (
			id VARCHAR(255) PRIMARY KEY,
		    file_id VARCHAR(255) NOT NULL,
			image_location TEXT NOT NULL,
			plagiarism_report JSONB,
			statistics JSONB,
			updated_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create analysis table: %w", err)
	}

	shinglesQuery := `
		CREATE TABLE IF NOT EXISTS shingles (
			id SERIAL PRIMARY KEY,
			file_id VARCHAR(255) NOT NULL,
			shingle_hash VARCHAR(32) NOT NULL,
			shingle_text TEXT NOT NULL,
			position_start INTEGER NOT NULL,
			position_end INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`

	_, err = db.Exec(shinglesQuery)
	if err != nil {
		return fmt.Errorf("failed to create shingles table: %w", err)
	}

	// Создание индексов для таблицы shingles
	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_shingle_hash ON shingles(shingle_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_file_id ON shingles(file_id)`,
	}

	for _, indexQuery := range indexQueries {
		_, err = db.Exec(indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
