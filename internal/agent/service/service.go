package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
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

func (s *Service) SendPollCounter(pollCounter int64) error {
	bodyStruct := SendMetricBody{
		ID:    "PollCount",
		MType: "counter",
		Delta: pointer(pollCounter),
	}

	bodyBytes, err := json.Marshal(bodyStruct)
	if err != nil {
		return err
	}

	response, err := s.client.Post(fmt.Sprintf("http://%s/update/", s.addr), "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	response.Body.Close()

	return nil
}

func (s *Service) SendRandomValue() error {
	bodyStruct := SendMetricBody{
		ID:    "RandomValue",
		MType: "gauge",
		Value: pointer(rand.Float64()),
	}

	bodyBytes, err := json.Marshal(bodyStruct)
	if err != nil {
		return err
	}

	response, err := s.client.Post(fmt.Sprintf("http://%s/update/", s.addr), "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	response.Body.Close()

	return nil
}

func (s *Service) SendGaugeMetric(name string, value float64) error {
	bodyStruct := SendMetricBody{
		ID:    name,
		MType: "gauge",
		Value: pointer(value),
	}

	bodyBytes, err := json.Marshal(bodyStruct)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/", s.addr), bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		time.Sleep(5 * time.Millisecond)

		request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/", s.addr), bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}

		request.Header.Set("Content-Type", "application/json")

		response, err = s.client.Do(request)
		if err != nil {
			return err
		}
	}
	defer response.Body.Close()

	return nil
}
