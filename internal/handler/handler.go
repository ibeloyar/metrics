package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/model"
)

type Service interface {
	GetMetric(name string) (*model.Metrics, *model.APIError)
	GetMetrics() ([]model.Metrics, *model.APIError)

	SetMetric(metricName, metricType string, metricValue float64) *model.APIError

	IsValidMetricType(metricType string) bool
}

type Handlers struct {
	service Service
}

func InitHandlers(s Service) *Handlers {
	return &Handlers{
		service: s,
	}
}

func (h *Handlers) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")
	v := chi.URLParam(r, "value")

	if !h.service.IsValidMetricType(t) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := strconv.ParseFloat(v, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	apiErr := h.service.SetMetric(n, t, value)
	if apiErr != nil {
		http.Error(w, apiErr.Message, apiErr.Code)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetMetric(w http.ResponseWriter, r *http.Request) {
	n := chi.URLParam(r, "name")
	t := chi.URLParam(r, "type")

	if !h.service.IsValidMetricType(t) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetric(n)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(fmt.Sprintf("%g", *metric.Value)))
}

func (h *Handlers) GetMetricsPage(w http.ResponseWriter, r *http.Request) {
	metrics, _ := h.service.GetMetrics()

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("<h1>Metrics</h1>"))
	w.Write([]byte("<table border='1'><thead><tr><th>Key</th><th>Value</th></tr></thead><tbody>"))
	for _, metric := range metrics {
		row := fmt.Sprintf("<tr><td>%s</td><td>%g</td></tr>", metric.ID, *metric.Value)
		w.Write([]byte(row))
	}
	w.Write([]byte("</tbody></table>"))
}
