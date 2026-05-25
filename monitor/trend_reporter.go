package monitor

import (
	"context"
	"time"

	"github.com/user/procwatch/logger"
)

// TrendReporter periodically computes trends from a History and logs results.
type TrendReporter struct {
	history  *History
	analyzer *TrendAnalyzer
	log      *logger.Logger
	interval time.Duration
	processes []string
	stopCh   chan struct{}
}

// NewTrendReporter creates a TrendReporter for the given processes.
func NewTrendReporter(h *History, a *TrendAnalyzer, l *logger.Logger, interval time.Duration, processes []string) *TrendReporter {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &TrendReporter{
		history:   h,
		analyzer:  a,
		log:       l,
		interval:  interval,
		processes: processes,
		stopCh:    make(chan struct{}),
	}
}

// Run starts the trend reporting loop until ctx is cancelled or Stop is called.
func (tr *TrendReporter) Run(ctx context.Context) {
	ticker := time.NewTicker(tr.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tr.report()
		case <-tr.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop halts the reporting loop.
func (tr *TrendReporter) Stop() {
	select {
	case <-tr.stopCh:
	default:
		close(tr.stopCh)
	}
}

func (tr *TrendReporter) report() {
	for _, proc := range tr.processes {
		samples := tr.history.All(proc)
		if len(samples) < 2 {
			continue
		}
		cpuVals := make([]float64, len(samples))
		memVals := make([]float64, len(samples))
		for i, s := range samples {
			cpuVals[i] = s.CPUPercent
			memVals[i] = float64(s.MemRSS)
		}
		cpuTrend := tr.analyzer.Analyze(proc, "cpu", cpuVals)
		memTrend := tr.analyzer.Analyze(proc, "mem", memVals)
		tr.log.Info("trend", map[string]interface{}{
			"process":       proc,
			"cpu_slope":     cpuTrend.Slope,
			"cpu_direction": string(cpuTrend.Direction),
			"mem_slope":     memTrend.Slope,
			"mem_direction": string(memTrend.Direction),
		})
	}
}
