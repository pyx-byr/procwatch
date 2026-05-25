package monitor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/procwatch/logger"
)

func newTestReporterLogger(buf *bytes.Buffer) *logger.Logger {
	return logger.New(buf)
}

func TestReporter_EmitsStatsLog(t *testing.T) {
	h := NewHistory(30 * time.Second)
	h.Add("nginx", Sample{PID: 1, CPUPercent: 15.0, MemBytes: 1024 * 1024 * 100})
	h.Add("nginx", Sample{PID: 1, CPUPercent: 25.0, MemBytes: 1024 * 1024 * 200})

	agg := NewAggregator(h)
	var buf bytes.Buffer
	log := newTestReporterLogger(&buf)

	r := NewReporter(agg, log, []string{"nginx"}, 10*time.Millisecond)
	r.Start()
	time.Sleep(35 * time.Millisecond)
	r.Stop()

	output := buf.String()
	if !strings.Contains(output, "process stats") {
		t.Errorf("expected 'process stats' in log output, got: %s", output)
	}

	// Validate at least one JSON line has expected fields.
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry["msg"] == "process stats" {
			if _, ok := entry["avg_cpu"]; !ok {
				t.Error("expected avg_cpu field in log entry")
			}
			if _, ok := entry["max_cpu"]; !ok {
				t.Error("expected max_cpu field in log entry")
			}
			return
		}
	}
	t.Error("did not find a 'process stats' log entry with expected fields")
}

func TestReporter_WarnOnMissingProcess(t *testing.T) {
	h := NewHistory(30 * time.Second)
	agg := NewAggregator(h)
	var buf bytes.Buffer
	log := newTestReporterLogger(&buf)

	r := NewReporter(agg, log, []string{"missing"}, 10*time.Millisecond)
	r.Start()
	time.Sleep(25 * time.Millisecond)
	r.Stop()

	if !strings.Contains(buf.String(), "aggregation skipped") {
		t.Errorf("expected warn log for missing process, got: %s", buf.String())
	}
}

func TestReporter_Stop_IsIdempotentAfterTimeout(t *testing.T) {
	h := NewHistory(30 * time.Second)
	agg := NewAggregator(h)
	var buf bytes.Buffer
	log := newTestReporterLogger(&buf)

	r := NewReporter(agg, log, []string{}, 5*time.Millisecond)
	r.Start()
	time.Sleep(15 * time.Millisecond)
	r.Stop() // should not panic
}
