package service

import (
	"net"
	"net/http"
	"time"

	"github.com/ibeloyar/metrics/internal/audit"
	"github.com/ibeloyar/metrics/internal/model"
)

type Storage interface {
	Ping() error
	GetMetric(name string) *model.Metrics
	GetMetrics() map[string]model.Metrics
	SetMetric(metric model.Metrics) error
	SetMetrics(metrics []model.Metrics) error
	IncrementCountMetricValue(name string, delta *int64) error

	Shutdown() error
}

type Service struct {
	storage      Storage
	auditSubject *audit.AuditSubject
}

func New(s Storage, a *audit.AuditSubject) *Service {
	return &Service{
		storage:      s,
		auditSubject: a,
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

func (s *Service) SetMetrics(metrics []model.Metrics, remoteAddr string) *model.APIError {
	err := s.storage.SetMetrics(metrics)
	if err != nil {
		return &model.APIError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	event := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metricNames(metrics),
		IPAddress: parseIP(remoteAddr),
	}
	s.auditSubject.NotifyAll(event)

	return nil
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

func metricNames(metrics []model.Metrics) []string {
	names := make([]string, 0, len(metrics))
	for _, m := range metrics {
		names = append(names, m.ID)
	}
	return names
}

func parseIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
