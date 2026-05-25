package monitor

import (
	"testing"
	"time"
)

func TestAggregator_NoSamples(t *testing.T) {
	h := NewHistory(30 * time.Second)
	a := NewAggregator(h)

	_, err := a.Compute("ghost")
	if err == nil {
		t.Fatal("expected error for missing process, got nil")
	}
}

func TestAggregator_SingleSample(t *testing.T) {
	h := NewHistory(30 * time.Second)
	a := NewAggregator(h)

	h.Add("web", Sample{PID: 42, CPUPercent: 10.0, MemBytes: 1024 * 1024 * 50})

	stats, err := a.Compute("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.AvgCPU != 10.0 {
		t.Errorf("expected AvgCPU=10.0, got %.2f", stats.AvgCPU)
	}
	if stats.MaxCPU != 10.0 {
		t.Errorf("expected MaxCPU=10.0, got %.2f", stats.MaxCPU)
	}
	if stats.SampleCount != 1 {
		t.Errorf("expected SampleCount=1, got %d", stats.SampleCount)
	}
	if stats.PID != 42 {
		t.Errorf("expected PID=42, got %d", stats.PID)
	}
}

func TestAggregator_MultiSample_Avg(t *testing.T) {
	h := NewHistory(30 * time.Second)
	a := NewAggregator(h)

	h.Add("api", Sample{PID: 7, CPUPercent: 20.0, MemBytes: 100})
	h.Add("api", Sample{PID: 7, CPUPercent: 40.0, MemBytes: 300})
	h.Add("api", Sample{PID: 7, CPUPercent: 60.0, MemBytes: 200})

	stats, err := a.Compute("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const wantAvgCPU = 40.0
	if stats.AvgCPU != wantAvgCPU {
		t.Errorf("expected AvgCPU=%.1f, got %.2f", wantAvgCPU, stats.AvgCPU)
	}
	if stats.MaxCPU != 60.0 {
		t.Errorf("expected MaxCPU=60.0, got %.2f", stats.MaxCPU)
	}
	const wantAvgMem = float64(600) / 3
	if stats.AvgMem != wantAvgMem {
		t.Errorf("expected AvgMem=%.1f, got %.2f", wantAvgMem, stats.AvgMem)
	}
	if stats.MaxMem != 300 {
		t.Errorf("expected MaxMem=300, got %.2f", stats.MaxMem)
	}
}

func TestAggregator_StatsString(t *testing.T) {
	s := Stats{
		ProcessName: "myapp",
		PID:         99,
		AvgCPU:      5.5,
		MaxCPU:      12.3,
		AvgMem:      1024 * 1024 * 200,
		MaxMem:      1024 * 1024 * 256,
		SampleCount: 10,
	}
	out := s.String()
	if out == "" {
		t.Error("expected non-empty string from Stats.String()")
	}
}
