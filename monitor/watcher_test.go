package monitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/user/procwatch/alert"
	"github.com/user/procwatch/config"
	"github.com/user/procwatch/logger"
)

type captureHandler struct {
	mu     sync.Mutex
	events []alert.Event
}

func (c *captureHandler) Handle(ev alert.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, ev)
	return nil
}

func (c *captureHandler) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.events)
}

func newTestWatcher(t *testing.T, cfg config.Process, sampler *Sampler) (*Watcher, *captureHandler) {
	t.Helper()
	h := NewHistory(5 * time.Second)
	agg := NewAggregator()
	throttle := NewThrottle(10 * time.Second)
	mgr := alert.NewManager()
	ch := &captureHandler{}
	mgr.Register(ch)
	log, _ := logger.New("/dev/null", "info")
	return NewWatcher(cfg, sampler, h, agg, throttle, mgr, log, 20*time.Millisecond), ch
}

func TestWatcher_Stop_Idempotent(t *testing.T) {
	cfg := config.Process{Name: "idle", CPUThreshold: 90, MemThreshold: 90}
	sampler := NewSampler("/proc")
	w, _ := newTestWatcher(t, cfg, sampler)
	w.Stop()
	w.Stop() // must not panic
}

func TestWatcher_RunAndStop(t *testing.T) {
	cfg := config.Process{Name: "idle", CPUThreshold: 90, MemThreshold: 90}
	sampler := NewSampler("/proc")
	w, _ := newTestWatcher(t, cfg, sampler)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("watcher did not stop after context cancel")
	}
}

func TestWatcher_StopBeforeRun(t *testing.T) {
	cfg := config.Process{Name: "idle", CPUThreshold: 90, MemThreshold: 90}
	sampler := NewSampler("/proc")
	w, _ := newTestWatcher(t, cfg, sampler)
	w.Stop()
	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("watcher should exit immediately when already stopped")
	}
}
