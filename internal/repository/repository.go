package repository

import (
	"fmt"
	"sync"

	"github.com/ibeloyar/metrics/internal/model"
)

type MemStorage struct {
	metrics map[string]model.Metrics
	mu      sync.RWMutex
}

func New() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]model.Metrics),
	}
}

func (s *MemStorage) GetMetric(name string) *model.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.metrics[name]
	if !ok {
		return nil
	}
	return &v
}

func (s *MemStorage) GetMetrics() map[string]model.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.metrics
}

func (s *MemStorage) SetMetric(metric model.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch metric.MType {
	case model.Gauge:
		s.metrics[metric.ID] = model.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
			Delta: nil,
			Hash:  "",
		}
	case model.Counter:
		s.metrics[metric.ID] = model.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: nil,
			Delta: metric.Delta,
			Hash:  "",
		}
	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}

	return nil
}

func (s *MemStorage) IncrementCountMetricValue(name string, delta *int64) error {
	oldMetric := s.GetMetric(name)
	if oldMetric == nil {
		return s.SetMetric(model.Metrics{
			ID:    name,
			MType: model.Counter,
			Value: nil,
			Delta: delta,
			Hash:  "",
		})
	}

	newDelta := delta

	if oldMetric.Delta != nil && delta != nil {
		v := *newDelta + *oldMetric.Delta
		newDelta = &v
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics[name] = model.Metrics{
		ID:    name,
		MType: model.Counter,
		Value: nil,
		Delta: newDelta,
		Hash:  "",
	}

	return nil
}
