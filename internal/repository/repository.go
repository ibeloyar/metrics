package repository

import (
	"github.com/ibeloyar/metrics/internal/model"
)

type Repository struct {
	Metrics map[string]model.Metrics
}

func New() *Repository {
	return &Repository{
		Metrics: make(map[string]model.Metrics),
	}
}

func (r *Repository) SetMetric(name, metricType string, value float64) error {
	r.Metrics[name] = model.Metrics{
		ID:    name,
		MType: metricType,
		Value: &value,
		Delta: nil,
		Hash:  "",
	}

	return nil
}

func (r *Repository) GetMetric(name string) *model.Metrics {
	v, ok := r.Metrics[name]
	if !ok {
		return nil
	}
	return &v
}

func (r *Repository) GetMetrics() map[string]model.Metrics {
	return r.Metrics
}
