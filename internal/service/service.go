package service

import (
	"net/http"

	"github.com/ibeloyar/metrics/internal/model"
)

type MemStorage interface {
	GetMetric(name string) *model.Metrics
	GetMetrics() map[string]model.Metrics

	SetMetric(name, metricType string, value float64) error
	IncrementCountMetricValue(name string, value float64) error
}

type Service struct {
	storage MemStorage
}

func New(s MemStorage) *Service {
	return &Service{
		storage: s,
	}
}

func (s *Service) SetMetric(metricName, metricType string, metricValue float64) *model.APIError {
	if !s.IsValidMetricType(metricType) {
		return &model.APIError{
			Code:    http.StatusBadRequest,
			Message: "invalid metric type",
		}
	}

	switch metricType {
	case model.Gauge:
		err := s.storage.SetMetric(metricName, metricType, metricValue)
		if err != nil {
			return &model.APIError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		return nil
	case model.Counter:
		err := s.storage.IncrementCountMetricValue(metricName, metricValue)
		if err != nil {
			return &model.APIError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		return nil
	default:
		return &model.APIError{
			Code:    http.StatusBadRequest,
			Message: "invalid metric type",
		}
	}
}

func (s *Service) GetMetric(name string) (*model.Metrics, *model.APIError) {
	metrics := s.storage.GetMetric(name)
	if metrics == nil {
		return nil, &model.APIError{
			Code:    http.StatusNotFound,
			Message: "metric not found",
		}
	}

	return metrics, nil
}

func (s *Service) GetMetrics() ([]model.Metrics, *model.APIError) {
	result := make([]model.Metrics, 0)
	metrics := s.storage.GetMetrics()

	for _, v := range metrics {
		result = append(result, v)
	}

	return result, nil
}
