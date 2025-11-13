package service

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

type Service struct {
	client *http.Client
	addr   string
}

func NewService(addr string) *Service {
	return &Service{
		client: &http.Client{
			Timeout: time.Second * 1,
		},
		addr: addr,
	}
}

func (s *Service) SendPollCounter(pollCounter int) error {
	response, err := s.client.Post(fmt.Sprintf("http://%s/update/counter/PollCount/%d", s.addr, pollCounter), "text/plain", nil)
	if err != nil {
		return err
	}
	response.Body.Close()

	return nil
}

func (s *Service) SendRandomValue() error {
	response, err := s.client.Post(fmt.Sprintf("http://%s/update/gauge/RandomValue/%f", s.addr, rand.Float64()), "text/plain", nil)
	if err != nil {
		return err
	}
	response.Body.Close()

	return nil
}

func (s *Service) SendGaugeMetric(name string, value float64) error {
	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://%s/update/gauge/%s/%f", s.addr, name, value),
		nil,
	)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "text/plain")

	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	response.Body.Close()

	return nil
}
