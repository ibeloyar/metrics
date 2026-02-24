package service

import (
	"errors"

	"github.com/ibeloyar/metrics/internal/model"
)

// IsValidMetricType checks if metric type is "gauge" or "counter".
func (s *Service) IsValidMetricType(metricType string) bool {
	if metricType == model.Counter || metricType == model.Gauge {
		return true
	}

	return false
}

// ValidateMetric validates single metric type and required fields.
func (s *Service) ValidateMetric(metric model.Metrics) error {
	if metric.MType != model.Counter && metric.MType != model.Gauge {
		return errors.New("invalid metric type")
	}

	if metric.MType == model.Counter && metric.Delta == nil {
		return errors.New("delta must be set for counter")
	}

	if metric.MType == model.Gauge && metric.Value == nil {
		return errors.New("value must be set for gauge")
	}

	return nil
}

// ValidateMetrics validates all metrics in batch.
func (s *Service) ValidateMetrics(metrics []model.Metrics) error {
	for _, metric := range metrics {
		if err := s.ValidateMetric(metric); err != nil {
			return err
		}
	}

	return nil
}
