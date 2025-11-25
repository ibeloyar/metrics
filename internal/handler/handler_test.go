package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository/memstorage"
	"github.com/stretchr/testify/assert"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

func pointer[T any](v T) *T {
	return &v
}

var testServerConfig = config.Config{
	Addr:            ":8080",
	StoreInterval:   300,
	FileStoragePath: "data/metrics.json",
	Restore:         true,
}

func TestUpdateMetric(t *testing.T) {
	type args struct {
		uri           string
		requestMethod string
		body          *model.Metrics
	}

	type want struct {
		code       int
		metricType string
	}

	r := chi.NewRouter()
	s := memstorage.New(testServerConfig.FileStoragePath, testServerConfig.StoreInterval, testServerConfig.Restore)

	router := InitRoutes(r, s)

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success counter update",
			args: args{
				uri:           "/update/",
				requestMethod: http.MethodPost,
				body: &model.Metrics{
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
				uri:           "/update/",
				requestMethod: http.MethodPost,
				body: &model.Metrics{
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
			name: "failed gauge get value with Not Found 404",
			args: args{
				uri:           "/value/gauge/Alloc",
				requestMethod: http.MethodGet,
				body:          nil,
			},
			want: want{
				code:       http.StatusNotFound,
				metricType: model.Gauge,
			},
		},
		{
			name: "failed with type error",
			args: args{
				uri:           "/update/",
				requestMethod: http.MethodPost,
				body: &model.Metrics{
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
			var err error
			var bodyReader io.Reader

			if tt.args.body != nil {
				var bodyBytes []byte

				if tt.name == "invalid JSON" {
					bodyBytes = []byte("invalid json")
				} else {
					bodyBytes, err = json.Marshal(tt.args.body)
					if err != nil {
						t.Fatalf("cannot marshal body: %v", err)
					}
				}

				bodyReader = bytes.NewReader(bodyBytes)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.args.requestMethod, tt.args.uri, bodyReader)

			router.ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
