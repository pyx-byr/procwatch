package monitor

import (
	"time"

	"github.com/user/procwatch/logger"
)

// BaselineReporter periodically logs baseline stats for all tracked processes.
type BaselineReporter struct {
	tracker  *BaselineTracker
	log      *logger.Logger
	processes []string
	interval time.Duration
	stop     chan struct{}
	done     chan struct{}
}

// NewBaselineReporter creates a BaselineReporter that logs stats at the given interval.
// If interval is zero, it defaults to 5 minutes.
func NewBaselineReporter(tracker *BaselineTracker, log *logger.Logger, processes []string, interval time.Duration) *BaselineReporter {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	return &BaselineReporter{
		tracker:   tracker,
		log:       log,
		processes: processes,
		interval:  interval,
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}
}

// Run starts the reporter loop. Call Stop to terminate it.
func (br *BaselineReporter) Run() {
	defer close(br.done)
	ticker := time.NewTicker(br.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			br.emit()
		case <-br.stop:
			return
		}
	}
}

// Stop signals the reporter to stop and waits for it to finish.
func (br *BaselineReporter) Stop() {
	select {
	case <-br.stop:
	default:
		close(br.stop)
	}
	<-br.done
}

func (br *BaselineReporter) emit() {
	for _, name := range br.processes {
		stats, ok := br.tracker.Compute(name)
		if !ok {
			br.log.Warn("baseline: no samples yet", map[string]interface{}{
				"process": name,
			})
			continue
		}
		br.log.Info("baseline_stats", map[string]interface{}{
			"process":    name,
			"avg_cpu":    stats.AvgCPU,
			"avg_mem_mb": stats.AvgMemory / 1024 / 1024,
			"samples":    stats.Samples,
		})
	}
}
