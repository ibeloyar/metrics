package service

import (
	"net/http"

	"github.com/ibeloyar/metrics/internal/model"
)

type Storage interface {
	Ping() error
	GetMetric(name string) *model.Metrics
	GetMetrics() map[string]model.Metrics
	SetMetric(metric model.Metrics) error
	IncrementCountMetricValue(name string, delta *int64) error

	Shutdown() error
}

type Service struct {
	storage Storage
}

func New(s Storage) *Service {
	return &Service{
		storage: s,
	}
}

func (s *Service) SetMetric(metric model.Metrics) *model.APIError {
	if !s.IsValidMetricType(metric.MType) {
		return &model.APIError{
			Code:    http.StatusBadRequest,
			Message: "invalid metric type",
		}
	}

	switch metric.MType {
	case model.Gauge:
		err := s.storage.SetMetric(metric)
		if err != nil {
			return &model.APIError{
				Code:    http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			}
		}
		return nil
	case model.Counter:
		err := s.storage.IncrementCountMetricValue(metric.ID, metric.Delta)
		if err != nil {
			return &model.APIError{
				Code:    http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
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
	metric := s.storage.GetMetric(name)
	if metric == nil {
		return nil, &model.APIError{
			Code:    http.StatusNotFound,
			Message: "metric not found",
		}
	}

	return metric, nil
}

func (s *Service) GetMetrics() ([]model.Metrics, *model.APIError) {
	result := make([]model.Metrics, 0)
	metrics := s.storage.GetMetrics()

	for _, v := range metrics {
		result = append(result, v)
	}

	return result, nil
}

func (s *Service) Ping() error {
	err := s.storage.Ping()
	if err != nil {
		return err
	}

	return nil
}
