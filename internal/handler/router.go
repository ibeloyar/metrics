package handler

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

type MetricHandlers interface {
	GetMetric(w http.ResponseWriter, r *http.Request)
	GetMetricQuery(w http.ResponseWriter, r *http.Request)
	UpdateMetric(w http.ResponseWriter, r *http.Request)
	UpdateMetrics(w http.ResponseWriter, r *http.Request)
	UpdateMetricQuery(w http.ResponseWriter, r *http.Request)
	GetMetricsPage(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
}

// InitRoutes configures Chi router with metrics API endpoints and optional pprof.
//
// Standard Metrics API endpoints:
//
//	GET      /                             - HTML metrics table
//	GET      /ping                         - health check
//	GET      /value/{type}/{name}          - get metric (query params)
//	POST     /value/                       - get metric (JSON)
//	POST     /update/                      - update metric (JSON + HMAC)
//	POST     /updates/                     - batch update (JSON + HMAC)
//	POST     /update/{type}/{name}/{value} - update metric (query params)
//
// Optional: /debug/pprof/* when pprofFlag=true
func InitRoutes(r *chi.Mux, metricHandlers MetricHandlers, pprofFlag bool) *chi.Mux {

	r.Get("/", metricHandlers.GetMetricsPage)
	r.Get("/ping", metricHandlers.Ping)

	r.Post("/value/", metricHandlers.GetMetric)
	r.Get("/value/{type}/{name}", metricHandlers.GetMetricQuery)

	r.Post("/update/", metricHandlers.UpdateMetric)
	r.Post("/updates/", metricHandlers.UpdateMetrics)
	r.Post("/update/{type}/{name}/{value}", metricHandlers.UpdateMetricQuery)

	if pprofFlag {
		r.Mount("/debug/pprof", http.HandlerFunc(pprof.Index))
		r.Handle("/debug/pprof/*", http.DefaultServeMux)
	}

	return r
}
