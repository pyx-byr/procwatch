package monitor

import (
	"testing"

	"github.com/procwatch/config"
)

func makeProc(name string, cpu, mem float64) config.Process {
	return config.Process{
		Name:         name,
		CPUThreshold: cpu,
		MemThreshold: mem,
	}
}

func TestCheckAlerts_NoViolation(t *testing.T) {
	stats := ProcessStats{PID: 1, Name: "myapp", CPUPercent: 10.0, MemoryMB: 50.0}
	proc := makeProc("myapp", 80.0, 200.0)
	alerts := CheckAlerts(stats, proc)
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestCheckAlerts_CPUViolation(t *testing.T) {
	stats := ProcessStats{PID: 42, Name: "worker", CPUPercent: 95.0, MemoryMB: 100.0}
	proc := makeProc("worker", 80.0, 512.0)
	alerts := CheckAlerts(stats, proc)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Kind != "cpu" {
		t.Errorf("expected cpu alert, got %s", alerts[0].Kind)
	}
	if alerts[0].PID != 42 {
		t.Errorf("expected PID 42, got %d", alerts[0].PID)
	}
}

func TestCheckAlerts_MemoryViolation(t *testing.T) {
	stats := ProcessStats{PID: 7, Name: "server", CPUPercent: 5.0, MemoryMB: 600.0}
	proc := makeProc("server", 90.0, 512.0)
	alerts := CheckAlerts(stats, proc)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Kind != "memory" {
		t.Errorf("expected memory alert, got %s", alerts[0].Kind)
	}
	if alerts[0].Value != 600.0 {
		t.Errorf("expected value 600.0, got %f", alerts[0].Value)
	}
}

func TestCheckAlerts_BothViolations(t *testing.T) {
	stats := ProcessStats{PID: 99, Name: "heavy", CPUPercent: 99.0, MemoryMB: 1024.0}
	proc := makeProc("heavy", 80.0, 512.0)
	alerts := CheckAlerts(stats, proc)
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(alerts))
	}
}

func TestCheckAlerts_ZeroThresholdsSkipped(t *testing.T) {
	stats := ProcessStats{PID: 3, Name: "idle", CPUPercent: 100.0, MemoryMB: 9999.0}
	proc := makeProc("idle", 0, 0)
	alerts := CheckAlerts(stats, proc)
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts when thresholds are zero, got %d", len(alerts))
	}
}

func TestNewCollector(t *testing.T) {
	cfg := &config.Config{}
	c := NewCollector(cfg)
	if c == nil {
		t.Fatal("expected non-nil collector")
	}
	if c.cfg != cfg {
		t.Error("collector cfg mismatch")
	}
}
