package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ibeloyar/metrics/internal/model"
	"github.com/ibeloyar/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
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
				code: http.StatusOK,
			},
		},
		{
			name: "failed with type error",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/wrong/name/1", nil),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router.ServeHTTP(tt.args.w, tt.args.r)

			res := tt.args.w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
