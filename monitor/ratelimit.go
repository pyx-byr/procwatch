package monitor

import (
	"sync"
	"time"
)

// RateLimit tracks per-process alert rate limiting using a token bucket approach.
// It limits how many alerts can be emitted per process within a given window.
type RateLimit struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	maxBurst int
	window   time.Duration
}

type bucket struct {
	tokens    int
	windowEnd time.Time
}

// NewRateLimit creates a RateLimit that allows up to maxBurst alerts per process
// within the given time window.
func NewRateLimit(maxBurst int, window time.Duration) *RateLimit {
	if maxBurst <= 0 {
		maxBurst = 3
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimit{
		buckets:  make(map[string]*bucket),
		maxBurst: maxBurst,
		window:   window,
	}
}

// Allow returns true if an alert for the given process name is permitted.
// It consumes one token from the bucket. When the window expires the bucket resets.
func (r *RateLimit) Allow(process string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[process]
	if !ok || now.After(b.windowEnd) {
		r.buckets[process] = &bucket{
			tokens:    r.maxBurst - 1,
			windowEnd: now.Add(r.window),
		}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// Remaining returns how many alert tokens are left for a process in the current window.
func (r *RateLimit) Remaining(process string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[process]
	if !ok || now.After(b.windowEnd) {
		return r.maxBurst
	}
	return b.tokens
}

// Reset clears the bucket for a process, restoring full capacity immediately.
func (r *RateLimit) Reset(process string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, process)
}
