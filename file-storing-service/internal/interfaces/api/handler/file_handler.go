package handler

import (
	"encoding/json"
	"filestoringservice/internal/application/service"
	"filestoringservice/internal/domain/file"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// FileHandler handles HTTP requests related to files
type FileHandler struct {
	fileService *service.FileService
}

// FileResponse represents the response structure for file operations
type FileResponse struct {
	ID          string `json:"id" example:"12345678-1234-1234-1234-123456789012"`
	Name        string `json:"name" example:"document.pdf"`
	Hash        string `json:"hash"`
	Size        int64  `json:"size" example:"1048576"`
	ContentType string `json:"content_type" example:"application/pdf"`
	Location    string `json:"location" example:"files/12345678-1234-1234-1234-123456789012"`
	UploadedAt  string `json:"uploaded_at" example:"2023-01-01T12:00:00Z"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message" example:"File not found"`
	Code    int    `json:"code" example:"404"`
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// UploadFile handles file upload requests
// @Summary Upload a file
// @Description Upload a new file to the server
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 201 {object} FileResponse "File uploaded successfully"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /files [post]
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Limit formFile size using domain constant
	err := r.ParseMultipartForm(file.MaxFileSize)
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the uploaded formFile
	formFile, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request: "+err.Error(), http.StatusBadRequest)
		return
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {

		}
	}(formFile)

	// Get formFile info
	filename := header.Filename
	contentType := header.Header.Get("Content-Type")
	size := header.Size

	// Upload formFile
	fileModel, err := h.fileService.UploadFile(r.Context(), filename, contentType, size, formFile)
	if err != nil {
		http.Error(w, "Failed to upload formFile: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Return formFile info as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Response structure
	response := map[string]any{
		"id":           fileModel.ID,
		"name":         fileModel.Name,
		"hash":         fileModel.Hash,
		"size":         fileModel.Size,
		"content_type": fileModel.ContentType,
		"location":     fileModel.Location,
		"uploaded_at":  fileModel.UploadedAt,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

// GetFile handles file retrieval requests
// @Summary Get a file by ID
// @Description Get file information by its ID
// @Tags files
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} FileResponse "File information"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "File not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /files/{id} [get]
func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	fileModel, err := h.fileService.GetFileByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if fileModel == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]any{
		"id":           fileModel.ID,
		"name":         fileModel.Name,
		"size":         fileModel.Size,
		"content_type": fileModel.ContentType,
		"location":     fileModel.Location,
		"uploaded_at":  fileModel.UploadedAt,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

// GetAllFiles handles requests to retrieve all files
// @Summary Get all files
// @Description Get information for all uploaded files
// @Tags files
// @Accept json
// @Produce json
// @Success 200 {array} FileResponse "List of all files"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /files [get]
func (h *FileHandler) GetAllFiles(w http.ResponseWriter, r *http.Request) {
	// Get all files
	files, err := h.fileService.GetAllFiles(r.Context())
	if err != nil {
		http.Error(w, "Failed to get files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var responses []map[string]any
	for _, fileModel := range files {
		response := map[string]any{
			"id":           fileModel.ID,
			"name":         fileModel.Name,
			"hash":         fileModel.Hash,
			"size":         fileModel.Size,
			"content_type": fileModel.ContentType,
			"location":     fileModel.Location,
			"uploaded_at":  fileModel.UploadedAt,
		}
		responses = append(responses, response)
	}

	// Return files as JSON
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(responses)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DownloadFile handles file download requests
// @Summary Download a file by ID
// @Description Download the actual file content by its ID
// @Tags files
// @Produce application/octet-stream
// @Param id path string true "File ID"
// @Success 200 {file} binary "File content"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "File not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /files/{id}/download [get]
func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	// Download file content and get metadata
	fileReader, fileModel, err := h.fileService.DownloadFile(r.Context(), id)
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
			// Log the error but don't interrupt the response
			fmt.Printf("Failed to close file reader: %v\n", closeErr)
		}
	}()

	// Set appropriate headers for file download
	w.Header().Set("Content-Type", fileModel.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileModel.Name))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileModel.Size))

	// Stream the file content to the response
	_, err = io.Copy(w, fileReader)
	if err != nil {
		http.Error(w, "Failed to stream file content", http.StatusInternalServerError)
		return
	}
}
