package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/ibeloyar/metrics/internal/model"

	service "github.com/ibeloyar/metrics/internal/service/mocks"
)

// Helpers
func ptrFloat64(v float64) *float64 { return &v }
func ptrInt64(v int64) *int64       { return &v }

func ExampleMetricsHandler_GetMetricQuery() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	mockService.EXPECT().IsValidMetricType(model.Gauge).Return(true)
	mockService.EXPECT().GetMetric("cpu_load").Return(
		&model.Metrics{
			ID:    "cpu_load",
			MType: model.Gauge,
			Value: ptrFloat64(1.5),
		},
		nil,
	)

	h := NewMetricsHandler(mockService, nil, "")
	r := chi.NewRouter()
	InitRoutes(r, h, false)

	req := httptest.NewRequest("GET", "/value/gauge/cpu_load", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	fmt.Println("Status:", rr.Code)
	fmt.Println("Content-Type:", rr.Header().Get("Content-Type"))
	fmt.Println("Body:", strings.TrimSpace(rr.Body.String()))
}

func ExampleMetricsHandler_UpdateMetric_json_hmac() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	mockService.EXPECT().IsValidMetricType(model.Counter).Return(true)
	mockService.EXPECT().ValidateMetric(gomock.Any()).Return(nil)
	mockService.EXPECT().SetMetric(gomock.Any()).Return(nil)

	metric := model.Metrics{
		ID:    "requests",
		MType: model.Counter,
		Delta: ptrInt64(100),
	}
	body, _ := json.Marshal(metric)
	key := "secret"
	expectedHash := GetHashBodySHA256(body, key)

	h := NewMetricsHandler(mockService, nil, key)
	r := chi.NewRouter()
	InitRoutes(r, h, false)

	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", expectedHash)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	fmt.Println("Status:", rr.Code)
	fmt.Println("Hash header:", rr.Header().Get("HashSHA256"))
}

func ExampleMetricsHandler_UpdateMetrics_batch() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockService := service.NewMockService(ctrl)
	metrics := []model.Metrics{
		{ID: "cpu", MType: model.Gauge, Value: ptrFloat64(1.2)},
		{ID: "requests", MType: model.Counter, Delta: ptrInt64(50)},
	}
	mockService.EXPECT().ValidateMetrics(metrics).Return(nil)
	mockService.EXPECT().SetMetrics(metrics, gomock.Any()).Return(nil)

	body, _ := json.Marshal(metrics)
	key := "secret"
	expectedHash := GetHashBodySHA256(body, key)

	h := NewMetricsHandler(mockService, nil, key)
	r := chi.NewRouter()
	InitRoutes(r, h, false)

	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", expectedHash)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	fmt.Println("Status:", rr.Code)
}
