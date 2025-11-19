package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func pointer[T any](v T) *T {
	return &v
}

func TestUpdateMetric(t *testing.T) {
	type args struct {
		uri  string
		body model.Metrics
	}

	type want struct {
		code       int
		metricType string
	}

	r := chi.NewRouter()
	s := repository.New()

	router := InitRoutes(r, s)

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success counter update",
			args: args{
				uri: "/update/",
				body: model.Metrics{
					ID:    "TestMetric",
					MType: model.Counter,
					Value: pointer(10.2),
					Delta: pointer(int64(10)),
				},
			},
			want: want{
				metricType: model.Counter,
				code:       http.StatusOK,
			},
		},
		{
			name: "success gauge update",
			args: args{
				uri: "/update/",
				body: model.Metrics{
					ID:    "TestMetric",
					MType: model.Gauge,
					Value: pointer(10.2),
				},
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "failed with type error",
			args: args{
				uri: "/update/",
				body: model.Metrics{
					ID:    "TestMetric",
					MType: "wrongType",
					Value: pointer(10.2),
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			var err error
			if tt.name == "invalid JSON" {
				bodyBytes = []byte("invalid json")
			} else {
				bodyBytes, err = json.Marshal(tt.args.body)
				if err != nil {
					t.Fatalf("cannot marshal body: %v", err)
				}
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, tt.args.uri, bytes.NewReader(bodyBytes))

			router.ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
