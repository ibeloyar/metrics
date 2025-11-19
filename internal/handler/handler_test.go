package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func (suite *Iteration7Suite) TestCollectAgentMetrics() {
	httpc := resty.New().SetHostURL(suite.serverAddress)

	tests := []struct {
		name   string
		method string
		value  float64
		delta  int64
		update int
		ok     bool
		static bool
	}{
		{method: "counter", name: "PollCount"},
		{method: "gauge", name: "RandomValue"},
		{method: "gauge", name: "Alloc"},
		{method: "gauge", name: "BuckHashSys", static: true},
		{method: "gauge", name: "Frees"},
		{method: "gauge", name: "GCCPUFraction", static: true},
		{method: "gauge", name: "GCSys", static: true},
		{method: "gauge", name: "HeapAlloc"},
		{method: "gauge", name: "HeapIdle"},
		{method: "gauge", name: "HeapInuse"},
		{method: "gauge", name: "HeapObjects"},
		{method: "gauge", name: "HeapReleased", static: true},
		{method: "gauge", name: "HeapSys", static: true},
		{method: "gauge", name: "LastGC", static: true},
		{method: "gauge", name: "Lookups", static: true},
		{method: "gauge", name: "MCacheInuse", static: true},
		{method: "gauge", name: "MCacheSys", static: true},
		{method: "gauge", name: "MSpanInuse", static: true},
		{method: "gauge", name: "MSpanSys", static: true},
		{method: "gauge", name: "Mallocs"},
		{method: "gauge", name: "NextGC", static: true},
		{method: "gauge", name: "NumForcedGC", static: true},
		{method: "gauge", name: "NumGC", static: true},
		{method: "gauge", name: "OtherSys", static: true},
		{method: "gauge", name: "PauseTotalNs", static: true},
		{method: "gauge", name: "StackInuse", static: true},
		{method: "gauge", name: "StackSys", static: true},
		{method: "gauge", name: "Sys", static: true},
		{method: "gauge", name: "TotalAlloc"},
	}

	req := httpc.R().SetHeader("Content-Type", "application/json")

	timer := time.NewTimer(time.Minute)

cont:
	for ok := 0; ok != len(tests); {
		select {
		case <-timer.C:
			break cont
		default:
		}
		for i, tt := range tests {
			if tt.ok {
				continue
			}
			var (
				resp *resty.Response
				err  error
			)
			time.Sleep(100 * time.Millisecond)

			var result Metrics
			resp, err = req.
				SetBody(&Metrics{
					ID:    tt.name,
					MType: tt.method,
				}).
				SetResult(&result).
				Post("/value/")

			dumpErr := suite.Assert().NoErrorf(err,
				"Ошибка при попытке сделать запрос с получением значения %s", tt.name)

			if resp.StatusCode() == http.StatusNotFound {
				continue
			}

			dumpErr = dumpErr && suite.Assert().Containsf(resp.Header().Get("Content-Type"), "application/json",
				"Заголовок ответа Content-Type содержит несоответствующее значение")
			dumpErr = dumpErr && suite.Assert().True(((result.MType == "gauge" && result.Value != nil) || (result.MType == "counter" && result.Delta != nil)),
				"Получен не однозначный результат (тип метода не соответствует возвращаемому значению) '%q %s'", req.Method, req.URL)
			dumpErr = dumpErr && suite.Assert().True(result.MType != "gauge" || result.Value != nil,
				"Получен не однозначный результат (возвращаемое значение value=nil не соответствет типу gauge) '%q %s'", req.Method, req.URL)
			dumpErr = dumpErr && suite.Assert().True(result.MType != "counter" || result.Delta != nil,
				"Получен не однозначный результат (возвращаемое значение delta=nil не соответствет типу counter) '%q %s'", req.Method, req.URL)
			dumpErr = dumpErr && suite.Assert().False(result.Delta == nil && result.Value == nil,
				"Получен результат без данных (Dalta == nil && Value == nil) '%q %s'", req.Method, req.URL)
			dumpErr = dumpErr && suite.Assert().False(result.Delta != nil && result.Value != nil,
				"Получен не однозначный результат (Dalta != nil && Value != nil) '%q %s'", req.Method, req.URL)
			dumpErr = dumpErr && suite.Assert().Equalf(http.StatusOK, resp.StatusCode(),
				"Несоответствие статус кода ответа ожидаемому в хендлере %q: %q ", req.Method, req.URL)
			dumpErr = dumpErr && suite.Assert().True(result.MType == "gauge" || result.MType == "counter",
				"Получен ответ с неизвестным значением типа: %q, '%q %s'", result.MType, req.Method, req.URL)

			if !dumpErr {
				dump := dumpRequest(req.RawRequest, true)
				suite.T().Logf("Оригинальный запрос:\n\n%s", dump)
				dump = dumpResponse(resp.RawResponse, true)
				suite.T().Logf("Оригинальный ответ:\n\n%s", dump)
				return
			}

			switch tt.method {
			case "gauge":
				if (tt.update != 0 && *result.Value != tt.value) || tt.static {
					tests[i].ok = true
					ok++
					suite.T().Logf("get %s: %q, value: %f", tt.method, tt.name, *result.Value)
				}
				tests[i].value = *result.Value
			case "counter":
				if (tt.update != 0 && *result.Delta != tt.delta) || tt.static {
					tests[i].ok = true
					ok++
					suite.T().Logf("get %s: %q, value: %d", tt.method, tt.name, *result.Delta)
				}
				tests[i].delta = *result.Delta
			}

			tests[i].update++
		}
	}
	for _, tt := range tests {
		suite.Run(tt.method+"/"+tt.name, func() {
			suite.Assert().Truef(tt.ok,
				"Отсутствует изменение метрики: %s, тип: %s", tt.name, tt.method)
		})
	}
}
