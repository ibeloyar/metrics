package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/repository"
	"github.com/ibeloyar/metrics/internal/service"
)

func InitRoutes(r *chi.Mux, repo *repository.MemStorage) *chi.Mux {
	handlers := InitHandlers(service.New(repo))

	r.Get("/", handlers.GetMetricsPage)
	r.Post("/value/", handlers.GetMetric)
	r.Post("/update/", handlers.UpdateMetric)

	return r
}
