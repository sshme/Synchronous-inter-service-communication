package plagiarism

import (
	"context"
	"testing"

	"fileanalysisservice/internal/domain/analysis"
	"fileanalysisservice/internal/interfaces/repository"
)

// MockAnalysisRepository is a mock implementation of AnalysisRepository
type MockAnalysisRepository struct{}

func (m *MockAnalysisRepository) Store(ctx context.Context, analysis *analysis.Analysis) error {
	return nil
}

func (m *MockAnalysisRepository) FindByID(ctx context.Context, id string) (*analysis.Analysis, error) {
	return nil, nil
}

// MockShingleRepository is a mock implementation of ShingleRepository
type MockShingleRepository struct {
	storedShingles map[string][]repository.ShingleData
	matches        []repository.ShingleMatch
}

func NewMockShingleRepository() *MockShingleRepository {
	return &MockShingleRepository{
		storedShingles: make(map[string][]repository.ShingleData),
		matches:        []repository.ShingleMatch{},
	}
}

func (m *MockShingleRepository) StoreShingles(ctx context.Context, fileID string, shingles []repository.ShingleData) error {
	m.storedShingles[fileID] = shingles
	return nil
}

func (m *MockShingleRepository) FindMatchingShingles(ctx context.Context, hashes []string, excludeFileID string) ([]repository.ShingleMatch, error) {
	return m.matches, nil
}

func (m *MockShingleRepository) DeleteShingles(ctx context.Context, fileID string) error {
	delete(m.storedShingles, fileID)
	return nil
}

func (m *MockShingleRepository) SetMatches(matches []repository.ShingleMatch) {
	m.matches = matches
}

func TestPlagiarismService_AnalyzePlagiarism(t *testing.T) {
	analysisRepo := &MockAnalysisRepository{}
	shingleRepo := NewMockShingleRepository()
	service := NewPlagiarismService(analysisRepo, shingleRepo)

	tests := []struct {
		name           string
		text           string
		fileID         string
		setupMatches   func(*MockShingleRepository)
		expectedUnique float64
		expectMatches  bool
	}{
		{
			name:           "Empty text",
			text:           "",
			fileID:         "file1",
			setupMatches:   func(repo *MockShingleRepository) {},
			expectedUnique: 100.0,
			expectMatches:  false,
		},
		{
			name:           "Unique text with no matches",
			text:           "Это уникальный текст без совпадений в базе данных",
			fileID:         "file1",
			setupMatches:   func(repo *MockShingleRepository) {},
			expectedUnique: 100.0,
			expectMatches:  false,
		},
		{
			name:   "Text with matches",
			text:   "Это текст с некоторыми совпадениями в базе данных для тестирования",
			fileID: "file1",
			setupMatches: func(repo *MockShingleRepository) {
				matches := []repository.ShingleMatch{
					{
						FileID:      "file2",
						ShingleHash: "hash1",
						ShingleText: "текст некоторыми совпадениями базе",
						StartPos:    10,
						EndPos:      50,
					},
					{
						FileID:      "file2",
						ShingleHash: "hash2",
						ShingleText: "совпадениями базе данных для",
						StartPos:    30,
						EndPos:      70,
					},
				}
				repo.SetMatches(matches)
			},
			expectedUnique: 85.0,
			expectMatches:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMatches(shingleRepo)

			report, err := service.AnalyzePlagiarism(context.Background(), tt.text, tt.fileID)

			if err != nil {
				t.Errorf("AnalyzePlagiarism() error = %v", err)
				return
			}

			if report == nil {
				t.Error("AnalyzePlagiarism() returned nil report")
				return
			}

			if tt.expectMatches && len(report.Matches) == 0 {
				t.Error("Expected matches but got none")
			}

			if !tt.expectMatches && len(report.Matches) > 0 {
				t.Errorf("Expected no matches but got %d", len(report.Matches))
			}

			if tt.text != "" && report.TotalShingles == 0 {
				t.Error("Expected shingles to be generated for non-empty text")
			}

			if report.UniquenessPercentage < 0 || report.UniquenessPercentage > 100 {
				t.Errorf("UniquenessPercentage = %v, want between 0 and 100", report.UniquenessPercentage)
			}
		})
	}
}

func TestPlagiarismService_findMatches(t *testing.T) {
	analysisRepo := &MockAnalysisRepository{}
	shingleRepo := NewMockShingleRepository()
	service := NewPlagiarismService(analysisRepo, shingleRepo)

	mockMatches := []repository.ShingleMatch{
		{
			FileID:      "file2",
			ShingleHash: "hash1",
			ShingleText: "первый совпавший шингл",
			StartPos:    0,
			EndPos:      25,
		},
		{
			FileID:      "file2",
			ShingleHash: "hash2",
			ShingleText: "второй совпавший шингл",
			StartPos:    20,
			EndPos:      45,
		},
		{
			FileID:      "file3",
			ShingleHash: "hash3",
			ShingleText: "шингл из другого файла",
			StartPos:    10,
			EndPos:      35,
		},
	}
	shingleRepo.SetMatches(mockMatches)

	hashes := []string{"hash1", "hash2", "hash3", "hash4", "hash5"}
	matches, err := service.findMatches(context.Background(), hashes, "file1")

	if err != nil {
		t.Errorf("findMatches() error = %v", err)
		return
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches (one per file), got %d", len(matches))
	}

	for _, match := range matches {
		if match.Similarity < 5.0 {
			t.Errorf("Match similarity %v is below threshold", match.Similarity)
		}
		if match.Source == "" {
			t.Error("Match source should not be empty")
		}
		if match.MatchedText == "" {
			t.Error("Match text should not be empty")
		}
	}
}

func TestPlagiarismService_CalculateTextStatistics(t *testing.T) {
	analysisRepo := &MockAnalysisRepository{}
	shingleRepo := NewMockShingleRepository()
	service := NewPlagiarismService(analysisRepo, shingleRepo)

	text := `Первый абзац с несколькими предложениями. Это второе предложение!

Второй абзац. Здесь тоже есть предложения?

Третий абзац с одним предложением.`

	stats := service.CalculateTextStatistics(text)

	if stats == nil {
		t.Error("CalculateTextStatistics() returned nil")
		return
	}

	if stats.ParagraphCount != 3 {
		t.Errorf("ParagraphCount = %v, want 3", stats.ParagraphCount)
	}

	if stats.WordCount == 0 {
		t.Error("WordCount should not be 0")
	}

	if stats.CharacterCount != len(text) {
		t.Errorf("CharacterCount = %v, want %v", stats.CharacterCount, len(text))
	}

	if stats.SentenceCount == 0 {
		t.Error("SentenceCount should not be 0")
	}
}
