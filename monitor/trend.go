package monitor

import (
	"fmt"
	"math"
	"sync"
)

// TrendDirection indicates whether a metric is rising, falling, or stable.
type TrendDirection string

const (
	TrendRising  TrendDirection = "rising"
	TrendFalling TrendDirection = "falling"
	TrendStable  TrendDirection = "stable"
)

// TrendResult holds the computed slope and direction for a metric series.
type TrendResult struct {
	Process   string
	Metric    string
	Slope     float64
	Direction TrendDirection
}

func (t TrendResult) String() string {
	return fmt.Sprintf("process=%s metric=%s slope=%.4f direction=%s",
		t.Process, t.Metric, t.Slope, t.Direction)
}

// TrendAnalyzer computes linear regression slopes over recent samples.
type TrendAnalyzer struct {
	mu        sync.Mutex
	threshold float64 // minimum |slope| to be considered rising/falling
}

// NewTrendAnalyzer creates a TrendAnalyzer with the given stability threshold.
func NewTrendAnalyzer(threshold float64) *TrendAnalyzer {
	if threshold <= 0 {
		threshold = 0.5
	}
	return &TrendAnalyzer{threshold: threshold}
}

// Analyze computes the linear regression slope over the provided values.
// Values are assumed to be equally spaced in time.
func (ta *TrendAnalyzer) Analyze(process, metric string, values []float64) TrendResult {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	slope := linearSlope(values)
	dir := TrendStable
	if slope > ta.threshold {
		dir = TrendRising
	} else if slope < -ta.threshold {
		dir = TrendFalling
	}
	return TrendResult{Process: process, Metric: metric, Slope: slope, Direction: dir}
}

// linearSlope returns the least-squares slope for the given series.
func linearSlope(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i, v := range values {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if math.Abs(denom) < 1e-9 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}
