package monitor

import (
	"testing"
	"time"
)

func makeSample(name string, cpu, mem float64) Sample {
	return Sample{
		Timestamp:  time.Now(),
		PID:        1234,
		Name:       name,
		CPUPercent: cpu,
		MemoryMB:   mem,
	}
}

func TestNewHistory_DefaultWindow(t *testing.T) {
	h := NewHistory(0)
	if h.window != 60 {
		t.Fatalf("expected default window 60, got %d", h.window)
	}
}

func TestHistory_AddAndLatest(t *testing.T) {
	h := NewHistory(10)
	h.Add(makeSample("nginx", 10.0, 50.0))
	h.Add(makeSample("nginx", 20.0, 60.0))

	s, ok := h.Latest("nginx")
	if !ok {
		t.Fatal("expected a sample, got none")
	}
	if s.CPUPercent != 20.0 {
		t.Errorf("expected latest CPU 20.0, got %f", s.CPUPercent)
	}
}

func TestHistory_Latest_Missing(t *testing.T) {
	h := NewHistory(10)
	_, ok := h.Latest("ghost")
	if ok {
		t.Error("expected no sample for unknown process")
	}
}

func TestHistory_WindowEviction(t *testing.T) {
	h := NewHistory(3)
	for i := 0; i < 5; i++ {
		h.Add(makeSample("svc", float64(i), float64(i)))
	}
	samples := h.All("svc")
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples after eviction, got %d", len(samples))
	}
	// oldest retained should be index 2 (cpu=2)
	if samples[0].CPUPercent != 2.0 {
		t.Errorf("expected oldest CPU 2.0, got %f", samples[0].CPUPercent)
	}
}

func TestHistory_AverageCPU(t *testing.T) {
	h := NewHistory(10)
	h.Add(makeSample("app", 10.0, 0))
	h.Add(makeSample("app", 20.0, 0))
	h.Add(makeSample("app", 30.0, 0))

	avg := h.AverageCPU("app")
	if avg != 20.0 {
		t.Errorf("expected avg CPU 20.0, got %f", avg)
	}
}

func TestHistory_AverageMemory(t *testing.T) {
	h := NewHistory(10)
	h.Add(makeSample("app", 0, 100.0))
	h.Add(makeSample("app", 0, 200.0))

	avg := h.AverageMemory("app")
	if avg != 150.0 {
		t.Errorf("expected avg memory 150.0, got %f", avg)
	}
}

func TestHistory_AverageCPU_Empty(t *testing.T) {
	h := NewHistory(10)
	if h.AverageCPU("nobody") != 0 {
		t.Error("expected 0 for empty history")
	}
}

func TestHistory_All_ReturnsCopy(t *testing.T) {
	h := NewHistory(10)
	h.Add(makeSample("svc", 5.0, 10.0))
	out := h.All("svc")
	out[0].CPUPercent = 999

	s, _ := h.Latest("svc")
	if s.CPUPercent == 999 {
		t.Error("All() should return a copy, not a reference")
	}
}
