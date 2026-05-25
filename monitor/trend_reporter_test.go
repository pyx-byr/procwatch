package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/user/procwatch/logger"
)

func newTrendReporterLogger() (*logger.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	l := logger.New(buf)
	return l, buf
}

func TestTrendReporter_Stop_Idempotent(t *testing.T) {
	h := NewHistory(5 * time.Second)
	a := NewTrendAnalyzer(0.5)
	l, _ := newTrendReporterLogger()
	tr := NewTrendReporter(h, a, l, 100*time.Millisecond, []string{"svc"})
	tr.Stop()
	tr.Stop() // must not panic
}

func TestTrendReporter_DefaultInterval(t *testing.T) {
	h := NewHistory(5 * time.Second)
	a := NewTrendAnalyzer(0.5)
	l, _ := newTrendReporterLogger()
	tr := NewTrendReporter(h, a, l, 0, []string{})
	if tr.interval != 30*time.Second {
		t.Fatalf("expected 30s default interval, got %v", tr.interval)
	}
}

func TestTrendReporter_EmitsTrendLog(t *testing.T) {
	h := NewHistory(30 * time.Second)
	for i := 0; i < 5; i++ {
		h.Add("svc", Sample{CPUPercent: float64(i * 10), MemRSS: uint64(i * 1024)})
	}
	a := NewTrendAnalyzer(0.5)
	l, buf := newTrendReporterLogger()
	tr := NewTrendReporter(h, a, l, 50*time.Millisecond, []string{"svc"})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	go tr.Run(ctx)
	<-ctx.Done()
	tr.Stop()

	if buf.Len() == 0 {
		t.Fatal("expected trend log output, got none")
	}
	dec := json.NewDecoder(buf)
	var entry map[string]interface{}
	if err := dec.Decode(&entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}
	if entry["msg"] != "trend" {
		t.Fatalf("expected msg=trend, got %v", entry["msg"])
	}
	if _, ok := entry["cpu_direction"]; !ok {
		t.Fatal("expected cpu_direction field in log")
	}
}

func TestTrendReporter_SkipsProcessWithFewSamples(t *testing.T) {
	h := NewHistory(30 * time.Second)
	h.Add("svc", Sample{CPUPercent: 10, MemRSS: 1024})
	a := NewTrendAnalyzer(0.5)
	l, buf := newTrendReporterLogger()
	tr := NewTrendReporter(h, a, l, 50*time.Millisecond, []string{"svc"})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	go tr.Run(ctx)
	<-ctx.Done()
	tr.Stop()

	if buf.Len() != 0 {
		t.Fatal("expected no log output for single sample")
	}
}
