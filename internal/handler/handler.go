package handler

import (
	"fmt"
	"net/http"
	"strconv"

	models "github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/service"
)

func UpdateMetric(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/plain")

	metricType := r.PathValue("type")
	metricName := r.PathValue("name")
	metricValue := r.PathValue("value")

	fmt.Println(metricType, metricName, metricValue)

	// Type validate
	if metricType != models.Gauge && metricType != models.Counter {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Name validate
	if ok := service.CheckMetricName(metricName); !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if metricType == models.Counter {
		v, err := strconv.Atoi(metricValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = service.UpdateCounterMetric(metricName, int64(v))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if metricType == models.Gauge {
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = service.UpdateGaugeMetric(metricName, int64(v))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
