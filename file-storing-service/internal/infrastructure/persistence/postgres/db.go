package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"filestoringservice/internal/infrastructure/config"
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

// createTables creates the necessary tables if they don't exist
func createTables(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS files (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
		    hash VARCHAR(255) NOT NULL,
			size BIGINT NOT NULL,
			content_type VARCHAR(255) NOT NULL,
			location TEXT NOT NULL,
			uploaded_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create files table: %w", err)
	}

	return nil
}
