package handler

import (
	"encoding/json"
	"fileanalysisservice/internal/application/service"
	"fmt"
	"io"
	"net/http"
)

// AnalyseHandler handles HTTP requests related to files
type AnalyseHandler struct {
	contentAnalyserService *service.ContentAnalyserService
}

func NewAnalysisHandler(contentAnalyserService *service.ContentAnalyserService) *AnalyseHandler {
	return &AnalyseHandler{
		contentAnalyserService: contentAnalyserService,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message" example:"File not found"`
	Code    int    `json:"code" example:"404"`
}

// GetAnalyse handles the analysis retrieval endpoint
// @Summary Retrieve file analysis
// @Description Get analysis details for a specific file by its ID
// @Tags analysis
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} map[string]any "Analysis details"
// @Failure 400 {object} ErrorResponse "Bad Request - File ID is required"
// @Failure 404 {object} ErrorResponse "Not Found - Analysis not found"
// @Failure 500 {object} ErrorResponse "Internal Server Error - Failed to get analysis"
// @Router /analysis/{id} [get]
func (h *AnalyseHandler) GetAnalyse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	analysisModel, err := h.contentAnalyserService.Analyse(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get analysis: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if analysisModel == nil {
		http.Error(w, "Analysis not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]any{
		"id":                analysisModel.ID,
		"file_id":           analysisModel.FileID,
		"image_location":    analysisModel.ImageLocation,
		"plagiarism_report": analysisModel.PlagiarismReport,
		"statistics":        analysisModel.Statistics,
		"updated_at":        analysisModel.UpdatedAt,
		"created_at":        analysisModel.CreatedAt,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

// DownloadCloud handles cloud analysis download requests
// @Summary Download a cloud image by ID
// @Description Download the actual analysis cloud image by its ID
// @Tags analysis
// @Produce application/octet-stream
// @Param id path string true "Analysis ID"
// @Success 200 {analysis} binary "Analysis image"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Analysis not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /analysis/{id}/download [get]
func (h *AnalyseHandler) DownloadCloud(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	fileReader, _, err := h.contentAnalyserService.DownloadImage(r.Context(), id)
	if err != nil {
		if err.Error() == "file not found" {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to download file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if closeErr := fileReader.Close(); closeErr != nil {
			fmt.Printf("Failed to close file reader: %v\n", closeErr)
		}
	}()

	w.Header().Set("Content-Type", "image/png")

	_, err = io.Copy(w, fileReader)
	if err != nil {
		http.Error(w, "Failed to stream file content", http.StatusInternalServerError)
		return
	}
}
