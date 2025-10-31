package handler

import (
	"fmt"
	"net/http"
	"strconv"

	models "github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository"
)

type Handlers struct {
	repository repository.Repository
}

func InitHandlers(r repository.Repository) *Handlers {
	return &Handlers{
		repository: r,
	}
}

func (h *Handlers) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	t := r.PathValue("type")
	n := r.PathValue("name")
	v := r.PathValue("value")

	fmt.Printf("%s %s %s\n", t, n, v)

	if t != models.Gauge && t != models.Counter {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := strconv.ParseFloat(v, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_ = h.repository.SetMetric(n, t, value)

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetMetric(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	metric := h.repository.GetMetric(name)
	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%f", *metric.Value)))
	w.WriteHeader(http.StatusOK)
}
