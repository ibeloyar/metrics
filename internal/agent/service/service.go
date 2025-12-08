package service

import (
	"bytes"
	"compress/gzip"
	"context"
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

func NewService(addr string) *Service {
	return &Service{
		client: &http.Client{},
		addr:   addr,
	}
}

func (s *Service) SendAllMetrics(metrics []SendMetricBody) error {
	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	return s.doSendAllWithRetry(bodyBytes)
}

func (s *Service) doSendAllWithRetry(body []byte) error {
	response, err := s.doSendAllGzip(body)
	if err != nil {
		var lastErr error

		lastErr = err

		for attempt := 0; attempt < 3; attempt++ {
			delay := getDelayFromAttempt(attempt)
			time.Sleep(delay)

			response, err = s.doSendAllGzip(body)
			if err != nil {
				lastErr = err
			}
			response.Body.Close()

			return nil
		}

		return lastErr
	}
	response.Body.Close()

	return nil
}

func (s *Service) doSendAllGzip(body []byte) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/updates/", s.addr), bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")

	response, err := s.client.Do(request)

	return response, err
}

func getDelayFromAttempt(attempt int) time.Duration {
	switch attempt {
	case 0:
		return 1 * time.Second
	case 1:
		return 3 * time.Second
	case 2:
		fallthrough
	default:
		return 5 * time.Second
	}
}
