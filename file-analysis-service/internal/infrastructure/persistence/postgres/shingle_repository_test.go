package postgres

import (
	"context"
	"database/sql"
	"testing"

	"fileanalysisservice/internal/interfaces/repository"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	query := `
		CREATE TABLE shingles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id TEXT NOT NULL,
			shingle_hash TEXT NOT NULL,
			shingle_text TEXT NOT NULL,
			position_start INTEGER NOT NULL,
			position_end INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err = db.Exec(query)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	indexQueries := []string{
		`CREATE INDEX idx_shingle_hash ON shingles(shingle_hash)`,
		`CREATE INDEX idx_file_id ON shingles(file_id)`,
	}

	for _, indexQuery := range indexQueries {
		_, err = db.Exec(indexQuery)
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}
	}

	return db
}

func TestShingleRepository_StoreShingles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewShingleRepository(db)
	ctx := context.Background()

	shingles := []repository.ShingleData{
		{
			Hash:     "hash1",
			Text:     "первый тестовый шингл",
			StartPos: 0,
			EndPos:   20,
		},
		{
			Hash:     "hash2",
			Text:     "второй тестовый шингл",
			StartPos: 15,
			EndPos:   35,
		},
	}

	err := repo.StoreShingles(ctx, "file1", shingles)
	if err != nil {
		t.Errorf("StoreShingles() error = %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM shingles WHERE file_id = ?", "file1").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count shingles: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 shingles, got %d", count)
	}
}

func TestShingleRepository_FindMatchingShingles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewShingleRepository(db)
	ctx := context.Background()

	testData := []struct {
		fileID string
		hash   string
		text   string
	}{
		{"file1", "hash1", "первый шингл"},
		{"file1", "hash2", "второй шингл"},
		{"file2", "hash1", "совпавший шингл"},
		{"file2", "hash3", "уникальный шингл"},
	}

	for _, data := range testData {
		_, err := db.Exec(
			"INSERT INTO shingles (file_id, shingle_hash, shingle_text, position_start, position_end) VALUES (?, ?, ?, ?, ?)",
			data.fileID, data.hash, data.text, 0, 20,
		)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	hashes := []string{"hash1", "hash2", "hash4"}
	matches, err := repo.FindMatchingShingles(ctx, hashes, "file3")

	if err != nil {
		t.Errorf("FindMatchingShingles() error = %v", err)
	}

	expectedMatches := 3
	if len(matches) != expectedMatches {
		t.Errorf("Expected %d matches, got %d", expectedMatches, len(matches))
	}

	for _, match := range matches {
		if match.FileID == "file3" {
			t.Error("Found match from excluded file")
		}
	}
}

func TestShingleRepository_DeleteShingles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewShingleRepository(db)
	ctx := context.Background()

	_, err := db.Exec(
		"INSERT INTO shingles (file_id, shingle_hash, shingle_text, position_start, position_end) VALUES (?, ?, ?, ?, ?)",
		"file1", "hash1", "тестовый шингл", 0, 20,
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	err = repo.DeleteShingles(ctx, "file1")
	if err != nil {
		t.Errorf("DeleteShingles() error = %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM shingles WHERE file_id = ?", "file1").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count shingles: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 shingles after deletion, got %d", count)
	}
}

func TestShingleRepository_StoreShingles_ReplaceExisting(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewShingleRepository(db)
	ctx := context.Background()

	shingles1 := []repository.ShingleData{
		{Hash: "hash1", Text: "первый шингл", StartPos: 0, EndPos: 15},
	}
	err := repo.StoreShingles(ctx, "file1", shingles1)
	if err != nil {
		t.Errorf("First StoreShingles() error = %v", err)
	}

	shingles2 := []repository.ShingleData{
		{Hash: "hash2", Text: "новый шингл", StartPos: 0, EndPos: 12},
		{Hash: "hash3", Text: "еще один шингл", StartPos: 10, EndPos: 25},
	}
	err = repo.StoreShingles(ctx, "file1", shingles2)
	if err != nil {
		t.Errorf("Second StoreShingles() error = %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM shingles WHERE file_id = ?", "file1").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count shingles: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 shingles after replacement, got %d", count)
	}

	var oldCount int
	err = db.QueryRow("SELECT COUNT(*) FROM shingles WHERE file_id = ? AND shingle_hash = ?", "file1", "hash1").Scan(&oldCount)
	if err != nil {
		t.Errorf("Failed to count old shingles: %v", err)
	}

	if oldCount != 0 {
		t.Errorf("Expected old shingle to be deleted, but found %d", oldCount)
	}
}

func TestShingleRepository_StoreShingles_EmptySlice(t *testing.T) {
	// Тест проверяет, что пустой слайс шинглов не вызывает ошибок
	repo := &ShingleRepository{db: nil} // db не используется для пустого слайса
	ctx := context.Background()

	err := repo.StoreShingles(ctx, "file1", []repository.ShingleData{})
	if err != nil {
		t.Errorf("StoreShingles() with empty slice should not error, got: %v", err)
	}
}

func TestShingleRepository_FindMatchingShingles_EmptyHashes(t *testing.T) {
	// Тест проверяет, что пустой слайс хешей возвращает пустой результат
	repo := &ShingleRepository{db: nil} // db не используется для пустого слайса
	ctx := context.Background()

	matches, err := repo.FindMatchingShingles(ctx, []string{}, "file1")
	if err != nil {
		t.Errorf("FindMatchingShingles() with empty hashes should not error, got: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("Expected empty matches, got %d", len(matches))
	}
}

func TestShingleRepository_ValidateShingleData(t *testing.T) {
	// Тест проверяет валидацию данных шинглов
	testCases := []struct {
		name     string
		shingles []repository.ShingleData
		valid    bool
	}{
		{
			name: "Valid shingles",
			shingles: []repository.ShingleData{
				{Hash: "hash1", Text: "text1", StartPos: 0, EndPos: 10},
				{Hash: "hash2", Text: "text2", StartPos: 5, EndPos: 15},
			},
			valid: true,
		},
		{
			name: "Empty hash",
			shingles: []repository.ShingleData{
				{Hash: "", Text: "text1", StartPos: 0, EndPos: 10},
			},
			valid: false,
		},
		{
			name: "Empty text",
			shingles: []repository.ShingleData{
				{Hash: "hash1", Text: "", StartPos: 0, EndPos: 10},
			},
			valid: false,
		},
		{
			name: "Invalid positions",
			shingles: []repository.ShingleData{
				{Hash: "hash1", Text: "text1", StartPos: 10, EndPos: 5},
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := validateShingleData(tc.shingles)
			if valid != tc.valid {
				t.Errorf("Expected validation result %v, got %v", tc.valid, valid)
			}
		})
	}
}

// validateShingleData проверяет корректность данных шинглов
func validateShingleData(shingles []repository.ShingleData) bool {
	for _, shingle := range shingles {
		if shingle.Hash == "" || shingle.Text == "" {
			return false
		}
		if shingle.StartPos < 0 || shingle.EndPos < 0 || shingle.StartPos >= shingle.EndPos {
			return false
		}
	}
	return true
}
