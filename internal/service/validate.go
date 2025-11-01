package service

import "github.com/ibeloyar/metrics/internal/model"

func (s *Service) IsValidMetricType(metricType string) bool {
	if metricType == model.Counter || metricType == model.Gauge {
		return true
	}

	return false
}
