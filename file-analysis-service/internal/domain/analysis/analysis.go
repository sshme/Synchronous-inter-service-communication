package analysis

import (
	"encoding/json"
	"errors"
	"time"
)

// PlagiarismMatch represents a single plagiarism match
type PlagiarismMatch struct {
	Source      string  `json:"source"`     // URL или название источника
	Similarity  float64 `json:"similarity"` // Процент схожести (0-100)
	MatchedText string  `json:"matched_text"`
	StartPos    int     `json:"start_pos"`
	EndPos      int     `json:"end_pos"`
}

// PlagiarismReport represents the plagiarism analysis report
type PlagiarismReport struct {
	UniquenessPercentage float64           `json:"uniqueness_percentage"` // Процент уникальности
	TotalShingles        int               `json:"total_shingles"`        // Общее количество шинглов
	UniqueShingles       int               `json:"unique_shingles"`       // Количество уникальных шинглов
	Matches              []PlagiarismMatch `json:"matches"`               // Найденные совпадения
	ProcessedAt          time.Time         `json:"processed_at"`          // Время обработки
}

// TextStatistics represents text analysis statistics
type TextStatistics struct {
	ParagraphCount int `json:"paragraph_count"`
	WordCount      int `json:"word_count"`
	CharacterCount int `json:"character_count"`
	SentenceCount  int `json:"sentence_count"`
}

// Analysis represents an analysis entity in the domain
type Analysis struct {
	ID               string            `json:"id"`
	FileID           string            `json:"file_id"`
	ImageLocation    string            `json:"image_location"`
	PlagiarismReport *PlagiarismReport `json:"plagiarism_report,omitempty"` // JSON отчет об антиплагиате
	Statistics       *TextStatistics   `json:"statistics,omitempty"`        // JSON статистика текста
	UpdatedAt        time.Time         `json:"updated_at"`
	CreatedAt        time.Time         `json:"created_at"`
}

// NewAnalysis creates a new analysis domain entity
func NewAnalysis(fileID string) (*Analysis, error) {
	now := time.Now()

	return &Analysis{
		FileID:    fileID,
		UpdatedAt: now,
		CreatedAt: now,
	}, nil
}

// SetImageLocation sets the analysis image location (words cloud)
func (a *Analysis) SetImageLocation(location string) error {
	if location == "" {
		return errors.New("location cannot be empty")
	}
	a.ImageLocation = location
	a.UpdatedAt = time.Now()
	return nil
}

// SetPlagiarismReport sets the plagiarism analysis report
func (a *Analysis) SetPlagiarismReport(report *PlagiarismReport) error {
	if report == nil {
		return errors.New("plagiarism report cannot be nil")
	}
	a.PlagiarismReport = report
	a.UpdatedAt = time.Now()
	return nil
}

// SetStatistics sets the text statistics
func (a *Analysis) SetStatistics(stats *TextStatistics) error {
	if stats == nil {
		return errors.New("statistics cannot be nil")
	}
	a.Statistics = stats
	a.UpdatedAt = time.Now()
	return nil
}

// GetPlagiarismReportJSON returns the plagiarism report as JSON string
func (a *Analysis) GetPlagiarismReportJSON() (string, error) {
	if a.PlagiarismReport == nil {
		return "", nil
	}
	data, err := json.Marshal(a.PlagiarismReport)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetStatisticsJSON returns the statistics as JSON string
func (a *Analysis) GetStatisticsJSON() (string, error) {
	if a.Statistics == nil {
		return "", nil
	}
	data, err := json.Marshal(a.Statistics)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetPlagiarismReportFromJSON sets the plagiarism report from JSON string
func (a *Analysis) SetPlagiarismReportFromJSON(jsonData string) error {
	if jsonData == "" {
		a.PlagiarismReport = nil
		return nil
	}
	var report PlagiarismReport
	if err := json.Unmarshal([]byte(jsonData), &report); err != nil {
		return err
	}
	a.PlagiarismReport = &report
	a.UpdatedAt = time.Now()
	return nil
}

// SetStatisticsFromJSON sets the statistics from JSON string
func (a *Analysis) SetStatisticsFromJSON(jsonData string) error {
	if jsonData == "" {
		a.Statistics = nil
		return nil
	}
	var stats TextStatistics
	if err := json.Unmarshal([]byte(jsonData), &stats); err != nil {
		return err
	}
	a.Statistics = &stats
	a.UpdatedAt = time.Now()
	return nil
}
