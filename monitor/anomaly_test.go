package monitor

import (
	"math"
	"testing"
)

func TestAnomalyDetector_NoData(t *testing.T) {
	d := NewAnomalyDetector(2.0, 10)
	result := d.Analyze("nginx", 50.0, 100.0)
	if result != nil {
		t.Errorf("expected nil result with no baseline data, got %v", result)
	}
}

func TestAnomalyDetector_DefaultThreshold(t *testing.T) {
	d := NewAnomalyDetector(0, 10) // 0 should default to 2.0
	if d.threshold != 2.0 {
		t.Errorf("expected default threshold 2.0, got %.2f", d.threshold)
	}
}

func TestAnomalyDetector_NoAnomaly(t *testing.T) {
	d := NewAnomalyDetector(2.0, 20)
	for i := 0; i < 10; i++ {
		d.Add("svc", 30.0, 200.0)
	}
	// value very close to mean — should not be anomalous
	result := d.Analyze("svc", 30.5, 201.0)
	if result == nil {
		t.Fatal("expected a result, got nil")
	}
	if result.CPUAnomaly {
		t.Errorf("unexpected CPU anomaly: z=%.2f", result.CPUZScore)
	}
	if result.MemAnomaly {
		t.Errorf("unexpected Mem anomaly: z=%.2f", result.MemZScore)
	}
}

func TestAnomalyDetector_CPUAnomaly(t *testing.T) {
	d := NewAnomalyDetector(2.0, 20)
	for i := 0; i < 10; i++ {
		d.Add("svc", 10.0, 200.0)
	}
	// spike far above mean
	result := d.Analyze("svc", 90.0, 200.0)
	if result == nil {
		t.Fatal("expected a result")
	}
	if !result.CPUAnomaly {
		t.Errorf("expected CPU anomaly, z=%.2f", result.CPUZScore)
	}
	if result.MemAnomaly {
		t.Errorf("unexpected Mem anomaly, z=%.2f", result.MemZScore)
	}
}

func TestAnomalyDetector_MemAnomaly(t *testing.T) {
	d := NewAnomalyDetector(2.0, 20)
	for i := 0; i < 10; i++ {
		d.Add("svc", 10.0, 100.0)
	}
	result := d.Analyze("svc", 10.0, 9999.0)
	if result == nil {
		t.Fatal("expected a result")
	}
	if !result.MemAnomaly {
		t.Errorf("expected Mem anomaly, z=%.2f", result.MemZScore)
	}
}

func TestZScore_ZeroStddev(t *testing.T) {
	z := zScore(50.0, 50.0, 0)
	if z != 0 {
		t.Errorf("expected 0 for zero stddev, got %.2f", z)
	}
}

func TestZScore_Positive(t *testing.T) {
	z := zScore(13.0, 10.0, 1.5)
	expected := 2.0
	if math.Abs(z-expected) > 0.001 {
		t.Errorf("expected z=%.3f, got z=%.3f", expected, z)
	}
}

func TestAnomalyResult_String(t *testing.T) {
	r := &AnomalyResult{
		ProcessName: "nginx",
		CPUAnomaly:  true,
		MemAnomaly:  false,
		CPUZScore:   3.14,
		MemZScore:   0.5,
	}
	s := r.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
