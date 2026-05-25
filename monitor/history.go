package monitor

import (
	"sync"
	"time"
)

// Sample holds a single resource usage snapshot for a process.
type Sample struct {
	Timestamp time.Time
	PID       int32
	Name      string
	CPUPercent float64
	MemoryMB   float64
}

// History stores a rolling window of resource samples per process name.
type History struct {
	mu      sync.RWMutex
	window  int
	samples map[string][]Sample
}

// NewHistory creates a History that retains at most windowSize samples per process.
func NewHistory(windowSize int) *History {
	if windowSize <= 0 {
		windowSize = 60
	}
	return &History{
		window:  windowSize,
		samples: make(map[string][]Sample),
	}
}

// Add appends a new sample, evicting the oldest if the window is full.
func (h *History) Add(s Sample) {
	h.mu.Lock()
	defer h.mu.Unlock()
	buf := h.samples[s.Name]
	buf = append(buf, s)
	if len(buf) > h.window {
		buf = buf[len(buf)-h.window:]
	}
	h.samples[s.Name] = buf
}

// Latest returns the most recent sample for the given process name, and whether one exists.
func (h *History) Latest(name string) (Sample, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	buf := h.samples[name]
	if len(buf) == 0 {
		return Sample{}, false
	}
	return buf[len(buf)-1], true
}

// All returns a copy of all samples for the given process name.
func (h *History) All(name string) []Sample {
	h.mu.RLock()
	defer h.mu.RUnlock()
	buf := h.samples[name]
	out := make([]Sample, len(buf))
	copy(out, buf)
	return out
}

// AverageCPU returns the mean CPU percent across all retained samples for a process.
// Returns 0 if no samples exist.
func (h *History) AverageCPU(name string) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	buf := h.samples[name]
	if len(buf) == 0 {
		return 0
	}
	var sum float64
	for _, s := range buf {
		sum += s.CPUPercent
	}
	return sum / float64(len(buf))
}

// AverageMemory returns the mean memory (MB) across all retained samples for a process.
func (h *History) AverageMemory(name string) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	buf := h.samples[name]
	if len(buf) == 0 {
		return 0
	}
	var sum float64
	for _, s := range buf {
		sum += s.MemoryMB
	}
	return sum / float64(len(buf))
}
