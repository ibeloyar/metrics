package repository

import (
	"github.com/ibeloyar/metrics/internal/model"
)

type MemStorage struct {
	metrics map[string]model.Metrics
}

func New() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]model.Metrics),
	}
}

func (s *MemStorage) GetMetric(name string) *model.Metrics {
	v, ok := s.metrics[name]
	if !ok {
		return nil
	}
	return &v
}

func (s *MemStorage) GetMetrics() map[string]model.Metrics {
	return s.metrics
}

func (s *MemStorage) SetMetric(name, metricType string, value float64) error {
	s.metrics[name] = model.Metrics{
		ID:    name,
		MType: metricType,
		Value: &value,
		Delta: nil,
		Hash:  "",
	}

	return nil
}

func (s *MemStorage) IncrementCountMetricValue(name string, value float64) error {
	oldMetric := s.GetMetric(name)
	if oldMetric == nil {
		return s.SetMetric(name, model.Counter, value)
	}

	newValue := value

	if oldMetric.Value != nil {
		newValue = newValue + *oldMetric.Value
	}

	s.metrics[name] = model.Metrics{
		ID:    name,
		MType: model.Counter,
		Value: &newValue,
		Delta: nil,
		Hash:  "",
	}

	return nil
}
