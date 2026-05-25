package monitor

import (
	"sync"
	"testing"
	"time"
)

// TestAggregator_ConcurrentCompute verifies that Compute is safe under
// concurrent writes to the underlying History.
func TestAggregator_ConcurrentCompute(t *testing.T) {
	h := NewHistory(30 * time.Second)
	a := NewAggregator(h)

	const workers = 8
	const iterations = 50

	var wg sync.WaitGroup

	// Writers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				h.Add("svc", Sample{
					PID:        int32(id),
					CPUPercent: float64(j),
					MemBytes:   uint64(j * 1024),
				})
			}
		}(i)
	}

	// Readers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = a.Compute("svc")
			}
		}()
	}

	wg.Wait()

	stats, err := a.Compute("svc")
	if err != nil {
		t.Fatalf("expected stats after concurrent writes, got error: %v", err)
	}
	if stats.SampleCount == 0 {
		t.Error("expected non-zero SampleCount")
	}
}
