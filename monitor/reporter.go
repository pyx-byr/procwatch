package monitor

import (
	"time"

	"github.com/user/procwatch/logger"
)

// Reporter periodically aggregates stats for configured processes
// and emits them via the structured logger.
type Reporter struct {
	agg      *Aggregator
	log      *logger.Logger
	names    []string
	interval time.Duration
	stop     chan struct{}
}

// NewReporter creates a Reporter that logs aggregated stats for the
// given process names at the specified interval.
func NewReporter(agg *Aggregator, log *logger.Logger, names []string, interval time.Duration) *Reporter {
	return &Reporter{
		agg:      agg,
		log:      log,
		names:    names,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start begins the reporting loop in a background goroutine.
func (r *Reporter) Start() {
	go r.run()
}

// Stop signals the reporting loop to exit.
func (r *Reporter) Stop() {
	close(r.stop)
}

func (r *Reporter) run() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.report()
		case <-r.stop:
			return
		}
	}
}

func (r *Reporter) report() {
	for _, name := range r.names {
		stats, err := r.agg.Compute(name)
		if err != nil {
			r.log.Warn("aggregation skipped", map[string]interface{}{
				"process": name,
				"error":   err.Error(),
			})
			continue
		}
		r.log.Info("process stats", map[string]interface{}{
			"process":      stats.ProcessName,
			"pid":          stats.PID,
			"avg_cpu":      stats.AvgCPU,
			"max_cpu":      stats.MaxCPU,
			"avg_mem_mb":   stats.AvgMem / 1024 / 1024,
			"max_mem_mb":   stats.MaxMem / 1024 / 1024,
			"sample_count": stats.SampleCount,
			"window_sec":   stats.Window.Seconds(),
		})
	}
}
