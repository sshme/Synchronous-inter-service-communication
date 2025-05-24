package handler

import (
	"net/http"
)

// InfoHandler handles HTTP requests related to server info
type InfoHandler struct {
}

func NewInfoHandler() *InfoHandler {
	return &InfoHandler{}
}

// HealthCheck handles the health check endpoint
// @Summary Health check endpoint
// @Description Check if the service is up and running
// @Tags health
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /info/health [get]
func (h *InfoHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}
