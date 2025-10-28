package service

import (
	memstorage "github.com/ibeloyar/metrics/internal/repository"
)

func CheckMetricName(id string) bool {
	metric := memstorage.GetMetricByID(id)
	if metric == nil {
		return false
	}
	return true
}

func UpdateCounterMetric(id string, valueType int64) error {
	return memstorage.UpdateCount(id, float64(valueType))
}

func UpdateGaugeMetric(id string, valueType int64) error {
	return memstorage.UpdateGauge(id, float64(valueType))
}
