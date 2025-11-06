package server

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/handler"
	"github.com/ibeloyar/metrics/internal/repository"
)

func Run() {
	var addr string
	
	flag.StringVar(&addr, "a", ":8080", "The address to listen on")

	flag.Parse()

	router := chi.NewRouter()
	repo := repository.New()

	fmt.Println("Starting server on " + addr)
	if err := http.ListenAndServe(addr, handler.InitRoutes(router, repo)); err != nil {
		panic(err)
	}
}
