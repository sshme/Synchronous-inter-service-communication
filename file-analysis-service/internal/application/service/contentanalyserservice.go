package service

import (
	"context"
	"fileanalysisservice/internal/infrastructure/filestoringservice"
	"fileanalysisservice/internal/infrastructure/quickchart"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"fileanalysisservice/internal/domain/analysis"
	"fileanalysisservice/internal/domain/plagiarism"
	"fileanalysisservice/internal/infrastructure/storage/s3"
	"fileanalysisservice/internal/interfaces/repository"
)

// ContentAnalyserService handles content-analysis-related business logic
type ContentAnalyserService struct {
	analysisRepository repository.AnalysisRepository
	shingleRepository  repository.ShingleRepository
	fileStoringService *filestoringservice.FileStoringService
	quickChartService  *quickchart.QuickChart
	fileStorage        *s3.FileStorage
	plagiarismService  *plagiarism.PlagiarismService
}

// NewContentAnalyserService creates a new analysis service
func NewContentAnalyserService(analysisRepository repository.AnalysisRepository, shingleRepository repository.ShingleRepository, fileStoringService *filestoringservice.FileStoringService, quickChartService *quickchart.QuickChart, storage *s3.FileStorage) *ContentAnalyserService {
	return &ContentAnalyserService{
		analysisRepository: analysisRepository,
		shingleRepository:  shingleRepository,
		fileStoringService: fileStoringService,
		quickChartService:  quickChartService,
		fileStorage:        storage,
		plagiarismService:  plagiarism.NewPlagiarismService(analysisRepository, shingleRepository),
	}
}

func (s *ContentAnalyserService) Analyse(ctx context.Context, id string) (*analysis.Analysis, error) {
	existingAnalysis, err := s.analysisRepository.FindByID(ctx, id)
	if err == nil && existingAnalysis != nil {
		log.Printf("Found existing analysis with id %s", id)
		return existingAnalysis, nil
	}

	analysisModel, err := analysis.NewAnalysis(id)
	if err != nil {
		return nil, err
	}

	content, err := s.fileStoringService.GetFileContent(id)
	if err != nil {
		return nil, err
	}

	log.Printf("Starting plagiarism analysis for file %s", id)
	plagiarismReport, err := s.plagiarismService.AnalyzePlagiarism(ctx, content, id)
	if err != nil {
		log.Printf("Failed to analyze plagiarism for file %s: %v", id, err)
	} else {
		err = analysisModel.SetPlagiarismReport(plagiarismReport)
		if err != nil {
			log.Printf("Failed to set plagiarism report for file %s: %v", id, err)
		}
	}

	log.Printf("Calculating text statistics for file %s", id)
	textStats := s.plagiarismService.CalculateTextStatistics(content)
	err = analysisModel.SetStatistics(textStats)
	if err != nil {
		log.Printf("Failed to set text statistics for file %s: %v", id, err)
	}

	path, err := s.quickChartService.WordCloud(content)

	if path != "" {
		defer func(path string) {
			err := os.Remove(path)
			if err != nil {
				log.Printf("failed to remove temp file: %v", err)
			}
		}(path)
	}

	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	fileInfo, err := s.fileStorage.Upload(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	analysisModel.ID = fileInfo.ID
	analysisModel.ImageLocation = fileInfo.Location
	analysisModel.UpdatedAt = time.Now()

	err = s.analysisRepository.Store(ctx, analysisModel)
	if err != nil {
		return nil, fmt.Errorf("failed to store analysis metadata: %w", err)
	}

	return analysisModel, nil
}

// DownloadImage retrieves an analysis's image from storage
func (s *ContentAnalyserService) DownloadImage(ctx context.Context, id string) (io.ReadCloser, *analysis.Analysis, error) {
	analysisModel, err := s.analysisRepository.FindByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get analysis metadata: %w", err)
	}

	if analysisModel == nil {
		return nil, nil, fmt.Errorf("file not found")
	}

	fileReader, err := s.fileStorage.Download(ctx, analysisModel.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file from storage: %w", err)
	}

	return fileReader, analysisModel, nil
}

// AnalyzePlagiarism performs plagiarism analysis on a specific file
func (s *ContentAnalyserService) AnalyzePlagiarism(ctx context.Context, id string) (*analysis.PlagiarismReport, error) {
	content, err := s.fileStoringService.GetFileContent(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}

	return s.plagiarismService.AnalyzePlagiarism(ctx, content, id)
}
