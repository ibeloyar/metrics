package memstorage

import (
	"errors"

	"github.com/ibeloyar/metrics/internal/model"
)

type MemStorage struct {
	metrics map[string]models.Metrics
}

var Storage MemStorage

func GetMetricByID(id string) *models.Metrics {
	v, ok := Storage.metrics[id]
	if !ok {
		return nil
	}

	return &v
}

func UpdateCount(id string, counter float64) error {
	metric := GetMetricByID(id)
	if metric == nil {
		return errors.New("metric not found")
	}

	Storage.metrics[id] = models.Metrics{
		ID:    id,
		MType: models.Counter,
		Value: &counter,
		Delta: nil,
		Hash:  "",
	}
	return nil
}

func UpdateGauge(id string, gauge float64) error {
	metric := GetMetricByID(id)
	if metric == nil {
		return errors.New("metric not found")
	}

	v := *Storage.metrics[id].Value + gauge

	Storage.metrics[id] = models.Metrics{
		ID:    id,
		MType: models.Gauge,
		Value: &v,
		Delta: nil,
		Hash:  "",
	}

	return nil
}
