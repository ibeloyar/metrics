package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/model"
	service "github.com/ibeloyar/metrics/internal/service/mocks"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Helpers
func floatPtr(f float64) *float64 { return &f }
func int64Ptr(i int64) *int64     { return &i }

func newTestLogger(t *testing.T) *zap.SugaredLogger {
	logger := zaptest.NewLogger(t)
	return logger.Sugar()
}

func TestMetricsHandler_GetMetricsPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	metrics := []model.Metrics{
		{ID: "cpu", MType: model.Gauge, Value: floatPtr(0.75)},
		{ID: "requests", MType: model.Counter, Delta: int64Ptr(100)},
	}
	mockSvc.EXPECT().GetMetrics().Return(metrics, nil)

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "secret")
	r := chi.NewRouter()
	// Предполагаем InitRoutes(r, h, false) или добавьте роуты вручную
	r.Get("/", h.GetMetricsPage)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("wrong status code: got %v want %v", status, http.StatusOK)
	}
	if rr.Header().Get("Content-Type") != "text/html" {
		t.Errorf("wrong Content-Type: got %v want %v", rr.Header().Get("Content-Type"), "text/html")
	}
	if !strings.Contains(rr.Body.String(), "cpu") || !strings.Contains(rr.Body.String(), "0.75") {
		t.Errorf("expected metrics in HTML table")
	}
}

func TestMetricsHandler_Ping_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().Ping().Return(nil)

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "")
	r := chi.NewRouter()
	r.Get("/ping", h.Ping)

	req := httptest.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestMetricsHandler_GetMetricQuery_Gauge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType("gauge").Return(true)
	metric := &model.Metrics{
		ID:    "cpu_load",
		MType: model.Gauge,
		Value: floatPtr(1.5),
	}
	mockSvc.EXPECT().GetMetric("cpu_load").Return(metric, nil)

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "")
	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", h.GetMetricQuery)

	req := httptest.NewRequest("GET", "/value/gauge/cpu_load", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("wrong status code: got %v want %v", status, http.StatusOK)
	}
	if rr.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("wrong Content-Type: got %v want %v", rr.Header().Get("Content-Type"), "text/plain")
	}
	if body := rr.Body.String(); body != "1.5" {
		t.Errorf("wrong body: got %q want %q", body, "1.5")
	}
}

func TestMetricsHandler_GetMetricQuery_Counter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType("counter").Return(true)
	metric := &model.Metrics{
		ID:    "requests",
		MType: model.Counter,
		Delta: int64Ptr(100),
	}
	mockSvc.EXPECT().GetMetric("requests").Return(metric, nil)

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "")
	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", h.GetMetricQuery)

	req := httptest.NewRequest("GET", "/value/counter/requests", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Body.String() != "100" {
		t.Errorf("wrong body: got %q want %q", rr.Body.String(), "100")
	}
}

func TestMetricsHandler_UpdateMetricQuery_Gauge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType("gauge").Return(true)
	mockSvc.EXPECT().SetMetric(&model.Metrics{
		ID:    "cpu",
		MType: model.Gauge,
		Value: floatPtr(0.75),
	}).Return(nil)

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "")
	r := chi.NewRouter()
	r.Post("/value/{type}/{name}/value/{value}", h.UpdateMetricQuery)

	req := httptest.NewRequest("POST", "/value/gauge/cpu/value/0.75", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestMetricsHandler_UpdateMetricQuery_InvalidValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType("gauge").Return(true)

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "")
	r := chi.NewRouter()
	r.Post("/value/{type}/{name}/value/{value}", h.UpdateMetricQuery)

	req := httptest.NewRequest("POST", "/value/gauge/cpu/value/invalid", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestMetricsHandler_UpdateMetric_JSON_HMAC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType(model.Counter).Return(true)
	mockSvc.EXPECT().ValidateMetric(gomock.Any()).Return(nil)
	mockSvc.EXPECT().SetMetric(gomock.Any()).Return(nil)

	metric := model.Metrics{
		ID:    "requests",
		MType: model.Counter,
		Delta: int64Ptr(100),
	}
	body, _ := json.Marshal(metric)
	expectedHash := handler.GetHashBodySHA256(body, "secret")

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "secret")
	r := chi.NewRouter()
	r.Post("/update/", h.UpdateMetric)

	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", expectedHash)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestMetricsHandler_UpdateMetrics_Batch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metrics := []model.Metrics{
		{ID: "cpu", MType: model.Gauge, Value: floatPtr(1.2)},
		{ID: "requests", MType: model.Counter, Delta: int64Ptr(50)},
	}

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().ValidateMetrics(metrics).Return(nil)
	mockSvc.EXPECT().SetMetrics(metrics, gomock.Any()).Return(nil)

	body, _ := json.Marshal(metrics)
	expectedHash := handler.GetHashBodySHA256(body, "secret")

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "secret")
	r := chi.NewRouter()
	r.Post("/updates/", h.UpdateMetrics)

	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", expectedHash)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestMetricsHandler_UpdateMetric_InvalidHMAC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)

	metric := model.Metrics{ID: "requests", MType: model.Counter, Delta: int64Ptr(100)}
	body, _ := json.Marshal(metric)

	h := handler.NewMetricsHandler(mockSvc, newTestLogger(t), "secret")
	r := chi.NewRouter()
	r.Post("/update/", h.UpdateMetric)

	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", "invalid-hash")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestMetricsHandler_GetMetric_JSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType("gauge").Return(true)
	metric := &model.Metrics{ID: "cpu", MType: model.Gauge, Value: floatPtr(1.5)}
	mockSvc.EXPECT().GetMetric("cpu").Return(metric, nil)

	logger := newTestLogger(t)
	h := handler.NewMetricsHandler(mockSvc, logger, "secret")
	r := chi.NewRouter()
	r.Post("/api/v1/{type}/{name}/value", h.GetMetric) // ← Точный роут!

	body := model.GetMetricBody{ID: "cpu", MType: "gauge"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/gauge/cpu/value", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("wrong status: got %d want %d", status, http.StatusOK)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("wrong Content-Type: got %q want %q", rr.Header().Get("Content-Type"), "application/json")
	}
	// Проверяем HMAC заголовок
	if hash := rr.Header().Get("HashSHA256"); hash == "" {
		t.Error("expected HashSHA256 header")
	}
}

func TestMetricsHandler_GetMetric_WrongType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType("unknown").Return(false)

	logger := newTestLogger(t)
	h := handler.NewMetricsHandler(mockSvc, logger, "")
	r := chi.NewRouter()
	r.Post("/api/v1/{type}/{name}/value", h.GetMetric)

	body := model.GetMetricBody{ID: "cpu", MType: "unknown"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/unknown/cpu/value", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("wrong status: got %d want %d", status, http.StatusBadRequest)
	}
}

func TestMetricsHandler_GetMetricsPage_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	apiErr := &model.APIError{Code: 500, Message: "db error"}
	mockSvc.EXPECT().GetMetrics().Return(nil, apiErr)

	logger := newTestLogger(t)
	h := handler.NewMetricsHandler(mockSvc, logger, "")
	r := chi.NewRouter()
	r.Get("/", h.GetMetricsPage)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != 500 {
		t.Fatalf("wrong status: got %d want %d", status, 500)
	}
}

func TestMetricsHandler_Ping_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().Ping().Return(fmt.Errorf("db down"))

	logger := newTestLogger(t)
	h := handler.NewMetricsHandler(mockSvc, logger, "")
	r := chi.NewRouter()
	r.Get("/ping", h.Ping)

	req := httptest.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("wrong status: got %d want %d", status, http.StatusInternalServerError)
	}
}

func TestMetricsHandler_UpdateMetrics_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := newTestLogger(t)
	h := handler.NewMetricsHandler(nil, logger, "")
	r := chi.NewRouter()
	r.Post("/updates/", h.UpdateMetrics)

	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("HashSHA256", "dummyhash")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("wrong status: got %d want %d", status, http.StatusInternalServerError)
	}
}

func TestMetricsHandler_UpdateMetric_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType("unknown").Return(false)

	metric := model.Metrics{ID: "test", MType: "unknown", Delta: int64Ptr(1)}
	body, _ := json.Marshal(metric)
	expectedHash := handler.GetHashBodySHA256(body, "secret")

	logger := newTestLogger(t)
	h := handler.NewMetricsHandler(mockSvc, logger, "secret")
	r := chi.NewRouter()
	r.Post("/update/", h.UpdateMetric)

	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", expectedHash)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("wrong status: got %d want %d", status, http.StatusBadRequest)
	}
}

// Добавить недостающие helpers

func TestMetricsHandler_GetMetricQuery_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := service.NewMockService(ctrl)
	mockSvc.EXPECT().IsValidMetricType("invalid").Return(false)

	logger := newTestLogger(t)
	h := handler.NewMetricsHandler(mockSvc, logger, "")
	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", h.GetMetricQuery)

	req := httptest.NewRequest("GET", "/value/invalid/cpu", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("wrong status: got %d want %d", status, http.StatusBadRequest)
	}
}
