package repository

import (
	"context"
)

type ShingleMatch struct {
	FileID      string
	ShingleHash string
	ShingleText string
	StartPos    int
	EndPos      int
}

type ShingleRepository interface {
	StoreShingles(ctx context.Context, fileID string, shingles []ShingleData) error
	FindMatchingShingles(ctx context.Context, hashes []string, excludeFileID string) ([]ShingleMatch, error)
	DeleteShingles(ctx context.Context, fileID string) error
}

type ShingleData struct {
	Hash     string
	Text     string
	StartPos int
	EndPos   int
}
