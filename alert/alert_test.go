package alert

import (
	"testing"
	"time"
)

func TestNewManager_NoHandlers(t *testing.T) {
	m := NewManager()
	if len(m.handlers) != 0 {
		t.Fatalf("expected 0 handlers, got %d", len(m.handlers))
	}
}

func TestRegister(t *testing.T) {
	m := NewManager()
	m.Register(func(e Event) {})
	m.Register(func(e Event) {})
	if len(m.handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(m.handlers))
	}
}

func TestEmit_DispatchesToAllHandlers(t *testing.T) {
	m := NewManager()
	var received []Event

	m.Register(func(e Event) { received = append(received, e) })
	m.Register(func(e Event) { received = append(received, e) })

	m.Emit("myapp", 42, "cpu", 95.5, 80.0, SeverityWarn)

	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
}

func TestEmit_EventFields(t *testing.T) {
	m := NewManager()
	var got Event
	m.Register(func(e Event) { got = e })

	before := time.Now().UTC()
	m.Emit("svc", 99, "memory", 512.0, 256.0, SeverityCritical)
	after := time.Now().UTC()

	if got.Process != "svc" {
		t.Errorf("process: want svc, got %s", got.Process)
	}
	if got.PID != 99 {
		t.Errorf("pid: want 99, got %d", got.PID)
	}
	if got.Metric != "memory" {
		t.Errorf("metric: want memory, got %s", got.Metric)
	}
	if got.Value != 512.0 {
		t.Errorf("value: want 512.0, got %.2f", got.Value)
	}
	if got.Threshold != 256.0 {
		t.Errorf("threshold: want 256.0, got %.2f", got.Threshold)
	}
	if got.Severity != SeverityCritical {
		t.Errorf("severity: want critical, got %s", got.Severity)
	}
	if got.Message == "" {
		t.Error("message should not be empty")
	}
	if got.Timestamp.Before(before) || got.Timestamp.After(after) {
		t.Error("timestamp out of expected range")
	}
}

func TestEmit_NoHandlers_NoPanic(t *testing.T) {
	m := NewManager()
	// Should not panic with zero handlers
	m.Emit("proc", 1, "cpu", 50.0, 40.0, SeverityWarn)
}
