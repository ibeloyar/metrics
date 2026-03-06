package filestorage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ibeloyar/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_Save_Load(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-metrics.json")

	storage := New(path)

	metrics := map[string]model.Metrics{
		"gauge1": {
			ID:    "gauge1",
			MType: "gauge",
			Value: func(v float64) *float64 { return &v }(123.456),
		},
		"counter1": {
			ID:    "counter1",
			MType: "counter",
			Delta: func(d int64) *int64 { return &d }(42),
		},
	}

	// Test Save
	err := storage.Save(metrics)
	require.NoError(t, err)

	// Check file created
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test Load
	loaded, err := storage.Load()
	require.NoError(t, err)
	assert.Equal(t, metrics, loaded)
}

func TestFileStorage_Load_NotExist(t *testing.T) {
	storage := New("/nonexistent/path.json")

	metrics, err := storage.Load()
	require.ErrorIs(t, err, os.ErrNotExist)
	assert.Nil(t, metrics)
}

func TestFileStorage_Load_InvalidJson(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")

	// Create invalid JSON file
	err := os.WriteFile(path, []byte("invalid json"), 0644)
	require.NoError(t, err)

	storage := New(path)
	metrics, err := storage.Load()

	require.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestFileStorage_Save_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")

	storage := New(path)

	metrics1 := map[string]model.Metrics{
		"first": {
			ID:    "first",
			MType: "gauge",
			Value: func(v float64) *float64 { return &v }(1.0),
		},
	}
	metrics2 := map[string]model.Metrics{
		"second": {
			ID:    "second",
			MType: "counter",
			Delta: func(d int64) *int64 { return &d }(100),
		},
	}

	// First save
	require.NoError(t, storage.Save(metrics1))

	// Overwrite
	require.NoError(t, storage.Save(metrics2))

	// Check only second metrics exist
	loaded, err := storage.Load()
	require.NoError(t, err)
	assert.Equal(t, metrics2, loaded)
	assert.NotContains(t, loaded, "first")
}

func TestFileStorage_DirAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")

	storage := New(path)
	metrics := map[string]model.Metrics{
		"test": {
			ID:    "test",
			MType: "gauge",
			Value: func(v float64) *float64 { return &v }(42.0),
		},
	}

	err := storage.Save(metrics)
	require.NoError(t, err)

	loaded, err := storage.Load()
	require.NoError(t, err)
	assert.Equal(t, metrics, loaded)
}

func TestFileStorage_Save_EmptyMetrics(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")

	storage := New(path)

	err := storage.Save(map[string]model.Metrics{})
	require.NoError(t, err)

	loaded, err := storage.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded)
}
