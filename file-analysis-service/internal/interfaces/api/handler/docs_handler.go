package handler

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

// DocsHandler handles HTTP requests related to files
type DocsHandler struct {
}

// NewDocsHandler creates a new file handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

func (h *DocsHandler) Docs(w http.ResponseWriter, r *http.Request) {
	httpSwagger.Handler(httpSwagger.URL("/analysis-api/docs/swagger.json"))(w, r)
}

func (h *DocsHandler) Swagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "./docs/swagger.json")
}
