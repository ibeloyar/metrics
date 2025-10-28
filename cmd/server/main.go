package main

import (
	"fmt"
	"net/http"
	"strconv"

	models "github.com/ibeloyar/metrics/internal/model"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /update/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		t := r.PathValue("type")
		n := r.PathValue("name")
		v := r.PathValue("value")

		fmt.Printf("%s %s %s\n", t, n, v)

		if t != models.Gauge && t != models.Counter {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err := strconv.ParseFloat(v, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
