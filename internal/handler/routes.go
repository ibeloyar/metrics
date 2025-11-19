package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/repository"
	"github.com/ibeloyar/metrics/internal/service"
)

func InitRoutes(r *chi.Mux, repo *repository.MemStorage) *chi.Mux {
	handlers := InitHandlers(service.New(repo))

	r.Get("/", handlers.GetMetricsPage)

	r.Post("/value/", handlers.GetMetricInfo)
	r.Get("/value/{type}/{name}", handlers.GetMetric)

	r.Post("/update/", handlers.UpdateMetric)
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricInQuery)

	return r
}
