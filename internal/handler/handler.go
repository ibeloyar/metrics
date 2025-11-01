package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository"
)

type Handlers struct {
	storage repository.MemStorage
}

func InitHandlers(s repository.MemStorage) *Handlers {
	return &Handlers{
		storage: s,
	}
}

func (h *Handlers) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")
	v := chi.URLParam(r, "value")

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

	_ = h.storage.SetMetric(n, t, value)

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetMetric(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	metric := h.storage.GetMetric(name)
	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%f", *metric.Value)))
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetMetricsPage(w http.ResponseWriter, r *http.Request) {
	metrics := h.storage.GetMetrics()

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("<h1>Metrics</h1>"))
	w.Write([]byte("<table border='1'><thead><tr><th>Key</th><th>Value</th></tr></thead><tbody>"))
	for nameID, metric := range metrics {
		row := fmt.Sprintf("<tr><td>%s</td><td>%v</td></tr>", nameID, *metric.Value)
		w.Write([]byte(row))
	}
	w.Write([]byte("</tbody></table>"))
}
