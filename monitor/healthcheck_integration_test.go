package monitor

import (
	"sync"
	"testing"
	"time"
)

func TestHealthChecker_ConcurrentUpdates(t *testing.T) {
	h := NewHealthChecker(5 * time.Second)
	var wg sync.WaitGroup

	processes := []string{"nginx", "redis", "postgres"}

	for _, name := range processes {
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(n string, pid int) {
				defer wg.Done()
				h.Update(n, pid)
			}(name, 1000+i)
		}
	}

	wg.Wait()

	all := h.All()
	if len(all) != len(processes) {
		t.Fatalf("expected %d statuses, got %d", len(processes), len(all))
	}
}

func TestHealthChecker_ConcurrentGetAndUpdate(t *testing.T) {
	h := NewHealthChecker(5 * time.Second)
	h.Update("nginx", 100)

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(pid int) {
			defer wg.Done()
			h.Update("nginx", pid)
		}(100 + i)
	}

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = h.Get("nginx")
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}

	s, ok := h.Get("nginx")
	if !ok {
		t.Fatal("expected nginx status to exist after concurrent access")
	}
	if s.PID <= 0 {
		t.Errorf("expected positive PID, got %d", s.PID)
	}
}
