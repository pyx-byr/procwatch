package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookHandler_Success(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := NewWebhookHandler(ts.URL, time.Second)
	e := Event{
		Timestamp: time.Now(),
		Process:   "myapp",
		PID:       42,
		Metric:    "cpu",
		Value:     95.5,
		Threshold: 80.0,
		Message:   "CPU threshold exceeded",
	}

	if err := h.Handle(e); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if received["process"] != "myapp" {
		t.Errorf("expected process=myapp, got %v", received["process"])
	}
	if received["metric"] != "cpu" {
		t.Errorf("expected metric=cpu, got %v", received["metric"])
	}
	if received["pid"].(float64) != 42 {
		t.Errorf("expected pid=42, got %v", received["pid"])
	}
}

func TestWebhookHandler_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	h := NewWebhookHandler(ts.URL, time.Second)
	e := Event{Timestamp: time.Now(), Process: "svc", Metric: "mem"}

	if err := h.Handle(e); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestWebhookHandler_ConnectionRefused(t *testing.T) {
	h := NewWebhookHandler("http://127.0.0.1:1", 500*time.Millisecond)
	e := Event{Timestamp: time.Now(), Process: "svc", Metric: "cpu"}

	if err := h.Handle(e); err == nil {
		t.Fatal("expected error for refused connection, got nil")
	}
}

func TestNewWebhookHandler_DefaultTimeout(t *testing.T) {
	h := NewWebhookHandler("http://example.com", 0)
	if h.client.Timeout != 5*time.Second {
		t.Errorf("expected default timeout 5s, got %v", h.client.Timeout)
	}
}
