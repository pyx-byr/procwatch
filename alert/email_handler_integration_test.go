//go:build integration
// +build integration

package alert

import (
	"net"
	"testing"
	"time"
)

// TestEmailHandler_Handle_Success spins up a fake SMTP server and verifies
// that EmailHandler.Handle completes without error.
func TestEmailHandler_Handle_Success(t *testing.T) {
	addr, done := startFakeSMTP(t)
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	cfg := EmailConfig{
		SMTPHost: host,
		SMTPPort: port,
		Username: "user",
		Password: "pass",
		From:     "alert@example.com",
		To:       []string{"ops@example.com"},
	}
	h := NewEmailHandler(cfg)
	e := Event{
		ProcessName: "myapp",
		Metric:      "memory",
		Value:       512.0,
		Threshold:   256.0,
		Timestamp:   time.Now(),
	}
	if err := h.Handle(e); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("fake SMTP server did not finish in time")
	}
}
