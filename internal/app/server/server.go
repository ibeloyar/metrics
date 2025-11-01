package server

import (
	"net/http"

	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/repository"
)

func Run() {
	mux := http.NewServeMux()
	repo := repository.New()

	if err := http.ListenAndServe(":8080", handler.InitRoutes(mux, repo)); err != nil {
		panic(err)
	}
}
