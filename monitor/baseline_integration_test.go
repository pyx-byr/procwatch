package monitor

import (
	"sync"
	"testing"
	"time"
)

func TestBaselineTracker_ConcurrentAdd(t *testing.T) {
	bt := NewBaselineTracker(1 * time.Minute)
	var wg sync.WaitGroup
	workers := 20
	iterations := 50
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				bt.Add("shared", float64(id)+float64(j)*0.1, float64(id*1024))
			}
		}(i)
	}
	wg.Wait()
	stats, ok := bt.Compute("shared")
	if !ok {
		t.Fatal("expected stats after concurrent adds")
	}
	expected := workers * iterations
	if stats.Samples != expected {
		t.Errorf("expected %d samples, got %d", expected, stats.Samples)
	}
}

func TestBaselineTracker_MultipleProcessesConcurrent(t *testing.T) {
	bt := NewBaselineTracker(1 * time.Minute)
	processes := []string{"alpha", "beta", "gamma", "delta"}
	var wg sync.WaitGroup
	for _, name := range processes {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				bt.Add(n, float64(i), float64(i*512))
			}
		}(name)
	}
	wg.Wait()
	for _, name := range processes {
		stats, ok := bt.Compute(name)
		if !ok {
			t.Errorf("expected stats for process %s", name)
			continue
		}
		if stats.Samples != 30 {
			t.Errorf("process %s: expected 30 samples, got %d", name, stats.Samples)
		}
	}
}
