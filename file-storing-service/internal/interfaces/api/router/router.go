package router

import (
	"net/http"

	"filestoringservice/internal/interfaces/api/handler"

	_ "filestoringservice/docs" // Import for swagger docs
)

// Router handles HTTP routing
type Router struct {
	fileHandler *handler.FileHandler
	infoHandler *handler.InfoHandler
	docsHandler *handler.DocsHandler
}

// NewRouter creates a new router
func NewRouter(fileHandler *handler.FileHandler, infoHandler *handler.InfoHandler, docsHandler *handler.DocsHandler) *Router {
	return &Router{
		fileHandler: fileHandler,
		infoHandler: infoHandler,
		docsHandler: docsHandler,
	}
}

// SetupRoutes sets up the HTTP routes
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Info routes
	mux.HandleFunc("GET /store-api/info/health", r.infoHandler.HealthCheck)

	// File routes
	mux.HandleFunc("POST /store-api/files", r.fileHandler.UploadFile)
	mux.HandleFunc("GET /store-api/files", r.fileHandler.GetAllFiles)
	mux.HandleFunc("GET /store-api/files/{id}", r.fileHandler.GetFile)
	mux.HandleFunc("GET /store-api/files/{id}/download", r.fileHandler.DownloadFile)

	// Swagger docs
	mux.HandleFunc("GET /store-api/docs/", r.docsHandler.Docs)
	mux.HandleFunc("GET /store-api/docs/swagger.json", r.docsHandler.Swagger)

	return mux
}
