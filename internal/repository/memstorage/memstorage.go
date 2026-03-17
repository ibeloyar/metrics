package memstorage

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ibeloyar/metrics/internal/model"
)

// FileStorage defines file persistence interface for metrics.
type FileStorage interface {
	Save(map[string]model.Metrics) error
	Load() (map[string]model.Metrics, error)
}

// MemStorage is thread-safe in-memory metrics storage with file persistence.
type MemStorage struct {
	metrics          map[string]model.Metrics
	saveMetricTicker *time.Ticker
	restore          bool

	fileStorage FileStorage

	mu sync.RWMutex
}

// New creates MemStorage instance.
//
// storeSaveInterval: 0 = save on every write, >0 = periodic save interval in time.Duration
// restore: true = load metrics from file on Init().
// fileStorage: JSON, YAML or other file implementation.
func New(fileStorage FileStorage, storeSaveInterval int, restore bool) *MemStorage {
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

// Init initializes storage.
//
// Loads metrics from file if restore=true (ignores os.ErrNotExist).
// Starts periodic save goroutine if storeSaveInterval > 0.
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

// SetInitMetrics replaces all metrics with provided map (used after restore).
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

// Shutdown stops ticker and saves all metrics to file.
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

// GetMetric returns single metric by ID or nil if not found.
func (s *MemStorage) GetMetric(name string) *model.Metrics {
	v, ok := s.metrics[name]
	if !ok {
		return nil
	}

	return &v
}

// GetMetrics returns shallow copy of all metrics.
func (s *MemStorage) GetMetrics() map[string]model.Metrics {
	return s.metrics
}

// SetMetric stores/updates single metric.
//
// Gauge: overwrites Value, sets Delta=nil.
// Counter: overwrites Delta, sets Value=nil.
// Unknown types return error.
// Saves immediately if no ticker configured.
func (s *MemStorage) SetMetric(metric *model.Metrics) error {
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

// SetMetrics stores/updates multiple metrics atomically.
//
// Gauge: overwrites Value.
// Counter: accumulates Delta (existing + new).
// Unknown types return error.
// Saves immediately if no ticker configured.
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
				return s.SetMetric(&model.Metrics{
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
			return fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}

	if s.saveMetricTicker == nil {
		metrics := s.GetMetrics()

		return s.fileStorage.Save(metrics)
	}

	return nil
}

// IncrementCountMetricValue increments counter delta.
//
// Creates counter if doesn't exist.
// Accumulates delta values.
// Saves immediately if no ticker configured.
func (s *MemStorage) IncrementCountMetricValue(name string, delta *int64) error {
	oldMetric := s.GetMetric(name)
	if oldMetric == nil {
		return s.SetMetric(&model.Metrics{
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

// Ping always returns nil (in-memory storage).
func (s *MemStorage) Ping() error {
	return nil
}
