//go:build wireinject
// +build wireinject

package di

import (
	hashRealizations "filestoringservice/internal/infrastructure/hash"
	hashInterface "filestoringservice/internal/interfaces/hash"
	"filestoringservice/internal/interfaces/repository"
	"github.com/google/wire"

	"filestoringservice/internal/application/service"
	"filestoringservice/internal/infrastructure/config"
	"filestoringservice/internal/infrastructure/persistence/postgres"
	"filestoringservice/internal/infrastructure/storage/s3"
	"filestoringservice/internal/interfaces/api/handler"
	"filestoringservice/internal/interfaces/api/router"
)

// RepositorySet provides repository implementations
var RepositorySet = wire.NewSet(
	postgres.NewFileRepository,
	wire.Bind(new(repository.FileRepository), new(*postgres.FileRepository)),
)

var HasherSet = wire.NewSet(
	hashRealizations.NewBLAKE3Hasher,
	wire.Bind(new(hashInterface.Hasher), new(*hashRealizations.BLAKE3Hasher)),
)

// InitializeApplication wires up all the dependencies
func InitializeApplication() (*Application, error) {
	wire.Build(
		// Configurations.
		config.Load,

		// Hasher
		HasherSet,

		// Databases.
		postgres.NewDB,

		// Repositories.
		RepositorySet,

		// Storages.
		s3.NewFileStorage,

		// Services.
		service.NewFileService,

		// Handlers.
		handler.NewFileHandler,
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
