package handler

import (
	"net/http"

	"github.com/ibeloyar/metrics/internal/repository"
)

func InitRoutes(mux *http.ServeMux, repository *repository.Repository) *http.ServeMux {
	handlers := InitHandlers(*repository)

	mux.HandleFunc("GET /metric/{name}", handlers.GetMetric)
	mux.HandleFunc("POST /update/{type}/{name}/{value}", handlers.UpdateMetric)

	return mux
}
