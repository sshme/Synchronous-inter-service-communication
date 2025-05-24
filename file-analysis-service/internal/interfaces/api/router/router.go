package router

import (
	"net/http"

	"fileanalysisservice/internal/interfaces/api/handler"

	_ "fileanalysisservice/docs" // Import for swagger docs
)

// Router handles HTTP routing
type Router struct {
	analyseHandler *handler.AnalyseHandler
	infoHandler    *handler.InfoHandler
	docsHandler    *handler.DocsHandler
}

// NewRouter creates a new router
func NewRouter(analyseHandler *handler.AnalyseHandler, infoHandler *handler.InfoHandler, docsHandler *handler.DocsHandler) *Router {
	return &Router{
		analyseHandler: analyseHandler,
		infoHandler:    infoHandler,
		docsHandler:    docsHandler,
	}
}

// SetupRoutes sets up the HTTP routes
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Info routes
	mux.HandleFunc("GET /analysis-api/info/health", r.infoHandler.HealthCheck)

	// Analyse routes
	mux.HandleFunc("GET /analysis-api/analysis/{id}", r.analyseHandler.GetAnalyse)
	mux.HandleFunc("GET /analysis-api/analysis/{id}/download", r.analyseHandler.DownloadCloud)

	// Swagger docs
	mux.HandleFunc("GET /analysis-api/docs/", r.docsHandler.Docs)
	mux.HandleFunc("GET /analysis-api/docs/swagger.json", r.docsHandler.Swagger)

	return mux
}
