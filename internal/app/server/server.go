package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/logger"
	"github.com/ibeloyar/metrics/internal/middleware/gzip"
	"github.com/ibeloyar/metrics/internal/repository"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

func Run(config config.Config) {
	lg, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	repo := repository.New()

	router.Use(gzip.Middleware)
	router.Use(logger.LoggingMiddleware(lg))
	router.Use(middleware.Recoverer)

	lg.Infof("Starting server on %s", config.Addr)
	if err := http.ListenAndServe(config.Addr, handler.InitRoutes(router, repo)); err != nil {
		lg.Fatal(err)
	}
}
