package filestoringservice

import (
	"fileanalysisservice/internal/infrastructure/config"
	"fmt"
	"io"
	"net/http"
)

type FileStoringService struct {
	basePath string
}

func NewFileStoringService(cfg *config.Config) *FileStoringService {
	return &FileStoringService{
		basePath: cfg.FileStoringServiceBaseURL,
	}
}

func (fileStoringService *FileStoringService) GetFileContent(id string) (string, error) {
	res, err := http.Get(fileStoringService.basePath + "/files/" + id + "/download")
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body")
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
