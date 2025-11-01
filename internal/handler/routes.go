package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/repository"
)

func InitRoutes(r *chi.Mux, s *repository.MemStorage) *chi.Mux {
	handlers := InitHandlers(*s)

	r.Get("/", handlers.GetMetricsPage)
	r.Get("/value/{name}", handlers.GetMetric)
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetric)

	return r
}
