package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint_StatusOK(t *testing.T) {
	mux := http.NewServeMux()
	RegisterHealth(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %v", body["status"])
	}

	if _, ok := body["uptime"]; !ok {
		t.Fatal("expected 'uptime' field in response")
	}

	if _, ok := body["goroutines"]; !ok {
		t.Fatal("expected 'goroutines' field in response")
	}
}

func TestHealthEndpoint_ContentType(t *testing.T) {
	mux := http.NewServeMux()
	RegisterHealth(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}
}

func TestHealthEndpoint_MethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	RegisterHealth(mux)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Go 1.22+ ServeMux returns 405 for wrong method when pattern has method
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405 for POST, got %d", w.Code)
	}
}
