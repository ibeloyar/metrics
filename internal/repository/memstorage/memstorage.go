package memstorage

import (
	"fmt"
	"sync"
	"time"

	"github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository/filestorage"
)

type FileStorage interface {
	SaveMetricsToFile(metrics map[string]model.Metrics) error
	LoadMetricsFromFile() (map[string]model.Metrics, error)
}

type MemStorage struct {
	mu sync.RWMutex

	metrics          map[string]model.Metrics
	saveMetricTicker *time.Ticker

	fileStorage FileStorage
	restore     bool
}

func New(fileStoragePath string, storeSaveInterval uint64, restore bool) *MemStorage {
	var saveMetricTicker *time.Ticker = nil

	if storeSaveInterval > 0 {
		saveMetricTicker = time.NewTicker(time.Duration(storeSaveInterval) * time.Second)
	}

	return &MemStorage{
		fileStorage:      filestorage.New(fileStoragePath),
		metrics:          make(map[string]model.Metrics),
		saveMetricTicker: saveMetricTicker,
		restore:          restore,
	}
}

func (s *MemStorage) Init() error {
	if s.restore {
		metrics, err := s.fileStorage.LoadMetricsFromFile()
		if err != nil {
			return err
		}

		s.metrics = metrics
	}

	if s.saveMetricTicker != nil {
		s.startSavingMetrics()
	}

	return nil
}

func (s *MemStorage) startSavingMetrics() {
	go func() {
		for range s.saveMetricTicker.C {
			err := s.fileStorage.SaveMetricsToFile(s.metrics)
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

	return s.fileStorage.SaveMetricsToFile(s.metrics)
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

		if s.saveMetricTicker == nil {
			return s.fileStorage.SaveMetricsToFile(s.metrics)
		}
	case model.Counter:
		s.metrics[metric.ID] = model.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: nil,
			Delta: metric.Delta,
			Hash:  "",
		}

		if s.saveMetricTicker == nil {
			return s.fileStorage.SaveMetricsToFile(s.metrics)
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

	if s.saveMetricTicker == nil {
		return s.fileStorage.SaveMetricsToFile(s.metrics)
	}

	return nil
}
