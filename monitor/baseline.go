package monitor

import (
	"fmt"
	"sync"
	"time"
)

// BaselineSample holds a single CPU/memory observation for baseline calculation.
type BaselineSample struct {
	CPU    float64
	Memory float64
	At     time.Time
}

// BaselineStats holds computed baseline statistics for a process.
type BaselineStats struct {
	AvgCPU    float64
	AvgMemory float64
	Samples   int
}

// String returns a human-readable summary of the baseline stats.
func (b BaselineStats) String() string {
	return fmt.Sprintf("baseline: avg_cpu=%.2f%% avg_mem=%.2fMB samples=%d",
		b.AvgCPU, b.AvgMemory/1024/1024, b.Samples)
}

// BaselineTracker accumulates samples per process and computes rolling baselines.
type BaselineTracker struct {
	mu      sync.RWMutex
	window  time.Duration
	samples map[string][]BaselineSample
}

// NewBaselineTracker creates a BaselineTracker with the given rolling window.
func NewBaselineTracker(window time.Duration) *BaselineTracker {
	if window <= 0 {
		window = 10 * time.Minute
	}
	return &BaselineTracker{
		window:  window,
		samples: make(map[string][]BaselineSample),
	}
}

// Add records a new sample for the named process, evicting samples outside the window.
func (bt *BaselineTracker) Add(name string, cpu, memory float64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-bt.window)
	existing := bt.samples[name]
	filtered := existing[:0]
	for _, s := range existing {
		if s.At.After(cutoff) {
			filtered = append(filtered, s)
		}
	}
	filtered = append(filtered, BaselineSample{CPU: cpu, Memory: memory, At: now})
	bt.samples[name] = filtered
}

// Compute returns the BaselineStats for the named process.
// Returns false if no samples are available.
func (bt *BaselineTracker) Compute(name string) (BaselineStats, bool) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	samples := bt.samples[name]
	if len(samples) == 0 {
		return BaselineStats{}, false
	}
	var sumCPU, sumMem float64
	for _, s := range samples {
		sumCPU += s.CPU
		sumMem += s.Memory
	}
	n := float64(len(samples))
	return BaselineStats{
		AvgCPU:    sumCPU / n,
		AvgMemory: sumMem / n,
		Samples:   len(samples),
	}, true
}

// Reset clears all samples for the named process.
func (bt *BaselineTracker) Reset(name string) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	delete(bt.samples, name)
}
