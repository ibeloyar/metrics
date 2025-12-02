package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/logger"
	"github.com/ibeloyar/metrics/internal/middleware/gzip"
	"github.com/ibeloyar/metrics/internal/repository/filestorage"
	"github.com/ibeloyar/metrics/internal/repository/memstorage"
	"github.com/ibeloyar/metrics/internal/repository/pgstorage"
	"github.com/ibeloyar/metrics/internal/service"
	"go.uber.org/zap"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

func Run(cfg config.Config) {
	lg, storage, err := initDependencies(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}
	defer lg.Sync()

	srv := buildServer(cfg, storage, lg)

	run(srv, storage, lg, cfg.Addr)
}

func initDependencies(cfg config.Config) (*zap.SugaredLogger, service.Storage, error) {
	var storage service.Storage

	lg, err := logger.New()
	if err != nil {
		return nil, nil, err
	}

	if cfg.DatabaseDSN != "" {
		pgStorage, err := pgstorage.New(cfg.DatabaseDSN)
		if err != nil {
			return nil, nil, err
		}

		storage = pgStorage
	} else {
		fileStorage := filestorage.New(cfg.FileStoragePath)
		repo := memstorage.New(fileStorage, cfg.StoreInterval, cfg.Restore)

		if err := repo.Init(); err != nil {
			shutdownErr := repo.Shutdown()
			if shutdownErr != nil {
				lg.Fatalf("Shutdown (repo) error after Init failure: %v", shutdownErr)
			}
			return nil, nil, err
		}
	}

	return lg, storage, nil
}

func buildServer(cfg config.Config, storage service.Storage, lg *zap.SugaredLogger) *http.Server {
	router := chi.NewRouter()

	router.Use(gzip.Middleware)
	router.Use(logger.LoggingMiddleware(lg))
	router.Use(middleware.Recoverer)

	s := service.New(storage)

	metricsHandler := handler.NewMetricsHandler(s)

	router = handler.InitRoutes(router, metricsHandler)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	return srv
}

func run(srv *http.Server, storage service.Storage, lg *zap.SugaredLogger, addr string) {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	lg.Infof("Starting server on %s", addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lg.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	<-signalCtx.Done()
	lg.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		lg.Fatalf("Shutdown (server) error: %v", err)
	}

	if err := storage.Shutdown(); err != nil {
		lg.Fatalf("Shutdown (repo) error: %v", err)
	}

	lg.Info("Server shutdown success")
}
