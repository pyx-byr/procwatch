package monitor

import (
	"testing"
	"time"
)

// TestHistory_ConcurrentAccess verifies that concurrent reads and writes
// do not cause data races (run with -race flag).
func TestHistory_ConcurrentAccess(t *testing.T) {
	h := NewHistory(20)
	done := make(chan struct{})

	// writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			h.Add(Sample{
				Timestamp:  time.Now(),
				PID:        42,
				Name:       "worker",
				CPUPercent: float64(i % 100),
				MemoryMB:   float64(i * 2),
			})
		}
		close(done)
	}()

	// reader goroutines
	for r := 0; r < 4; r++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					h.Latest("worker")
					h.AverageCPU("worker")
					h.AverageMemory("worker")
					_ = h.All("worker")
				}
			}
		}()
	}

	<-done
}

// TestHistory_MultipleProcesses ensures independent windows per process name.
func TestHistory_MultipleProcesses(t *testing.T) {
	h := NewHistory(5)

	for i := 0; i < 3; i++ {
		h.Add(makeSample("alpha", float64(i+1), 0))
		h.Add(makeSample("beta", float64((i+1)*10), 0))
	}

	if h.AverageCPU("alpha") == h.AverageCPU("beta") {
		t.Error("alpha and beta should have different average CPU values")
	}
	if len(h.All("alpha")) != 3 || len(h.All("beta")) != 3 {
		t.Error("each process should have exactly 3 samples")
	}
}
