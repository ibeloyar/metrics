package service

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/ibeloyar/metrics/internal/audit"
	config "github.com/ibeloyar/metrics/internal/config/server"
	"github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository/filestorage"
	"github.com/ibeloyar/metrics/internal/repository/memstorage"
)

func pointer[T any](v T) *T {
	return &v
}

func TestService_IsValidMetricType(t *testing.T) {
	cfg := config.Config{
		StoreInterval:   300,
		Restore:         true,
		FileStoragePath: "./testdata",
	}

	fileStorage := filestorage.New(cfg.FileStoragePath)
	auditSubject := audit.NewSubject()
	repo := memstorage.New(fileStorage, cfg.StoreInterval, cfg.Restore)
	srv := New(repo, auditSubject)

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

func TestService_SetMetric(t *testing.T) {
	storage := memstorage.New(filestorage.New(""), 30, true)
	service := New(storage, audit.NewSubject())

	tests := []struct {
		name    string
		metric  *model.Metrics
		wantErr *model.APIError
	}{
		{
			name: "gauge_success",
			metric: &model.Metrics{
				ID:    "myGauge",
				MType: model.Gauge,
				Value: pointer(123.45),
			},
			wantErr: nil,
		},
		{
			name: "counter_success",
			metric: &model.Metrics{
				ID:    "myCounter",
				MType: model.Counter,
				Delta: pointer(int64(1)),
			},
			wantErr: nil,
		},
		{
			name: "invalid_type_top",
			metric: &model.Metrics{
				ID:    "invalid",
				MType: "unknown", // IsValidMetricType return false
			},
			wantErr: &model.APIError{
				Code:    http.StatusBadRequest,
				Message: "invalid metric type",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := service.SetMetric(tt.metric)

			if tt.wantErr == nil {
				if gotErr != nil {
					t.Errorf("SetMetric() error = %v, want nil", gotErr)
				}
			} else {
				if gotErr == nil ||
					gotErr.Code != tt.wantErr.Code ||
					gotErr.Message != tt.wantErr.Message {
					t.Errorf("SetMetric() error = %+v, want %+v", gotErr, tt.wantErr)
				}
			}
		})
	}
}

func TestService_SetMetrics(t *testing.T) {
	storage := memstorage.New(filestorage.New(""), 30, true)
	service := New(storage, audit.NewSubject())

	tests := []struct {
		name    string
		metrics []model.Metrics
		addr    string
		wantErr *model.APIError
	}{
		{
			name: "success",
			metrics: []model.Metrics{
				{ID: "gauge1", MType: model.Gauge, Value: pointer(123.45)},
				{ID: "counter1", MType: model.Counter, Delta: pointer(int64(1))},
			},
			addr:    "127.0.0.1:8080",
			wantErr: nil,
		},
		{
			name:    "empty_metrics",
			metrics: []model.Metrics{},
			addr:    "10.0.0.1:8080",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := service.SetMetrics(tt.metrics, tt.addr)

			if tt.wantErr == nil {
				if gotErr != nil {
					t.Errorf("SetMetrics() error = %v, want nil", gotErr)
				}
			} else {
				if gotErr == nil || gotErr.Code != tt.wantErr.Code || gotErr.Message != tt.wantErr.Message {
					t.Errorf("SetMetrics() error = %+v, want %+v", gotErr, tt.wantErr)
				}
			}
		})
	}
}

func TestService_GetMetric(t *testing.T) {
	storage := memstorage.New(filestorage.New(""), 30, true)
	service := New(storage, audit.NewSubject())

	testMetric := &model.Metrics{
		ID:    "testGauge",
		MType: model.Gauge,
		Value: pointer(123.45),
	}
	storage.SetMetric(testMetric)

	tests := []struct {
		name       string
		metricName string
		wantMetric *model.Metrics
		wantErr    *model.APIError
	}{
		{
			name:       "metric_found",
			metricName: "testGauge",
			wantMetric: testMetric,
			wantErr:    nil,
		},
		{
			name:       "metric_not_found",
			metricName: "nonexistent",
			wantMetric: nil,
			wantErr: &model.APIError{
				Code:    http.StatusNotFound,
				Message: "metric not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetric, gotErr := service.GetMetric(tt.metricName)

			if tt.wantErr == nil {
				if gotErr != nil {
					t.Errorf("GetMetric() error = %v, want nil", gotErr)
					return
				}
			} else {
				if gotErr == nil ||
					gotErr.Code != tt.wantErr.Code ||
					gotErr.Message != tt.wantErr.Message {
					t.Errorf("GetMetric() error = %+v, want %+v", gotErr, tt.wantErr)
					return
				}
			}

			if tt.wantMetric == nil {
				if gotMetric != nil {
					t.Errorf("GetMetric() = %v, want nil", gotMetric)
				}
			} else {
				if !reflect.DeepEqual(gotMetric, tt.wantMetric) {
					t.Errorf("GetMetric() = %v, want %v", gotMetric, tt.wantMetric)
				}
			}
		})
	}
}

func TestService_GetMetrics(t *testing.T) {
	storage := memstorage.New(filestorage.New(""), 30, true)
	service := New(storage, audit.NewSubject())

	testMetrics := []model.Metrics{
		{
			ID:    "gauge1",
			MType: model.Gauge,
			Value: pointer(123.45),
		},
		{
			ID:    "counter1",
			MType: model.Counter,
			Delta: pointer(int64(42)),
		},
	}

	for _, metric := range testMetrics {
		storage.SetMetric(&metric)
	}

	tests := []struct {
		name       string
		setup      func()
		wantResult []model.Metrics
		wantErr    *model.APIError
	}{
		{
			name:       "with_metrics",
			setup:      func() {},
			wantResult: testMetrics,
			wantErr:    nil,
		},
		{
			name: "empty_storage",
			setup: func() {
				service.storage = memstorage.New(filestorage.New(""), 30, true)
			},
			wantResult: []model.Metrics{},
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			gotResult, gotErr := service.GetMetrics()

			if tt.wantErr != nil {
				if gotErr == nil || gotErr.Code != tt.wantErr.Code || gotErr.Message != tt.wantErr.Message {
					t.Errorf("GetMetrics() error = %+v, want %+v", gotErr, tt.wantErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("GetMetrics() error = %v, want nil", gotErr)
					return
				}

				if !reflect.DeepEqual(gotResult, tt.wantResult) {
					t.Errorf("GetMetrics() = %v, want %v", gotResult, tt.wantResult)
				}
			}
		})
	}
}

func TestService_Ping(t *testing.T) {
	storage := memstorage.New(filestorage.New(""), 30, true)
	service := New(storage, audit.NewSubject())

	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "success",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := service.Ping()

			if gotErr != nil {
				t.Errorf("Ping() error = %v, want nil", gotErr)
			}
		})
	}
}
