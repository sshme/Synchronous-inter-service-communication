package quickchart

import (
	"fileanalysisservice/internal/infrastructure/config"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"os"
)

type QuickChart struct {
	basePath string
}

func NewQuickChart(cfg *config.Config) *QuickChart {
	return &QuickChart{
		basePath: cfg.WordCloudBaseURL,
	}
}

func (qc *QuickChart) WordCloud(content string) (string, error) {
	encodedContent := url.QueryEscape(content)
	apiURL := qc.basePath + fmt.Sprintf("?text=%s&format=png", encodedContent)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch word cloud: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch word cloud: status code %d", resp.StatusCode)
	}

	path := fmt.Sprintf("wordcloud_%s.png", uuid.NewString())

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return path, fmt.Errorf("failed to save word cloud image: %w", err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}

	return path, nil
}
