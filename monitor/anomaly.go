package monitor

import (
	"fmt"
	"math"
	"sync"
)

// AnomalyResult holds the outcome of an anomaly check for a single process.
type AnomalyResult struct {
	ProcessName string
	CPUAnomaly  bool
	MemAnomaly  bool
	CPUZScore   float64
	MemZScore   float64
}

// AnomalyDetector detects statistical anomalies in process resource usage
// using a Z-score approach over a rolling baseline.
type AnomalyDetector struct {
	mu        sync.Mutex
	threshold float64 // Z-score threshold, e.g. 2.0
	baseline  *BaselineTracker
}

// NewAnomalyDetector creates an AnomalyDetector with the given Z-score threshold
// and an underlying BaselineTracker using the specified window size.
func NewAnomalyDetector(zThreshold float64, windowSize int) *AnomalyDetector {
	if zThreshold <= 0 {
		zThreshold = 2.0
	}
	return &AnomalyDetector{
		threshold: zThreshold,
		baseline:  NewBaselineTracker(windowSize),
	}
}

// Add records a new sample for the given process.
func (d *AnomalyDetector) Add(name string, cpu, mem float64) {
	d.baseline.Add(name, cpu, mem)
}

// Analyze checks whether the given cpu/mem values are anomalous relative to
// the stored baseline for the process. Returns nil if insufficient data.
func (d *AnomalyDetector) Analyze(name string, cpu, mem float64) *AnomalyResult {
	d.mu.Lock()
	defer d.mu.Unlock()

	stats, ok := d.baseline.Stats(name)
	if !ok {
		return nil
	}

	cpuZ := zScore(cpu, stats.AvgCPU, stats.StdCPU)
	memZ := zScore(mem, stats.AvgMem, stats.StdMem)

	return &AnomalyResult{
		ProcessName: name,
		CPUAnomaly:  math.Abs(cpuZ) > d.threshold,
		MemAnomaly:  math.Abs(memZ) > d.threshold,
		CPUZScore:   cpuZ,
		MemZScore:   memZ,
	}
}

// String returns a human-readable summary of the anomaly result.
func (r *AnomalyResult) String() string {
	return fmt.Sprintf(
		"process=%s cpu_anomaly=%v cpu_z=%.2f mem_anomaly=%v mem_z=%.2f",
		r.ProcessName, r.CPUAnomaly, r.CPUZScore, r.MemAnomaly, r.MemZScore,
	)
}

func zScore(value, mean, stddev float64) float64 {
	if stddev == 0 {
		return 0
	}
	return (value - mean) / stddev
}
