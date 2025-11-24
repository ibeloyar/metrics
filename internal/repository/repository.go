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

// Задание по треку «Сервис сбора метрик и алертинга»
// Доработайте код сервера, чтобы он мог

// 1) с заданной периодичностью (STORE_INTERVAL) сохранять текущие значения метрик на диск в указанный файл (FILE_STORAGE_PATH)
// 2) на старте — опционально загружать сохранённые ранее значения (RESTORE true/false).

// Сервер должен принимать соответствующие параметры конфигурации через флаги и переменные окружения:
// Флаг -i, переменная окружения STORE_INTERVAL — интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной).
// Флаг -f, переменная окружения FILE_STORAGE_PATH — путь до файла, куда сохраняются текущие значения. Имя файла для значения по умолчанию придумайте сами.
// Флаг -r, переменная окружения RESTORE — булево значение (true/false), определяющее, следует ли загружать ранее сохранённые значения из указанного файла при старте сервера.
// Пример содержимого файла:
// [
//     {"id":"LastGC","type":"gauge","value":1257894000000000000},
//     {"id":"NumGC","type":"counter","delta":42},
//     ...
// ]
// Приоритет параметров сервера должен быть таким:
// Если указана переменная окружения, то используется она.
// Если нет переменной окружения, но есть флаг, то используется он.
// Если нет ни переменной окружения, ни флага, то используется значение по умолчанию.

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
		return fmt.Errorf("%s: %w", "Error", err)
	}
	if err := os.Mkdir(filepath.Dir(s.fileStoragePath), filePermissionAll); err != nil && !os.IsExist(err) {
		return fmt.Errorf("%s: %w", "Error", err)
	}
	if err := os.WriteFile(s.fileStoragePath, data, filePermission); err != nil {
		return fmt.Errorf("%s: %w", "Error", err)
	}

	return nil
}

func (s *MemStorage) restoreData() error {
	data, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		defaultFileStorageData := make(map[string]model.Metrics)

		data, err = json.MarshalIndent(defaultFileStorageData, "", "    ")
		if err != nil {
			return fmt.Errorf("%s: %w", "Error", err)
		}
		if err := os.Mkdir(filepath.Dir(s.fileStoragePath), filePermissionAll); err != nil && !os.IsExist(err) {
			return fmt.Errorf("%s: %w", "Error", err)
		}
		if err := os.WriteFile(s.fileStoragePath, data, filePermission); err != nil {
			return fmt.Errorf("%s: %w", "Error", err)
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
