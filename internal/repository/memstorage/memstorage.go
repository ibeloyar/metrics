package memstorage

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ibeloyar/metrics/internal/model"
)

type FileStorage interface {
	Save(map[string]model.Metrics) error
	Load() (map[string]model.Metrics, error)
}

type MemStorage struct {
	metrics          map[string]model.Metrics
	saveMetricTicker *time.Ticker
	restore          bool

	fileStorage FileStorage

	mu sync.RWMutex
}

func New(fileStorage FileStorage, storeSaveInterval uint64, restore bool) *MemStorage {
	var saveMetricTicker *time.Ticker = nil

	if storeSaveInterval > 0 {
		saveMetricTicker = time.NewTicker(time.Duration(storeSaveInterval) * time.Second)
	}

	return &MemStorage{
		fileStorage:      fileStorage,
		metrics:          make(map[string]model.Metrics),
		saveMetricTicker: saveMetricTicker,
		restore:          restore,
	}
}

func (s *MemStorage) Init() error {
	if s.restore {
		metrics, err := s.fileStorage.Load()
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if metrics != nil {
			s.SetInitMetrics(metrics)
		}
	}

	if s.saveMetricTicker != nil {
		s.startSavingMetrics()
	}

	return nil
}

func (s *MemStorage) SetInitMetrics(metrics map[string]model.Metrics) {
	s.metrics = metrics
}

func (s *MemStorage) startSavingMetrics() {
	go func() {
		for range s.saveMetricTicker.C {
			metrics := s.GetMetrics()

			err := s.fileStorage.Save(metrics)
			if err != nil {
				return
			}
		}
	}()
}

func (s *MemStorage) Shutdown() error {
	if s.saveMetricTicker != nil {
		s.saveMetricTicker.Stop()
	}

	metrics := s.GetMetrics()

	err := s.fileStorage.Save(metrics)
	if err != nil {
		return err
	}

	return nil
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

func (s *MemStorage) SetMetric(metric model.Metrics) error {
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

	if s.saveMetricTicker == nil {
		metrics := s.GetMetrics()

		return s.fileStorage.Save(metrics)
	}

	return nil
}

func (s *MemStorage) SetMetrics(metrics []model.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range metrics {
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
			oldMetric, ok := s.metrics[metric.ID]
			if !ok {
				return s.SetMetric(model.Metrics{
					ID:    metric.ID,
					MType: model.Counter,
					Value: nil,
					Delta: metric.Delta,
					Hash:  "",
				})
			}

			newDelta := metric.Delta

			if oldMetric.Delta != nil && metric.Delta != nil {
				v := *newDelta + *oldMetric.Delta
				newDelta = &v
			}

			s.metrics[metric.ID] = model.Metrics{
				ID:    metric.ID,
				MType: model.Counter,
				Value: nil,
				Delta: newDelta,
				Hash:  "",
			}
		default:
			fmt.Printf("unknown metric type: %s", metric.MType)
			return fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}

	if s.saveMetricTicker == nil {
		metrics := s.GetMetrics()

		return s.fileStorage.Save(metrics)
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

	if s.saveMetricTicker == nil {
		metrics := s.GetMetrics()

		return s.fileStorage.Save(metrics)
	}

	return nil
}

func (s *MemStorage) Ping() error {
	return nil
}
