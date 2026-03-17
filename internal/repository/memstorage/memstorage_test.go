package memstorage

import (
	"errors"
	"testing"

	"github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

var testConfig = config.Config{
	Addr:            ":8080",
	StoreInterval:   300,
	FileStoragePath: "mocks/metrics.json",
	Restore:         true,
}

func ptrFloat64(v float64) *float64 { return &v }
func ptrInt64(v int64) *int64       { return &v }

func TestMemStorage_Init(t *testing.T) {
	fs := filestorage.New("test.json")
	ms := New(fs, 10, true)
	err := ms.Init()
	require.NoError(t, err)
}

// TestSetMetric - require test, because if this function does not work, all other tests are useless
func TestSetMetric(t *testing.T) {

	fileStorage := filestorage.New(testConfig.FileStoragePath)
	storage := New(fileStorage, testConfig.StoreInterval, testConfig.Restore)

	t.Run("success set metric", func(t *testing.T) {
		metricName := "test_metric"
		metricType := model.Counter
		metricDelta := int64(2)

		err := storage.SetMetric(&model.Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &metricDelta,
		})

		require.Nil(t, err)
		require.Equal(t, len(storage.metrics), 1)

		metric, ok := storage.metrics[metricName]
		if !ok {
			require.Error(t, errors.New("added metric not found"))
		}

		require.Equal(t, metricName, metric.ID)
		require.Equal(t, metricType, metric.MType)
		require.Equal(t, metricDelta, *metric.Delta)
	})
}

func TestIncrementCountMetricValue(t *testing.T) {
	fileStorage := filestorage.New(testConfig.FileStoragePath)
	storage := New(fileStorage, testConfig.StoreInterval, testConfig.Restore)

	t.Run("success increment count metric", func(t *testing.T) {
		metricName := "test_metric"
		metricType := model.Counter
		metricDelta := int64(5)

		err := storage.IncrementCountMetricValue(metricName, &metricDelta)
		assert.Nil(t, err)

		metric, ok := storage.metrics[metricName]
		if !ok {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, metricName, metric.ID)
		assert.Equal(t, metricType, metric.MType)
		assert.Equal(t, metricDelta, *metric.Delta)

		addedDelta := int64(10)

		err = storage.IncrementCountMetricValue(metricName, &addedDelta)
		require.Nil(t, err)

		metric, ok = storage.metrics[metricName]
		if !ok {
			assert.Fail(t, err.Error())
		}

		require.Equal(t, metricName, metric.ID)
		require.Equal(t, metricType, metric.MType)
		require.Equal(t, metricDelta+addedDelta, *metric.Delta)
	})
}

func TestGetMetric(t *testing.T) {
	fileStorage := filestorage.New(testConfig.FileStoragePath)
	storage := New(fileStorage, testConfig.StoreInterval, testConfig.Restore)

	metricName := "test_metric"
	metricType := model.Gauge
	metricValue := 2.05

	err := storage.SetMetric(&model.Metrics{
		ID:    metricName,
		MType: metricType,
		Value: &metricValue,
	})
	if err != nil {
		assert.Fail(t, err.Error())
	}

	t.Run("success get metric by name", func(t *testing.T) {
		metric := storage.GetMetric(metricName)

		assert.NotNil(t, metric)
		assert.Equal(t, metricName, metric.ID)
		assert.Equal(t, metricType, metric.MType)
		assert.Equal(t, metricValue, *metric.Value)
	})

	t.Run("failed get metric by name (not found)", func(t *testing.T) {
		metric := storage.GetMetric("wrong name")

		assert.Nil(t, metric)
	})
}

func pointer[T any](v T) *T {
	return &v
}

func TestGetMetrics(t *testing.T) {
	fileStorage := filestorage.New(testConfig.FileStoragePath)
	storage := New(fileStorage, testConfig.StoreInterval, testConfig.Restore)

	metricNames := []string{"one", "two", "three"}

	for i, v := range metricNames {
		err := storage.SetMetric(&model.Metrics{
			ID:    v,
			MType: model.Gauge,
			Value: pointer(float64(i) + 0.01),
		})
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}

	require.Equal(t, len(storage.metrics), 3)

	t.Run("success get all metrics", func(t *testing.T) {
		metrics := storage.GetMetrics()

		assert.Equal(t, len(metricNames), len(metrics))
		assert.Equal(t, metricNames[0], metrics[metricNames[0]].ID)
		assert.Equal(t, 0.01, *metrics[metricNames[0]].Value)
		assert.Equal(t, metricNames[1], metrics[metricNames[1]].ID)
		assert.Equal(t, 1.01, *metrics[metricNames[1]].Value)
		assert.Equal(t, metricNames[2], metrics[metricNames[2]].ID)
		assert.Equal(t, 2.01, *metrics[metricNames[2]].Value)
	})
}

func TestMemStorage_Ping(t *testing.T) {
	storage := &MemStorage{}

	err := storage.Ping()

	assert.NoError(t, err)
	assert.Nil(t, err)
}

func TestMemStorage_SetMetrics_Gauge(t *testing.T) {
	s := &MemStorage{
		fileStorage: filestorage.New(testConfig.FileStoragePath),
		metrics:     make(map[string]model.Metrics),
	}

	metrics := []model.Metrics{{
		ID:    "gauge1",
		MType: model.Gauge,
		Value: ptrFloat64(123.45),
	}}

	err := s.SetMetrics(metrics)
	require.NoError(t, err)

	stored, exists := s.metrics["gauge1"]
	require.True(t, exists)
	assert.Equal(t, model.Gauge, stored.MType)
	assert.Equal(t, ptrFloat64(123.45), stored.Value)
	assert.Nil(t, stored.Delta)
}

func TestMemStorage_SetMetrics_CounterNew(t *testing.T) {
	s := &MemStorage{
		fileStorage: filestorage.New(testConfig.FileStoragePath),
		metrics:     make(map[string]model.Metrics),
	}

	metrics := []model.Metrics{{
		ID:    "counter1",
		MType: model.Counter,
		Delta: ptrInt64(42),
	}}

	err := s.SetMetrics(metrics)
	require.NoError(t, err)

	stored, exists := s.metrics["counter1"]
	require.True(t, exists)
	assert.Equal(t, model.Counter, stored.MType)
	assert.Equal(t, ptrInt64(42), stored.Delta)
	assert.Nil(t, stored.Value)
}

func TestMemStorage_SetMetrics_CounterIncrement(t *testing.T) {
	s := &MemStorage{
		fileStorage: filestorage.New(testConfig.FileStoragePath),
		metrics: map[string]model.Metrics{
			"counter1": {
				ID:    "counter1",
				MType: model.Counter,
				Delta: ptrInt64(10),
			},
		},
	}

	metrics := []model.Metrics{{
		ID:    "counter1",
		MType: model.Counter,
		Delta: ptrInt64(32),
	}}

	err := s.SetMetrics(metrics)
	require.NoError(t, err)

	stored, exists := s.metrics["counter1"]
	require.True(t, exists)
	assert.Equal(t, ptrInt64(42), stored.Delta)
}

func TestMemStorage_SetMetrics_MultipleMetrics(t *testing.T) {
	s := &MemStorage{
		metrics:     make(map[string]model.Metrics),
		fileStorage: filestorage.New(testConfig.FileStoragePath),
	}

	metrics := []model.Metrics{
		{
			ID:    "gauge1",
			MType: model.Gauge,
			Value: ptrFloat64(1.23),
		},
		{
			ID:    "counter1",
			MType: model.Counter,
			Delta: ptrInt64(5),
		},
	}

	err := s.SetMetrics(metrics)
	require.NoError(t, err)

	assert.Equal(t, 2, len(s.metrics))
	assertGauge(t, s.metrics["gauge1"], 1.23)
	assertCounter(t, s.metrics["counter1"], 5)
}

func TestMemStorage_SetMetrics_UnknownType(t *testing.T) {
	s := &MemStorage{
		fileStorage: filestorage.New(testConfig.FileStoragePath),
		metrics:     make(map[string]model.Metrics),
	}

	metrics := []model.Metrics{{
		ID:    "test",
		MType: "unknown",
		Delta: ptrInt64(1),
	}}

	err := s.SetMetrics(metrics)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown metric type: unknown")
	assert.Empty(t, s.metrics)
}

func TestMemStorage_SetMetrics_WithFileSave(t *testing.T) {
	s := &MemStorage{
		fileStorage: filestorage.New(testConfig.FileStoragePath),
		metrics:     make(map[string]model.Metrics),
	}

	metrics := []model.Metrics{{
		ID:    "gauge1",
		MType: model.Gauge,
		Value: ptrFloat64(99.9),
	}}

	err := s.SetMetrics(metrics)
	require.NoError(t, err)

	loaded, err := s.fileStorage.Load()
	require.NoError(t, err)
	assert.Equal(t, map[string]model.Metrics{
		"gauge1": {
			ID:    "gauge1",
			MType: model.Gauge,
			Value: ptrFloat64(99.9),
		},
	}, loaded)
}

func TestMemStorage_SetMetrics_EmptySlice(t *testing.T) {
	s := &MemStorage{
		fileStorage: filestorage.New(testConfig.FileStoragePath),
		metrics:     make(map[string]model.Metrics),
	}

	err := s.SetMetrics([]model.Metrics{})
	require.NoError(t, err)
	assert.Empty(t, s.metrics)
}

func assertGauge(t *testing.T, m model.Metrics, expected float64) {
	assert.Equal(t, model.Gauge, m.MType)
	require.NotNil(t, m.Value)
	assert.InDelta(t, expected, *m.Value, 0.001)
}

func assertCounter(t *testing.T, m model.Metrics, expected int64) {
	assert.Equal(t, model.Counter, m.MType)
	require.NotNil(t, m.Delta)
	assert.Equal(t, expected, *m.Delta)
}
