package handler

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/model"
)

type Service interface {
	GetMetric(name string) (*model.Metrics, *model.APIError)
	GetMetrics() ([]model.Metrics, *model.APIError)

	SetMetric(metric model.Metrics) *model.APIError

	IsValidMetricType(metricType string) bool
}

type Handlers struct {
	service Service
}

func InitHandlers(s Service) *Handlers {
	return &Handlers{
		service: s,
	}
}

func (h *Handlers) GetMetricQuery(w http.ResponseWriter, r *http.Request) {
	n := chi.URLParam(r, "name")
	t := chi.URLParam(r, "type")

	if !h.service.IsValidMetricType(t) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := h.service.GetMetric(n)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	if metric.MType == model.Gauge {
		w.Write([]byte(strconv.FormatFloat(*metric.Value, 'g', -1, 64)))
		return
	} else if metric.MType == model.Counter {
		w.Write([]byte(strconv.FormatInt(*metric.Delta, 10)))
		return
	}
}

func (h *Handlers) UpdateMetricQuery(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	n := chi.URLParam(r, "name")
	v := chi.URLParam(r, "value")

	if !h.service.IsValidMetricType(t) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if t == model.Counter {
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		apiErr := h.service.SetMetric(model.Metrics{
			ID:    n,
			MType: model.Counter,
			Delta: &value,
		})
		if apiErr != nil {
			http.Error(w, apiErr.Message, apiErr.Code)
			return
		}
	}

	if t == model.Gauge {
		value, err := strconv.ParseFloat(v, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		apiErr := h.service.SetMetric(model.Metrics{
			ID:    n,
			MType: model.Gauge,
			Value: &value,
		})
		if apiErr != nil {
			http.Error(w, apiErr.Message, apiErr.Code)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetMetric(w http.ResponseWriter, r *http.Request) {
	var body model.GetMetricBody

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "reading request body error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		http.Error(w, "unmarshal request body error", http.StatusInternalServerError)
		return
	}

	if !h.service.IsValidMetricType(body.MType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, apiErr := h.service.GetMetric(body.ID)
	if apiErr != nil && apiErr.Code == http.StatusNotFound {
		metric := model.Metrics{
			ID:    body.ID,
			MType: body.MType,
		}

		if body.MType == "gauge" {
			g := float64(0)
			metric.Value = &g
		}

		if body.MType == "counter" {
			d := int64(0)
			metric.Delta = &d
		}

		response, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	response, _ := json.Marshal(metric)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handlers) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	var body model.Metrics

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "reading request body error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		http.Error(w, "unmarshal request body error", http.StatusInternalServerError)
		return
	}

	if !h.service.IsValidMetricType(body.MType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	apiErr := h.service.SetMetric(body)
	if apiErr != nil {
		http.Error(w, apiErr.Message, apiErr.Code)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetMetricsPage(w http.ResponseWriter, r *http.Request) {
	metricsPageTemplate := `
	<h1>Metrics</h1>
	<table border="1">
		<thead>
			<tr><th>Key</th><th>Value</th></tr>
		</thead>
		<tbody>
			{{range .}}
			<tr><td>{{.ID}}</td><td>{{.Value}}</td></tr>
			{{end}}
		</tbody>
	</table>
	`

	metrics, apiErr := h.service.GetMetrics()
	if apiErr != nil {
		http.Error(w, apiErr.Message, apiErr.Code)
		return
	}

	t := template.Must(template.New("metrics").Parse(metricsPageTemplate))

	err := t.Execute(w, metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
}
