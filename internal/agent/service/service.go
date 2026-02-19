package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	hashHeaderName = "HashSHA256"

	maxSendAttempts     = 3
	firstRetryDuration  = 1 * time.Second
	secondRetryDuration = 3 * time.Second
	lastRetryDuration   = 5 * time.Second
)

type Service struct {
	client *http.Client
	addr   string
	key    string
}

type SendMetricBody struct {
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // gauge || counter
	Delta *int64   `json:"delta,omitempty"` // metric value if MType counter
	Value *float64 `json:"value,omitempty"` // metric value if MType gauge
}

func NewService(addr string, key string) *Service {
	client := retryablehttp.NewClient()
	client.RetryMax = maxSendAttempts
	client.RetryWaitMin = firstRetryDuration
	client.RetryWaitMax = lastRetryDuration
	client.Backoff = CustomBackoff
	standardClient := client.StandardClient()

	return &Service{
		client: standardClient,
		addr:   addr,
		key:    key,
	}
}

func (s *Service) SendMetrics(metrics []SendMetricBody) error {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)

	if _, err = gw.Write(bodyBytes); err != nil {
		return err
	}

	if err = gw.Close(); err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/updates/", s.addr), bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	if s.key != "" {
		request.Header.Set(hashHeaderName, GetHashBodySHA256(bodyBytes, s.key))
	}

	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	response.Body.Close()

	return nil
}

func CustomBackoff(min, max time.Duration, attemptNum int, _ *http.Response) time.Duration {
	switch attemptNum {
	case 0:
		return min
	case 1:
		return secondRetryDuration
	case 2:
		fallthrough
	default:
		return max
	}
}

func GetHashBodySHA256(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
