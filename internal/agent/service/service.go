package service

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Service struct {
	client *http.Client
	addr   string
}

type SendMetricBody struct {
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // gauge || counter
	Delta *int64   `json:"delta,omitempty"` // metric value if MType counter
	Value *float64 `json:"value,omitempty"` // metric value if MType gauge
}

func pointer[T any](v T) *T {
	return &v
}

func NewService(addr string) *Service {
	return &Service{
		client: &http.Client{},
		addr:   addr,
	}
}

//func (s *Service) SendPollCounter(pollCounter int64) error {
//	bodyBytes, err := json.Marshal(SendMetricBody{
//		ID:    "PollCount",
//		MType: "counter",
//		Delta: pointer(pollCounter),
//	})
//	if err != nil {
//		return err
//	}
//
//	return s.doSendWithRetry(bodyBytes)
//}

//func (s *Service) SendRandomValue() error {
//	bodyBytes, err := json.Marshal(SendMetricBody{
//		ID:    "RandomValue",
//		MType: "gauge",
//		Value: pointer(rand.Float64()),
//	})
//	if err != nil {
//		return err
//	}
//
//	return s.doSendWithRetry(bodyBytes)
//}

//func (s *Service) SendGaugeMetric(name string, value float64) error {
//	bodyBytes, err := json.Marshal(SendMetricBody{
//		ID:    name,
//		MType: "gauge",
//		Value: pointer(value),
//	})
//	if err != nil {
//		return err
//	}
//
//	return s.doSendWithRetry(bodyBytes)
//}

func (s *Service) doSendWithRetry(body []byte) error {
	response, err := s.doSendGzip(body)
	if err != nil {
		time.Sleep(5 * time.Millisecond)

		response, err = s.doSendGzip(body)
		if err != nil {
			return err
		}
		response.Body.Close()

		return nil
	}
	response.Body.Close()

	return nil
}

func (s *Service) doSendGzip(body []byte) (*http.Response, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)

	_, err := gw.Write(body)
	if err != nil {
		return nil, err
	}

	err = gw.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/updates/", s.addr), bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")

	response, err := s.client.Do(request)

	return response, err
}

func (s *Service) SendAllValues(metrics []SendMetricBody) error {
	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	return s.doSendWithRetry(bodyBytes)
}
