package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/repository/memstorage"
	"github.com/ibeloyar/metrics/internal/service"
)

func InitRoutes(r *chi.Mux, repo *memstorage.MemStorage) *chi.Mux {
	handlers := InitHandlers(service.New(repo))

	r.Get("/", handlers.GetMetricsPage)

	r.Post("/value/", handlers.GetMetric)
	r.Get("/value/{type}/{name}", handlers.GetMetricQuery)

	r.Post("/update/", handlers.UpdateMetric)
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricQuery)

	return r
}
