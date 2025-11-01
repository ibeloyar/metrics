package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/repository"
)

func Run() {
	router := chi.NewRouter()
	repo := repository.New()

	if err := http.ListenAndServe(":8080", handler.InitRoutes(router, repo)); err != nil {
		panic(err)
	}
}
