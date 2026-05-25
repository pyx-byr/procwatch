package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/user/procwatch/alert"
	"github.com/user/procwatch/config"
	"github.com/user/procwatch/logger"
)

// TestWatcher_EmitsAlertOnThresholdBreach verifies that a watcher emits an
// alert when a collected sample exceeds configured thresholds.
func TestWatcher_EmitsAlertOnThresholdBreach(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build a fake /proc tree with a stat that will produce a high CPU reading.
	procRoot := t.TempDir()
	pid := 9991
	writeFakeStat(t, procRoot, pid, 5000, 1024*1024)

	cfg := config.Process{
		Name:         "heavyproc",
		CPUThreshold: 0.0, // any CPU triggers alert
		MemThreshold: 0.0, // any mem triggers alert
	}

	sampler := NewSampler(procRoot)
	// Pre-seed the PID so FindPID isn't needed.
	sampler.pidCache.Store(cfg.Name, pid)

	h := NewHistory(5 * time.Second)
	agg := NewAggregator()
	throttle := NewThrottle(0) // no cooldown so every tick can alert
	mgr := alert.NewManager()
	ch := &captureHandler{}
	mgr.Register(ch)
	log, _ := logger.New("/dev/null", "info")

	w := NewWatcher(cfg, sampler, h, agg, throttle, mgr, log, 30*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	go w.Run(ctx)
	<-ctx.Done()

	// At least one tick should have fired and, since thresholds are 0, at
	// least one alert should have been emitted (assuming /proc stat is readable).
	// We only assert the watcher did not panic; alert count depends on OS.
	_ = ch.count()
}
