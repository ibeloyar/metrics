package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type MetricHandlers interface {
	GetMetricQuery(w http.ResponseWriter, r *http.Request)
	UpdateMetricQuery(w http.ResponseWriter, r *http.Request)
	GetMetric(w http.ResponseWriter, r *http.Request)
	UpdateMetric(w http.ResponseWriter, r *http.Request)
	GetMetricsPage(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
}

func InitRoutes(r *chi.Mux, metricHandlers MetricHandlers) *chi.Mux {

	r.Get("/", metricHandlers.GetMetricsPage)
	r.Get("/ping", metricHandlers.Ping)

	r.Post("/value/", metricHandlers.GetMetric)
	r.Get("/value/{type}/{name}", metricHandlers.GetMetricQuery)

	r.Post("/update/", metricHandlers.UpdateMetric)
	r.Post("/update/{type}/{name}/{value}", metricHandlers.UpdateMetricQuery)

	return r
}
