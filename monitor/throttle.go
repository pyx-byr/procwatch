package monitor

import (
	"sync"
	"time"
)

// Throttle prevents alert flooding by suppressing repeated alerts
// for the same process within a cooldown window.
type Throttle struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
}

// NewThrottle creates a Throttle with the given cooldown duration.
// If cooldown is zero, a default of 60 seconds is used.
func NewThrottle(cooldown time.Duration) *Throttle {
	if cooldown <= 0 {
		cooldown = 60 * time.Second
	}
	return &Throttle{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
	}
}

// Allow returns true if an alert for the given key should be emitted.
// It records the current time as the last emission time when returning true.
func (t *Throttle) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if ts, ok := t.last[key]; ok {
		if now.Sub(ts) < t.cooldown {
			return false
		}
	}
	t.last[key] = now
	return true
}

// Reset clears the throttle state for the given key, allowing the next
// alert for that key to pass through immediately.
func (t *Throttle) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.last, key)
}

// ResetAll clears all throttle state.
func (t *Throttle) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.last = make(map[string]time.Time)
}
