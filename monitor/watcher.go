package monitor

import (
	"context"
	"time"

	"github.com/user/procwatch/alert"
	"github.com/user/procwatch/config"
	"github.com/user/procwatch/logger"
)

// Watcher ties together the Sampler, History, Aggregator, and alert pipeline
// for a single watched process. It runs a periodic collection loop.
type Watcher struct {
	cfg      config.Process
	sampler  *Sampler
	history  *History
	agg      *Aggregator
	throttle *Throttle
	alerts   *alert.Manager
	log      *logger.Logger
	interval time.Duration
	stop     chan struct{}
}

// NewWatcher creates a Watcher for the given process config.
func NewWatcher(
	cfg config.Process,
	sampler *Sampler,
	history *History,
	agg *Aggregator,
	throttle *Throttle,
	alerts *alert.Manager,
	log *logger.Logger,
	interval time.Duration,
) *Watcher {
	return &Watcher{
		cfg:      cfg,
		sampler:  sampler,
		history:  history,
		agg:      agg,
		throttle: throttle,
		alerts:   alerts,
		log:      log,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Run starts the watch loop, blocking until ctx is cancelled or Stop is called.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.tick()
		case <-w.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals the watcher loop to exit.
func (w *Watcher) Stop() {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
}

func (w *Watcher) tick() {
	sample, err := w.sampler.Collect(w.cfg.Name)
	if err != nil {
		w.log.Warn("collect_error", map[string]interface{}{
			"process": w.cfg.Name,
			"error":   err.Error(),
		})
		return
	}
	w.history.Add(w.cfg.Name, sample)
	w.agg.Add(sample)

	events := CheckAlerts(w.cfg, sample)
	for _, ev := range events {
		if w.throttle.Allow(ev.Process + ":" + ev.Kind) {
			w.alerts.Emit(ev)
		}
	}
}
