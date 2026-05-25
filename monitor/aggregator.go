package monitor

import (
	"fmt"
	"time"
)

// Stats holds aggregated statistics for a process over a time window.
type Stats struct {
	ProcessName string
	PID         int32
	AvgCPU      float64
	MaxCPU      float64
	AvgMem      float64
	MaxMem      float64
	SampleCount int
	Window      time.Duration
	ComputedAt  time.Time
}

// String returns a human-readable summary of the stats.
func (s Stats) String() string {
	return fmt.Sprintf(
		"process=%s pid=%d samples=%d avg_cpu=%.2f%% max_cpu=%.2f%% avg_mem=%.2fMB max_mem=%.2fMB",
		s.ProcessName, s.PID, s.SampleCount,
		s.AvgCPU, s.MaxCPU,
		s.AvgMem/1024/1024, s.MaxMem/1024/1024,
	)
}

// Aggregator computes rolling statistics from a History store.
type Aggregator struct {
	history *History
}

// NewAggregator creates an Aggregator backed by the given History.
func NewAggregator(h *History) *Aggregator {
	return &Aggregator{history: h}
}

// Compute returns aggregated Stats for the named process.
// Returns an error if no samples are available.
func (a *Aggregator) Compute(name string) (Stats, error) {
	samples := a.history.Latest(name)
	if len(samples) == 0 {
		return Stats{}, fmt.Errorf("aggregator: no samples for process %q", name)
	}

	var sumCPU, maxCPU, sumMem, maxMem float64
	var pid int32

	for _, s := range samples {
		sumCPU += s.CPUPercent
		if s.CPUPercent > maxCPU {
			maxCPU = s.CPUPercent
		}
		sumMem += float64(s.MemBytes)
		if float64(s.MemBytes) > maxMem {
			maxMem = float64(s.MemBytes)
		}
		pid = s.PID
	}

	n := float64(len(samples))
	return Stats{
		ProcessName: name,
		PID:         pid,
		AvgCPU:      sumCPU / n,
		MaxCPU:      maxCPU,
		AvgMem:      sumMem / n,
		MaxMem:      maxMem,
		SampleCount: len(samples),
		Window:      a.history.window,
		ComputedAt:  time.Now(),
	}, nil
}
