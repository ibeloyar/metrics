package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/repository"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

func Run(config config.Config) {
	router := chi.NewRouter()
	repo := repository.New()

	fmt.Println("Starting server on " + config.Addr)
	if err := http.ListenAndServe(config.Addr, handler.InitRoutes(router, repo)); err != nil {
		panic(err)
	}
}
