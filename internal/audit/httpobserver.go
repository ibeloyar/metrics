package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPAuditObserver struct {
	url string
	hc  *http.Client
}

func NewHTTPAuditObserver(url string) (*HTTPAuditObserver, error) {
	if url == "" {
		return nil, nil
	}

	return &HTTPAuditObserver{
		url: url,
		hc:  &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func (h *HTTPAuditObserver) Notify(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	resp, err := h.hc.Post(h.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
