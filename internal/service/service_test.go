package service

import (
	"testing"

	"github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository/filestorage"
	"github.com/ibeloyar/metrics/internal/repository/memstorage"
	"github.com/ibeloyar/metrics/internal/repository/pgstorage"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

func TestService(t *testing.T) {
	cfg := config.Config{
		StoreInterval:   300,
		Restore:         true,
		FileStoragePath: "./testdata",
	}
	fileStorage := filestorage.New(cfg.FileStoragePath)
	dbStorage, err := pgstorage.New(cfg.DatabaseDSN)
	if err != nil {
		t.Error("pgstorage init error")
	}
	repo := memstorage.New(fileStorage, cfg.StoreInterval, cfg.Restore)
	srv := New(repo, dbStorage)

	t.Run("IsValidMetricType", func(t *testing.T) {
		if !srv.IsValidMetricType(model.Gauge) {
			t.Error("Expected true for valid metric type gauge, got false")
		}
		if !srv.IsValidMetricType(model.Counter) {
			t.Error("Expected true for valid metric type counter, got false")
		}

		invalidTypes := []string{"", "rate", "summary", "histogram", "COUNTER"}
		for _, typ := range invalidTypes {
			if srv.IsValidMetricType(typ) {
				t.Errorf("Expected false for invalid metric type %q, got true", typ)
			}
		}
	})
}
