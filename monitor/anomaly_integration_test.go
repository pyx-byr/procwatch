package monitor

import (
	"sync"
	"testing"
)

// TestAnomalyDetector_ConcurrentAdd verifies thread-safety when multiple
// goroutines add samples and analyze concurrently.
func TestAnomalyDetector_ConcurrentAdd(t *testing.T) {
	d := NewAnomalyDetector(2.0, 50)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			d.Add("svc", float64(n)*1.5, float64(n)*10)
		}(i)
	}
	wg.Wait()

	// After concurrent adds, Analyze should not panic.
	result := d.Analyze("svc", 50.0, 300.0)
	// result may be nil if baseline has no stddev data yet; that's acceptable.
	_ = result
}

// TestAnomalyDetector_MultipleProcesses ensures independent anomaly detection
// across different process names.
func TestAnomalyDetector_MultipleProcesses(t *testing.T) {
	d := NewAnomalyDetector(2.0, 20)
	processes := []string{"nginx", "redis", "postgres"}

	for _, p := range processes {
		for i := 0; i < 10; i++ {
			d.Add(p, 20.0, 150.0)
		}
	}

	for _, p := range processes {
		result := d.Analyze(p, 20.5, 151.0)
		if result == nil {
			t.Errorf("expected result for process %s", p)
			continue
		}
		if result.CPUAnomaly || result.MemAnomaly {
			t.Errorf("unexpected anomaly for process %s: %s", p, result)
		}
	}

	// Inject a spike only for nginx.
	result := d.Analyze("nginx", 999.0, 150.0)
	if result == nil {
		t.Fatal("expected result for nginx spike")
	}
	if !result.CPUAnomaly {
		t.Errorf("expected CPU anomaly for nginx spike, z=%.2f", result.CPUZScore)
	}

	// Other processes should be unaffected.
	for _, p := range []string{"redis", "postgres"} {
		r := d.Analyze(p, 20.0, 150.0)
		if r != nil && (r.CPUAnomaly || r.MemAnomaly) {
			t.Errorf("unexpected anomaly for %s after nginx spike: %s", p, r)
		}
	}
}
