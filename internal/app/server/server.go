package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ibeloyar/metrics/internal/audit"
	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/logger"
	"github.com/ibeloyar/metrics/internal/middleware/gzip"
	"github.com/ibeloyar/metrics/internal/repository/filestorage"
	"github.com/ibeloyar/metrics/internal/repository/memstorage"
	"github.com/ibeloyar/metrics/internal/repository/pgstorage"
	"github.com/ibeloyar/metrics/internal/service"
	"github.com/ibeloyar/metrics/internal/service/crypto"
	"go.uber.org/zap"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

const ShutdownTimeout = 20 * time.Second

func Run(cfg config.Config) error {
	var storage service.Storage

	lg, err := logger.New()
	if err != nil {
		return err
	}
	defer lg.Sync()

	if cfg.DatabaseDSN != "" {
		pgStorage, pgStorageErr := pgstorage.New(cfg.DatabaseDSN)
		if pgStorageErr != nil {
			return pgStorageErr
		}

		storage = pgStorage
	} else {
		fileStorage := filestorage.New(cfg.FileStoragePath)
		repo := memstorage.New(fileStorage, cfg.StoreInterval, cfg.Restore)

		if initErr := repo.Init(); initErr != nil {
			shutdownErr := repo.Shutdown()
			if shutdownErr != nil {
				return shutdownErr
			}
			return initErr
		}

		storage = repo
	}

	auditSubject, err := initAudit(cfg)
	if err != nil {
		return fmt.Errorf("audit init: %w", err)
	}

	srv := buildServer(cfg, storage, lg, auditSubject)

	return runServer(srv, storage, lg, cfg.Addr)
}

func buildServer(cfg config.Config, storage service.Storage, lg *zap.SugaredLogger, auditSubject *audit.AuditSubject) *http.Server {
	router := chi.NewRouter()

	router.Use(gzip.Middleware)

	if cfg.CryptoKey != "" {
		cryptoManager := crypto.NewCryptoManager(cfg.CryptoKey)

		router.Use(cryptoManager.CryptoMiddleware)
	}

	router.Use(logger.LoggingMiddleware(lg))
	router.Use(middleware.Recoverer)

	s := service.New(storage, auditSubject)

	metricsHandler := handler.NewMetricsHandler(s, lg, cfg.Key)

	router = handler.InitRoutes(router, metricsHandler, cfg.Pprof)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadTimeout:       10 * time.Second,  // time to read request body
		ReadHeaderTimeout: 5 * time.Second,   // time to read headers
		WriteTimeout:      30 * time.Second,  // time to send response
		IdleTimeout:       120 * time.Second, // idle connection timeout
	}

	return srv
}

func runServer(srv *http.Server, storage service.Storage, lg *zap.SugaredLogger, addr string) error {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	lg.Infof("starting server on %s", addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lg.Fatalf("server ListenAndServe error: %v", err)
		}
	}()

	<-signalCtx.Done()
	lg.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := srv.Shutdown(context.Background()); err != nil {
			lg.Warnf("shutdown (server) error: %v", err)
		}
		if err := storage.Shutdown(); err != nil {
			lg.Warnf("shutdown (repo) error: %v", err)
		}
	}()

	select {
	case <-shutdownCtx.Done():
		if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
			lg.Warn("agent shutdown timeout exceeded, forcing exit")
		}
	case <-done:
		lg.Info("agent graceful shutdown completed")
	}

	return nil
}

func initAudit(cfg config.Config) (*audit.AuditSubject, error) {
	auditSubject := audit.NewSubject()

	if cfg.AuditFile != "" {
		if fObs, err := audit.NewFileAuditObserver(cfg.AuditFile); err == nil && fObs != nil {
			auditSubject.Register(fObs)
		} else {
			return nil, fmt.Errorf("file audit init error: %v", err)
		}
	}

	if cfg.AuditURL != "" {
		if hObs, err := audit.NewHTTPAuditObserver(cfg.AuditURL); err == nil && hObs != nil {
			auditSubject.Register(hObs)
		} else {
			return nil, fmt.Errorf("http audit init error: %v", err)
		}
	}

	return auditSubject, nil
}
