package monitor

import (
	"testing"
	"time"
)

func TestBaselineTracker_DefaultWindow(t *testing.T) {
	bt := NewBaselineTracker(0)
	if bt.window != 10*time.Minute {
		t.Fatalf("expected default window 10m, got %v", bt.window)
	}
}

func TestBaselineTracker_NoSamples(t *testing.T) {
	bt := NewBaselineTracker(5 * time.Minute)
	_, ok := bt.Compute("myapp")
	if ok {
		t.Fatal("expected false for process with no samples")
	}
}

func TestBaselineTracker_SingleSample(t *testing.T) {
	bt := NewBaselineTracker(5 * time.Minute)
	bt.Add("myapp", 12.5, 256*1024*1024)
	stats, ok := bt.Compute("myapp")
	if !ok {
		t.Fatal("expected stats to be available")
	}
	if stats.AvgCPU != 12.5 {
		t.Errorf("expected AvgCPU 12.5, got %.2f", stats.AvgCPU)
	}
	if stats.Samples != 1 {
		t.Errorf("expected 1 sample, got %d", stats.Samples)
	}
}

func TestBaselineTracker_MultiSampleAvg(t *testing.T) {
	bt := NewBaselineTracker(5 * time.Minute)
	bt.Add("svc", 10.0, 100)
	bt.Add("svc", 20.0, 200)
	bt.Add("svc", 30.0, 300)
	stats, ok := bt.Compute("svc")
	if !ok {
		t.Fatal("expected stats")
	}
	if stats.AvgCPU != 20.0 {
		t.Errorf("expected AvgCPU 20.0, got %.2f", stats.AvgCPU)
	}
	if stats.AvgMemory != 200.0 {
		t.Errorf("expected AvgMemory 200.0, got %.2f", stats.AvgMemory)
	}
	if stats.Samples != 3 {
		t.Errorf("expected 3 samples, got %d", stats.Samples)
	}
}

func TestBaselineTracker_WindowEviction(t *testing.T) {
	bt := NewBaselineTracker(50 * time.Millisecond)
	bt.Add("proc", 99.0, 999)
	time.Sleep(80 * time.Millisecond)
	bt.Add("proc", 1.0, 1)
	stats, ok := bt.Compute("proc")
	if !ok {
		t.Fatal("expected stats after eviction")
	}
	if stats.Samples != 1 {
		t.Errorf("expected 1 sample after eviction, got %d", stats.Samples)
	}
	if stats.AvgCPU != 1.0 {
		t.Errorf("expected AvgCPU 1.0 after eviction, got %.2f", stats.AvgCPU)
	}
}

func TestBaselineTracker_Reset(t *testing.T) {
	bt := NewBaselineTracker(5 * time.Minute)
	bt.Add("app", 50.0, 512)
	bt.Reset("app")
	_, ok := bt.Compute("app")
	if ok {
		t.Fatal("expected no stats after reset")
	}
}

func TestBaselineStats_String(t *testing.T) {
	s := BaselineStats{AvgCPU: 23.5, AvgMemory: 256 * 1024 * 1024, Samples: 42}
	out := s.String()
	if out == "" {
		t.Fatal("expected non-empty string")
	}
}
