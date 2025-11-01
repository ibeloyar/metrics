package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetMetric - require test, because if this function does not work, all other tests are useless
func TestSetMetric(t *testing.T) {
	storage := New()

	t.Run("success set metric by name", func(t *testing.T) {
		metricName := "test_metric"
		metricType := "counter"
		metricValue := 2.05

		err := storage.SetMetric(metricName, metricType, metricValue)

		require.Nil(t, err)
		require.Equal(t, len(storage.metrics), 1)

		metric, ok := storage.metrics[metricName]
		if !ok {
			require.Error(t, errors.New("added metric not found"))
		}

		require.Equal(t, metricName, metric.ID)
		require.Equal(t, metricType, metric.MType)
		require.Equal(t, metricValue, *metric.Value)
	})
}

func TestGetMetric(t *testing.T) {
	storage := New()

	metricName := "test_metric"
	metricType := "counter"
	metricValue := 2.05

	err := storage.SetMetric(metricName, metricType, metricValue)
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

func TestGetMetrics(t *testing.T) {
	storage := New()

	metricNames := []string{"one", "two", "three"}

	for i, v := range metricNames {
		err := storage.SetMetric(v, "gauge", float64(i)+0.01)
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
