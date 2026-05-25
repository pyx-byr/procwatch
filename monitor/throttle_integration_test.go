package monitor

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestThrottle_ConcurrentAccess verifies that the throttle is safe
// to use from multiple goroutines simultaneously.
func TestThrottle_ConcurrentAccess(t *testing.T) {
	th := NewThrottle(50 * time.Millisecond)
	const goroutines = 20
	var allowed int64

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if th.Allow("proc") {
				atomic.AddInt64(&allowed, 1)
			}
		}()
	}
	wg.Wait()

	// Only one goroutine should have been allowed through.
	if allowed != 1 {
		t.Fatalf("expected exactly 1 allowed, got %d", allowed)
	}
}

// TestThrottle_MultipleProcessesConcurrent verifies independent throttling
// across many distinct keys under concurrent load.
func TestThrottle_MultipleProcessesConcurrent(t *testing.T) {
	th := NewThrottle(5 * time.Second)
	keys := []string{"nginx", "redis", "postgres", "memcached"}
	var allowed int64

	var wg sync.WaitGroup
	for _, k := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			if th.Allow(key) {
				atomic.AddInt64(&allowed, 1)
			}
		}(k)
	}
	wg.Wait()

	if int(allowed) != len(keys) {
		t.Fatalf("expected %d allowed (one per key), got %d", len(keys), allowed)
	}
}
