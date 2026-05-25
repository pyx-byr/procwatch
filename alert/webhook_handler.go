package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookHandler sends alert events to a remote HTTP endpoint as JSON payloads.
type WebhookHandler struct {
	URL    string
	client *http.Client
}

// NewWebhookHandler creates a WebhookHandler that POSTs alerts to the given URL.
func NewWebhookHandler(url string, timeout time.Duration) *WebhookHandler {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &WebhookHandler{
		URL: url,
		client: &http.Client{Timeout: timeout},
	}
}

// webhookPayload is the JSON body sent to the webhook endpoint.
type webhookPayload struct {
	Timestamp string  `json:"timestamp"`
	Process   string  `json:"process"`
	PID       int     `json:"pid"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Message   string  `json:"message"`
}

// Handle serialises the Event and POSTs it to the configured webhook URL.
func (w *WebhookHandler) Handle(e Event) error {
	payload := webhookPayload{
		Timestamp: e.Timestamp.UTC().Format(time.RFC3339),
		Process:   e.Process,
		PID:       e.PID,
		Metric:    e.Metric,
		Value:     e.Value,
		Threshold: e.Threshold,
		Message:   e.Message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post to %s: %w", w.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, w.URL)
	}
	return nil
}
