package main

// @title           File Storing Service API
// @version         1.0
// @description     A service for uploading and retrieving files
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost
// @BasePath  /store-api

// @schemes http https
// @produce  json
// @consumes json multipart/form-data

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"filestoringservice/internal/di"
)

func main() {
	// Initialize the application
	app, err := di.InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	server := &http.Server{
		Addr:    ":" + app.Config.ServerPort,
		Handler: app.Router.SetupRoutes(),
	}

	// Channel for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", app.Config.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
