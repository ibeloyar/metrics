package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/logger"
	"github.com/ibeloyar/metrics/internal/middleware/gzip"
	"github.com/ibeloyar/metrics/internal/repository/memstorage"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

func Run(config config.Config) {
	lg, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	repo := memstorage.New(config.FileStoragePath, config.StoreInterval, config.Restore)

	if err := repo.Init(); err != nil {
		repo.Shutdown()
		lg.Fatal(err)
	}

	router.Use(gzip.Middleware)
	router.Use(logger.LoggingMiddleware(lg))
	router.Use(middleware.Recoverer)

	lg.Infof("Starting server on %s %d", config.Addr, config.StoreInterval)
	if err := http.ListenAndServe(config.Addr, handler.InitRoutes(router, repo)); err != nil {
		repo.Shutdown()
		lg.Fatal(err)
	}
}
