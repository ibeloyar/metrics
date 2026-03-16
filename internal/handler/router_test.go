package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockHandlers struct {
	calls []string
}

func (m *mockHandlers) GetMetric(w http.ResponseWriter, r *http.Request) {
	m.calls = append(m.calls, "GetMetric")
}
func (m *mockHandlers) GetMetricQuery(w http.ResponseWriter, r *http.Request) {
	m.calls = append(m.calls, "GetMetricQuery")
}
func (m *mockHandlers) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	m.calls = append(m.calls, "UpdateMetric")
}
func (m *mockHandlers) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	m.calls = append(m.calls, "UpdateMetrics")
}
func (m *mockHandlers) UpdateMetricQuery(w http.ResponseWriter, r *http.Request) {
	m.calls = append(m.calls, "UpdateMetricQuery")
}
func (m *mockHandlers) GetMetricsPage(w http.ResponseWriter, r *http.Request) {
	m.calls = append(m.calls, "GetMetricsPage")
}
func (m *mockHandlers) Ping(w http.ResponseWriter, r *http.Request) {
	m.calls = append(m.calls, "Ping")
}

func TestInitRoutes(t *testing.T) {
	tests := []struct {
		name      string
		pprofFlag bool
		routes    []struct{ method, path string }
	}{
		{
			name:      "no_pprof",
			pprofFlag: false,
			routes: []struct{ method, path string }{
				{"GET", "/"},
				{"GET", "/ping"},
				{"POST", "/value/"},
				{"GET", "/value/gauge/cpu"},
				{"POST", "/update/"},
				{"POST", "/updates/"},
				{"POST", "/update/counter/clicks/1"},
			},
		},
		{
			name:      "with_pprof",
			pprofFlag: true,
			routes: []struct{ method, path string }{
				{"GET", "/"},
				{"GET", "/debug/pprof"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			handlers := &mockHandlers{}

			InitRoutes(r, handlers, tt.pprofFlag)

			for _, route := range tt.routes {
				req, _ := http.NewRequest(route.method, route.path, nil)
				rr := httptest.NewRecorder()
				r.ServeHTTP(rr, req)

				// The route must respond with 200-399 (handler called)
				if rr.Code >= http.StatusBadRequest {
					t.Errorf("%s %s failed: status %d", route.method, route.path, rr.Code)
				}
			}

			// pprof
			reqPprof, _ := http.NewRequest("GET", "/debug/pprof", nil)
			rrPprof := httptest.NewRecorder()
			r.ServeHTTP(rrPprof, reqPprof)

			if tt.pprofFlag {
				if rrPprof.Code >= http.StatusBadRequest {
					t.Error("pprof route should work when pprofFlag=true")
				}
			}
		})
	}
}
