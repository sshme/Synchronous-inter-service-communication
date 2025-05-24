//go:build wireinject
// +build wireinject

package di

import (
	"fileanalysisservice/internal/infrastructure/filestoringservice"
	"fileanalysisservice/internal/infrastructure/quickchart"
	"fileanalysisservice/internal/interfaces/repository"

	"github.com/google/wire"

	"fileanalysisservice/internal/application/service"
	"fileanalysisservice/internal/infrastructure/config"
	"fileanalysisservice/internal/infrastructure/persistence/postgres"
	"fileanalysisservice/internal/infrastructure/storage/s3"
	"fileanalysisservice/internal/interfaces/api/handler"
	"fileanalysisservice/internal/interfaces/api/router"
)

// RepositorySet provides repository implementations
var RepositorySet = wire.NewSet(
	postgres.NewAnalysisRepository,
	wire.Bind(new(repository.AnalysisRepository), new(*postgres.AnalysisRepository)),
	postgres.NewShingleRepository,
	wire.Bind(new(repository.ShingleRepository), new(*postgres.ShingleRepository)),
)

// InitializeApplication wires up all the dependencies
func InitializeApplication() (*Application, error) {
	wire.Build(
		// Configurations.
		config.Load,

		// Databases.
		postgres.NewDB,

		// External Services.
		filestoringservice.NewFileStoringService,
		quickchart.NewQuickChart,

		// Repositories.
		RepositorySet,

		// Storages.
		s3.NewFileStorage,

		// Services.
		service.NewContentAnalyserService,

		// Handlers.
		handler.NewAnalysisHandler,
		handler.NewInfoHandler,
		handler.NewDocsHandler,

		// Routers.
		router.NewRouter,

		// Application.
		NewApplication,
	)

	return &Application{}, nil
}

// Application is the main application container
type Application struct {
	Router *router.Router
	Config *config.Config
}

// NewApplication creates a new application
func NewApplication(router *router.Router, config *config.Config) *Application {
	return &Application{
		Router: router,
		Config: config,
	}
}
