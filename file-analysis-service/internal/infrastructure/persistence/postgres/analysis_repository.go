package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"fileanalysisservice/internal/domain/analysis"
)

// AnalysisRepository implements the repository.FileRepository interface with PostgreSQL
type AnalysisRepository struct {
	db *sql.DB
}

// NewAnalysisRepository creates a new PostgreSQL file repository
func NewAnalysisRepository(db *sql.DB) *AnalysisRepository {
	return &AnalysisRepository{
		db: db,
	}
}

// Store saves a file to the database
func (r *AnalysisRepository) Store(ctx context.Context, analysis *analysis.Analysis) error {
	query := `
		INSERT INTO analysis (id, file_id, image_location, plagiarism_report, statistics, updated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			file_id = EXCLUDED.file_id,
			image_location = EXCLUDED.image_location,
			plagiarism_report = EXCLUDED.plagiarism_report,
			statistics = EXCLUDED.statistics,
			updated_at = EXCLUDED.updated_at
	`

	// Convert reports to JSON strings
	var plagiarismReportJSON, statisticsJSON *string

	if reportJSON, err := analysis.GetPlagiarismReportJSON(); err != nil {
		return fmt.Errorf("failed to marshal plagiarism report: %w", err)
	} else if reportJSON != "" {
		plagiarismReportJSON = &reportJSON
	}

	if statsJSON, err := analysis.GetStatisticsJSON(); err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	} else if statsJSON != "" {
		statisticsJSON = &statsJSON
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		analysis.ID,
		analysis.FileID,
		analysis.ImageLocation,
		plagiarismReportJSON,
		statisticsJSON,
		analysis.UpdatedAt,
		analysis.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store analysis: %w", err)
	}

	return nil
}

// findBy implements universal find logic.
func (r *AnalysisRepository) findBy(ctx context.Context, key string, value any) (*analysis.Analysis, error) {
	query := fmt.Sprintf(`
		SELECT id, file_id, image_location, plagiarism_report, statistics, updated_at, created_at
		FROM analysis
		WHERE %s = $1
	`, key)

	row := r.db.QueryRowContext(ctx, query, value)

	var f analysis.Analysis
	var updatedAt, createdAt time.Time
	var plagiarismReportJSON, statisticsJSON sql.NullString

	err := row.Scan(
		&f.ID,
		&f.FileID,
		&f.ImageLocation,
		&plagiarismReportJSON,
		&statisticsJSON,
		&updatedAt,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find analysis: %w", err)
	}

	f.UpdatedAt = updatedAt
	f.CreatedAt = createdAt

	if plagiarismReportJSON.Valid {
		if err := f.SetPlagiarismReportFromJSON(plagiarismReportJSON.String); err != nil {
			return nil, fmt.Errorf("failed to parse plagiarism report JSON: %w", err)
		}
	}

	if statisticsJSON.Valid {
		if err := f.SetStatisticsFromJSON(statisticsJSON.String); err != nil {
			return nil, fmt.Errorf("failed to parse statistics JSON: %w", err)
		}
	}

	return &f, nil
}

func (r *AnalysisRepository) FindByID(ctx context.Context, id string) (*analysis.Analysis, error) {
	return r.findBy(ctx, "id", id)
}
