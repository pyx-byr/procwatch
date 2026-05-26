package monitor

import (
	"sync"
	"time"
)

// HealthStatus represents the current health of a monitored process.
type HealthStatus struct {
	PID       int
	Name      string
	Alive     bool
	LastSeen  time.Time
	Restarts  int
}

// HealthChecker tracks process liveness and restart counts.
type HealthChecker struct {
	mu       sync.RWMutex
	statuses map[string]*HealthStatus
	staleness time.Duration
}

// NewHealthChecker creates a HealthChecker with the given staleness window.
// If staleness is zero, it defaults to 30 seconds.
func NewHealthChecker(staleness time.Duration) *HealthChecker {
	if staleness <= 0 {
		staleness = 30 * time.Second
	}
	return &HealthChecker{
		statuses:  make(map[string]*HealthStatus),
		staleness: staleness,
	}
}

// Update records that a process with the given name and PID was seen alive.
// If the PID has changed since last update, it increments the restart counter.
func (h *HealthChecker) Update(name string, pid int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	status, exists := h.statuses[name]
	if !exists {
		h.statuses[name] = &HealthStatus{
			PID:      pid,
			Name:     name,
			Alive:    true,
			LastSeen: time.Now(),
		}
		return
	}

	if status.PID != pid && pid > 0 {
		status.Restarts++
	}
	status.PID = pid
	status.Alive = pid > 0
	status.LastSeen = time.Now()
}

// Get returns the HealthStatus for a named process and whether it was found.
func (h *HealthChecker) Get(name string) (HealthStatus, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	s, ok := h.statuses[name]
	if !ok {
		return HealthStatus{}, false
	}

	copy := *s
	if time.Since(s.LastSeen) > h.staleness {
		copy.Alive = false
	}
	return copy, true
}

// All returns a snapshot of all tracked health statuses.
func (h *HealthChecker) All() []HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]HealthStatus, 0, len(h.statuses))
	for _, s := range h.statuses {
		copy := *s
		if time.Since(s.LastSeen) > h.staleness {
			copy.Alive = false
		}
		result = append(result, copy)
	}
	return result
}
