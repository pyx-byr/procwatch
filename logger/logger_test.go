package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func newTestLogger() (*Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return New(buf), buf
}

func decodeEntry(t *testing.T, buf *bytes.Buffer) Entry {
	t.Helper()
	var e Entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}
	return e
}

func TestInfo(t *testing.T) {
	l, buf := newTestLogger()
	l.Info("started", map[string]interface{}{"pid": 42})
	e := decodeEntry(t, buf)
	if e.Level != LevelInfo {
		t.Errorf("expected INFO, got %s", e.Level)
	}
	if e.Message != "started" {
		t.Errorf("unexpected message: %s", e.Message)
	}
	if e.Fields["pid"] == nil {
		t.Error("expected pid field")
	}
}

func TestAlert(t *testing.T) {
	l, buf := newTestLogger()
	l.Alert("cpu threshold breached", map[string]interface{}{"process": "nginx", "cpu": 95.2})
	e := decodeEntry(t, buf)
	if e.Level != LevelAlert {
		t.Errorf("expected ALERT, got %s", e.Level)
	}
	if !strings.Contains(e.Message, "cpu") {
		t.Errorf("unexpected message: %s", e.Message)
	}
}

func TestWarn(t *testing.T) {
	l, buf := newTestLogger()
	l.Warn("process not found", nil)
	e := decodeEntry(t, buf)
	if e.Level != LevelWarn {
		t.Errorf("expected WARN, got %s", e.Level)
	}
	if e.Fields != nil {
		t.Error("expected nil fields")
	}
}

func TestError(t *testing.T) {
	l, buf := newTestLogger()
	l.Error("read failed", map[string]interface{}{"err": "permission denied"})
	e := decodeEntry(t, buf)
	if e.Level != LevelError {
		t.Errorf("expected ERROR, got %s", e.Level)
	}
}

func TestTimestampPresent(t *testing.T) {
	l, buf := newTestLogger()
	l.Info("check timestamp", nil)
	e := decodeEntry(t, buf)
	if e.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestDefaultWriterIsStdout(t *testing.T) {
	l := New(nil)
	if l.out == nil {
		t.Error("expected non-nil writer")
	}
}
