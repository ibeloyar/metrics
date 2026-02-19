package gzip

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareNoCompression(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "image/png") // NOT json/html

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	rr := httptest.NewRecorder()
	Middleware(handler).ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "hello" {
		t.Errorf("Expected 'hello', got %q", rr.Body.String())
	}
}

func TestMiddlewareAcceptJSONNoGzip(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")
	// NOT Accept-Encoding: gzip

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("direct response"))
	})

	rr := httptest.NewRecorder()
	Middleware(handler).ServeHTTP(rr, req)

	if rr.Body.String() != "direct response" {
		t.Errorf("Expected direct response, got %q", rr.Body.String())
	}
}

func TestMiddlewareRequestNoCompression(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader("plain body"))
	req.Header.Set("Content-Type", "text/plain") // NOT json/html

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "plain body" {
			t.Errorf("Expected original body, got %s", body)
		}
	})

	rr := httptest.NewRecorder()
	Middleware(handler).ServeHTTP(rr, req)
}

func TestMiddlewareGzipAcceptEnabled(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("compressible content"))
	})

	rr := httptest.NewRecorder()
	Middleware(handler).ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Expected 200, got %d", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Error("Handler must be executed")
	}
}

func TestMiddlewareRequestGzipEnabled(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(""))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = 0

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler works"))
	})

	rr := httptest.NewRecorder()
	Middleware(handler).ServeHTTP(rr, req)

	// 500 = newCompressReader("") return error
	if rr.Code != 500 {
		t.Errorf("Expected 500 (newCompressReader error), got %d", rr.Code)
	}

	if handlerCalled {
		t.Error("Handler should NOT be called when newCompressReader fails.")
	}
}

func TestMiddlewareWriteHeaderStatus(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Encoding", "gzip")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "value")
		w.WriteHeader(201)
		w.Write([]byte("created"))
	})

	rr := httptest.NewRecorder()
	Middleware(handler).ServeHTTP(rr, req)

	if rr.Code != 201 {
		t.Errorf("Expected 201, got %d", rr.Code)
	}
	if rr.Header().Get("X-Test") != "value" {
		t.Error("Headers must pass")
	}
}

func TestMiddlewareMixedConditions(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader("body"))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "image/png") // not json/html

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "body" { // NOT decompressed
			t.Errorf("Expected original body, got %s", body)
		}
	})

	rr := httptest.NewRecorder()
	Middleware(handler).ServeHTTP(rr, req)
}
