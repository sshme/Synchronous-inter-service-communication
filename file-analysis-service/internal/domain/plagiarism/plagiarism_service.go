package plagiarism

import (
	"context"
	"fmt"
	"log"
	"time"

	"fileanalysisservice/internal/domain/analysis"
	"fileanalysisservice/internal/interfaces/repository"
)

// Service handles plagiarism detection logic
type Service struct {
	textProcessor      *TextProcessor
	analysisRepository repository.AnalysisRepository
	shingleRepository  repository.ShingleRepository
	shingleSize        int
}

// NewPlagiarismService creates a new plagiarism service
func NewPlagiarismService(analysisRepository repository.AnalysisRepository, shingleRepository repository.ShingleRepository) *Service {
	return &Service{
		textProcessor:      NewTextProcessor(),
		analysisRepository: analysisRepository,
		shingleRepository:  shingleRepository,
		shingleSize:        4,
	}
}

// AnalyzePlagiarism performs plagiarism analysis on the given text
func (ps *Service) AnalyzePlagiarism(ctx context.Context, text string, currentFileID string) (*analysis.PlagiarismReport, error) {
	log.Printf("Starting plagiarism analysis for file %s", currentFileID)

	processedText := ps.textProcessor.ProcessText(text)
	if processedText == "" {
		return &analysis.PlagiarismReport{
			UniquenessPercentage: 100.0,
			TotalShingles:        0,
			UniqueShingles:       0,
			Matches:              []analysis.PlagiarismMatch{},
			ProcessedAt:          time.Now(),
		}, nil
	}

	shingles := ps.textProcessor.GenerateShingles(processedText, ps.shingleSize)
	if len(shingles) == 0 {
		return &analysis.PlagiarismReport{
			UniquenessPercentage: 100.0,
			TotalShingles:        0,
			UniqueShingles:       0,
			Matches:              []analysis.PlagiarismMatch{},
			ProcessedAt:          time.Now(),
		}, nil
	}

	currentHashes := ps.textProcessor.HashShingles(shingles)
	currentHashSet := make(map[string]bool)
	for _, hash := range currentHashes {
		currentHashSet[hash] = true
	}

	err := ps.storeShingles(ctx, currentFileID, shingles, currentHashes)
	if err != nil {
		log.Printf("Failed to store shingles for file %s: %v", currentFileID, err)
	}

	matches, err := ps.findMatches(ctx, currentHashes, currentFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find matches: %w", err)
	}

	uniqueShingles := ps.calculateUniqueShingles(currentHashSet, matches)
	uniquenessPercentage := float64(uniqueShingles) / float64(len(shingles)) * 100

	report := &analysis.PlagiarismReport{
		UniquenessPercentage: uniquenessPercentage,
		TotalShingles:        len(shingles),
		UniqueShingles:       uniqueShingles,
		Matches:              matches,
		ProcessedAt:          time.Now(),
	}

	log.Printf("Plagiarism analysis completed for file %s: %.2f%% unique", currentFileID, uniquenessPercentage)
	return report, nil
}

// storeShingles stores shingles for the current file
func (ps *Service) storeShingles(ctx context.Context, fileID string, shingles []string, hashes []string) error {
	if len(shingles) != len(hashes) {
		return fmt.Errorf("shingles and hashes length mismatch")
	}

	shingleData := make([]repository.ShingleData, len(shingles))
	for i, shingle := range shingles {
		shingleData[i] = repository.ShingleData{
			Hash:     hashes[i],
			Text:     shingle,
			StartPos: i * ps.shingleSize,
			EndPos:   (i * ps.shingleSize) + len(shingle),
		}
	}

	return ps.shingleRepository.StoreShingles(ctx, fileID, shingleData)
}

// findMatches searches for plagiarism matches in the database
func (ps *Service) findMatches(ctx context.Context, currentHashes []string, currentFileID string) ([]analysis.PlagiarismMatch, error) {
	dbMatches, err := ps.shingleRepository.FindMatchingShingles(ctx, currentHashes, currentFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query database for matches: %w", err)
	}

	if len(dbMatches) == 0 {
		return []analysis.PlagiarismMatch{}, nil
	}

	// Group matches by file ID and calculate similarity
	fileMatches := make(map[string][]repository.ShingleMatch)
	for _, match := range dbMatches {
		fileMatches[match.FileID] = append(fileMatches[match.FileID], match)
	}

	var matches []analysis.PlagiarismMatch
	totalHashes := len(currentHashes)

	for fileID, fileShingles := range fileMatches {
		matchedHashes := len(fileShingles)
		similarity := float64(matchedHashes) / float64(totalHashes) * 100

		if similarity >= 5.0 {
			startPos := fileShingles[0].StartPos
			endPos := fileShingles[0].EndPos
			matchedTexts := []string{fileShingles[0].ShingleText}

			for _, shingle := range fileShingles[1:] {
				if shingle.StartPos < startPos {
					startPos = shingle.StartPos
				}
				if shingle.EndPos > endPos {
					endPos = shingle.EndPos
				}
				matchedTexts = append(matchedTexts, shingle.ShingleText)
			}

			matchedText := matchedTexts[0]
			if len(matchedTexts) > 1 {
				matchedText += " ... " + matchedTexts[len(matchedTexts)-1]
			}

			match := analysis.PlagiarismMatch{
				Source:      fmt.Sprintf("Документ %s", fileID),
				Similarity:  similarity,
				MatchedText: matchedText,
				StartPos:    startPos,
				EndPos:      endPos,
			}

			matches = append(matches, match)
		}
	}

	return matches, nil
}

// calculateUniqueShingles calculates the number of unique shingles
func (ps *Service) calculateUniqueShingles(currentHashes map[string]bool, matches []analysis.PlagiarismMatch) int {
	totalHashes := len(currentHashes)

	matchedHashes := 0
	for _, match := range matches {
		estimatedMatched := int(float64(totalHashes) * match.Similarity / 100.0)
		matchedHashes += estimatedMatched
	}

	if matchedHashes > totalHashes {
		matchedHashes = totalHashes
	}

	uniqueHashes := totalHashes - matchedHashes
	if uniqueHashes < 0 {
		uniqueHashes = 0
	}

	return uniqueHashes
}

// CalculateTextStatistics calculates text statistics
func (ps *Service) CalculateTextStatistics(text string) *analysis.TextStatistics {
	stats := ps.textProcessor.CalculateTextStatistics(text)
	return &analysis.TextStatistics{
		ParagraphCount: stats.ParagraphCount,
		WordCount:      stats.WordCount,
		CharacterCount: stats.CharacterCount,
		SentenceCount:  stats.SentenceCount,
	}
}

// SetShingleSize sets the size of shingles (n-grams) for analysis
func (ps *Service) SetShingleSize(size int) {
	if size > 0 {
		ps.shingleSize = size
	}
}
