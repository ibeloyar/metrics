package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ibeloyar/metrics/internal/model"
)

const (
	filePermission    = 0o644
	filePermissionAll = 0o777
)

type MemStorage struct {
	metrics          map[string]model.Metrics
	mu               sync.RWMutex
	saveMetricTicker *time.Ticker
	fileStoragePath  string
	restore          bool
}

func New(fileStoragePath string, storeSaveInterval uint64, restore bool) *MemStorage {
	var saveMetricTicker *time.Ticker = nil

	if storeSaveInterval > 0 {
		saveMetricTicker = time.NewTicker(time.Duration(storeSaveInterval) * time.Second)
	}

	return &MemStorage{
		metrics:          make(map[string]model.Metrics),
		saveMetricTicker: saveMetricTicker,
		fileStoragePath:  fileStoragePath,
		restore:          restore,
	}
}

func (s *MemStorage) Init() error {
	if s.restore {
		err := s.restoreData()
		if err != nil {
			return err
		}
	}

	if s.saveMetricTicker != nil {
		s.startSavingMetrics()
	}

	return nil
}

func (s *MemStorage) startSavingMetrics() {
	go func() {
		for range s.saveMetricTicker.C {
			err := s.writeMetricsToFile()
			if err != nil {
				return
			}
		}
	}()
}

func (s *MemStorage) Close() {
	if s.saveMetricTicker != nil {
		s.saveMetricTicker.Stop()
	}
}

func (s *MemStorage) writeMetricsToFile() error {
	data, err := json.MarshalIndent(s.metrics, "", "    ")
	if err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Dir(s.fileStoragePath), filePermissionAll); err != nil && !os.IsExist(err) {
		return err
	}
	if err := os.WriteFile(s.fileStoragePath, data, filePermission); err != nil {
		return err
	}

	return nil
}

func (s *MemStorage) restoreData() error {
	data, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		defaultFileStorageData := make(map[string]model.Metrics)

		data, err = json.MarshalIndent(defaultFileStorageData, "", "    ")
		if err != nil {
			return err
		}
		if err := os.Mkdir(filepath.Dir(s.fileStoragePath), filePermissionAll); err != nil && !os.IsExist(err) {
			return err
		}
		if err := os.WriteFile(s.fileStoragePath, data, filePermission); err != nil {
			return err
		}
	}

	if err := json.Unmarshal(data, &s.metrics); err != nil {
		return err
	}

	return nil
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
			return s.writeMetricsToFile()
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
			return s.writeMetricsToFile()
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
		return s.writeMetricsToFile()
	}

	return nil
}
