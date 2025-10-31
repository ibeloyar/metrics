package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ibeloyar/metrics/internal/repository"
	"github.com/stretchr/testify/assert"

	"github.com/ibeloyar/metrics/internal/model"
)

func TestUpdateMetric(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	type want struct {
		code       int
		metricType string
	}

	handlers := InitHandlers(repository.Repository{
		Metrics: make(map[string]model.Metrics),
	})
	
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{type}/{name}/{value}", handlers.UpdateMetric)

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success counter update",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/counter/name/1", nil),
			},
			want: want{
				metricType: model.Counter,
				code:       http.StatusOK,
			},
		},

		{
			name: "success gauge update",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/gauge/name/1", nil),
			},
			want: want{
				metricType: model.Gauge,
				code:       http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux.ServeHTTP(tt.args.w, tt.args.r)

			res := tt.args.w.Result()

			gotMetricType := tt.args.r.PathValue("type")

			assert.Equal(t, tt.want.metricType, gotMetricType)
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
