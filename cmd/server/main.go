package main

import (
	"log"
	"net/http"

	"github.com/ibeloyar/metrics/internal/handler"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /update/{type}/{name}/{value}", handler.UpdateMetric)

	if err := http.ListenAndServe("127.0.0.1:8080", mux); err != nil {
		log.Fatal(err)
	}
}
