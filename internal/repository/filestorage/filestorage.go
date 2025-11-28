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

type FileStorage struct {
	path string
}

func New(path string) *FileStorage {
	return &FileStorage{
		path: path,
	}
}

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
