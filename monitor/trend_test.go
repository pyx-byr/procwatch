package monitor

import (
	"testing"
)

func TestLinearSlope_Empty(t *testing.T) {
	if got := linearSlope(nil); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestLinearSlope_SingleValue(t *testing.T) {
	if got := linearSlope([]float64{42}); got != 0 {
		t.Fatalf("expected 0 for single value, got %v", got)
	}
}

func TestLinearSlope_Rising(t *testing.T) {
	slope := linearSlope([]float64{1, 2, 3, 4, 5})
	if slope < 0.9 || slope > 1.1 {
		t.Fatalf("expected slope ~1.0, got %v", slope)
	}
}

func TestLinearSlope_Falling(t *testing.T) {
	slope := linearSlope([]float64{10, 8, 6, 4, 2})
	if slope > -1.9 || slope < -2.1 {
		t.Fatalf("expected slope ~-2.0, got %v", slope)
	}
}

func TestLinearSlope_Flat(t *testing.T) {
	slope := linearSlope([]float64{5, 5, 5, 5})
	if slope != 0 {
		t.Fatalf("expected 0 for flat series, got %v", slope)
	}
}

func TestTrendAnalyzer_Rising(t *testing.T) {
	ta := NewTrendAnalyzer(0.5)
	res := ta.Analyze("myapp", "cpu", []float64{10, 20, 30, 40, 50})
	if res.Direction != TrendRising {
		t.Fatalf("expected rising, got %s", res.Direction)
	}
	if res.Process != "myapp" || res.Metric != "cpu" {
		t.Fatalf("unexpected fields: %+v", res)
	}
}

func TestTrendAnalyzer_Falling(t *testing.T) {
	ta := NewTrendAnalyzer(0.5)
	res := ta.Analyze("myapp", "mem", []float64{80, 60, 40, 20})
	if res.Direction != TrendFalling {
		t.Fatalf("expected falling, got %s", res.Direction)
	}
}

func TestTrendAnalyzer_Stable(t *testing.T) {
	ta := NewTrendAnalyzer(2.0)
	res := ta.Analyze("myapp", "cpu", []float64{10, 10.1, 9.9, 10})
	if res.Direction != TrendStable {
		t.Fatalf("expected stable, got %s", res.Direction)
	}
}

func TestNewTrendAnalyzer_DefaultThreshold(t *testing.T) {
	ta := NewTrendAnalyzer(0)
	if ta.threshold != 0.5 {
		t.Fatalf("expected default threshold 0.5, got %v", ta.threshold)
	}
}

func TestTrendResult_String(t *testing.T) {
	r := TrendResult{Process: "svc", Metric: "cpu", Slope: 1.23, Direction: TrendRising}
	s := r.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}
