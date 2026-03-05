package service

import (
	"testing"

	"github.com/ibeloyar/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrInt64(v int64) *int64 {
	return &v
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func TestIsValidMetricType(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name   string
		mtype  string
		want   bool
		reason string
	}{
		{"valid counter", model.Counter, true, ""},
		{"valid gauge", model.Gauge, true, ""},
		{"invalid type", "unknown", false, ""},
		{"empty type", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.IsValidMetricType(tt.mtype)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateMetric_Valid(t *testing.T) {
	s := &Service{}

	tests := []model.Metrics{
		{
			MType: model.Counter,
			Delta: ptrInt64(42),
		},
		{
			MType: model.Gauge,
			Value: ptrFloat64(123.45),
		},
	}

	for _, metric := range tests {
		err := s.ValidateMetric(metric)
		assert.NoError(t, err)
	}
}

func TestValidateMetric_InvalidType(t *testing.T) {
	s := &Service{}
	metric := model.Metrics{
		MType: "unknown",
		Delta: ptrInt64(1),
	}

	err := s.ValidateMetric(metric)
	require.Error(t, err)
	assert.Equal(t, "invalid metric type", err.Error())
}

func TestValidateMetric_CounterNoDelta(t *testing.T) {
	s := &Service{}
	metric := model.Metrics{
		MType: model.Counter,
		// Delta == nil
	}

	err := s.ValidateMetric(metric)
	require.Error(t, err)
	assert.Equal(t, "delta must be set for counter", err.Error())
}

func TestValidateMetric_GaugeNoValue(t *testing.T) {
	s := &Service{}
	metric := model.Metrics{
		MType: model.Gauge,
		// Value == nil
	}

	err := s.ValidateMetric(metric)
	require.Error(t, err)
	assert.Equal(t, "value must be set for gauge", err.Error())
}

func TestValidateMetrics_AllValid(t *testing.T) {
	s := &Service{}
	metrics := []model.Metrics{
		{
			MType: model.Counter,
			Delta: ptrInt64(10),
		},
		{
			MType: model.Gauge,
			Value: ptrFloat64(99.9),
		},
	}

	err := s.ValidateMetrics(metrics)
	assert.NoError(t, err)
}

func TestValidateMetrics_FirstInvalid(t *testing.T) {
	s := &Service{}
	metrics := []model.Metrics{
		{
			MType: "invalid",
			Delta: ptrInt64(1),
		},
		{
			MType: model.Gauge,
			Value: ptrFloat64(1.0),
		},
	}

	err := s.ValidateMetrics(metrics)
	require.Error(t, err)
	assert.Equal(t, "invalid metric type", err.Error())
}

func TestValidateMetrics_MiddleInvalid(t *testing.T) {
	s := &Service{}
	metrics := []model.Metrics{
		{
			MType: model.Counter,
			Delta: ptrInt64(1),
		},
		{
			MType: model.Counter,
			// Delta == nil
		},
		{
			MType: model.Gauge,
			Value: ptrFloat64(1.0),
		},
	}

	err := s.ValidateMetrics(metrics)
	require.Error(t, err)
	assert.Equal(t, "delta must be set for counter", err.Error())
}

func TestValidateMetrics_EmptySlice(t *testing.T) {
	s := &Service{}
	err := s.ValidateMetrics([]model.Metrics{})
	assert.NoError(t, err)
}
