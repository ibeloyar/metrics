package service

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKey  = "test-key"
	testAddr = "localhost:8080"
)

// Helpers
func ptrFloat64(v float64) *float64 { return &v }
func ptrInt64(v int64) *int64       { return &v }

func TestNewService(t *testing.T) {
	svc := NewService(testAddr, testKey)

	assert.NotNil(t, svc.client)
	assert.Equal(t, testAddr, svc.addr)
	assert.Equal(t, testKey, svc.key)
}

func TestSendMetrics_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := []SendMetricBody{
			{ID: "test:gauge", MType: "gauge", Value: ptrFloat64(123.45)},
			{ID: "test:counter", MType: "counter", Delta: ptrInt64(1)},
		}

		bodyBytes, err := json.Marshal(metrics)
		require.NoError(t, err)
		expectedHash := GetHashBodySHA256(bodyBytes, testKey)

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/updates/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, expectedHash, r.Header.Get(hashHeaderName))

		// verify gzip content
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		gr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gr.Close()

		require.NoError(t, json.NewDecoder(gr).Decode(&metrics))
		assert.Len(t, metrics, 2)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewService(server.Listener.Addr().String(), testKey)

	metrics := []SendMetricBody{
		{
			ID:    "test:gauge",
			MType: "gauge",
			Value: ptrFloat64(123.45),
		},
		{
			ID:    "test:counter",
			MType: "counter",
			Delta: ptrInt64(1),
		},
	}

	err := svc.SendMetrics(metrics)
	assert.NoError(t, err)
}

func TestSendMetrics_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	svc := NewService(server.Listener.Addr().String(), "")

	err := svc.SendMetrics([]SendMetricBody{{ID: "test", MType: "gauge"}})
	assert.NoError(t, err)
}

func TestSendMetrics_NoKeyNoHash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "", r.Header.Get(hashHeaderName))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewService(server.Listener.Addr().String(), "")

	err := svc.SendMetrics([]SendMetricBody{{ID: "test", MType: "gauge"}})
	assert.NoError(t, err)
}

func TestCustomBackoff(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		min, max time.Duration
		expected time.Duration
	}{
		{
			name:     "first attempt",
			attempt:  0,
			min:      1 * time.Second,
			max:      5 * time.Second,
			expected: 1 * time.Second,
		},
		{
			name:     "second attempt",
			attempt:  1,
			min:      1 * time.Second,
			max:      5 * time.Second,
			expected: 3 * time.Second,
		},
		{
			name:     "third attempt",
			attempt:  2,
			min:      1 * time.Second,
			max:      5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "fourth attempt (max)",
			attempt:  3,
			min:      1 * time.Second,
			max:      5 * time.Second,
			expected: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CustomBackoff(tt.min, tt.max, tt.attempt, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetHashBodySHA256(t *testing.T) {
	data := []byte("test data")

	hash := GetHashBodySHA256(data, testKey)
	t.Logf("HMAC(%q, %q) = %s", testKey, "test data", hash)

	assert.Len(t, hash, 64)

	emptyHash := GetHashBodySHA256(data, "")
	assert.Len(t, emptyHash, 64)
}
