package repository

import (
	"context"

	"fileanalysisservice/internal/domain/analysis"
)

type AnalysisRepository interface {
	Store(ctx context.Context, file *analysis.Analysis) error
	FindByID(ctx context.Context, id string) (*analysis.Analysis, error)
}
