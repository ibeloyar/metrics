package filestorage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ibeloyar/metrics/internal/model"
)

const (
	filePermission    = 0o644
	filePermissionAll = 0o777
)

// FileStorage implements JSON file persistence for metrics.
type FileStorage struct {
	path string
}

// New creates new FileStorage instance.
//
// Creates target directory automatically on first Save().
// path should be absolute path including filename (e.g. "/tmp/metrics.json").
func New(path string) *FileStorage {
	return &FileStorage{
		path: path,
	}
}

// Save serializes metrics map to pretty-printed JSON.
//
// Automatically creates parent directories if they don't exist.
// Uses 0644 permissions for file, 0777 for directories.
// Overwrites existing file atomically.
func (s *FileStorage) Save(metrics map[string]model.Metrics) error {
	data, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Dir(s.path), filePermissionAll); err != nil && !os.IsExist(err) {
		return err
	}
	if err := os.WriteFile(s.path, data, filePermission); err != nil {
		return err
	}

	return nil
}

// Load reads and deserializes metrics from JSON file.
//
// Returns os.ErrNotExist if file doesn't exist.
// Returns json.Unmarshal error if JSON is malformed.
func (s *FileStorage) Load() (map[string]model.Metrics, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]model.Metrics)

	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}
