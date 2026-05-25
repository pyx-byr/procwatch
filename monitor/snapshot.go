package monitor

import (
	"fmt"
	"sync"
	"time"
)

// Snapshot holds a point-in-time view of all monitored processes.
type Snapshot struct {
	Timestamp time.Time
	Entries   map[string]SnapshotEntry
}

// SnapshotEntry holds aggregated stats for a single process at snapshot time.
type SnapshotEntry struct {
	Name      string
	PID       int
	AvgCPU    float64
	AvgMemory float64
	Samples   int
}

func (e SnapshotEntry) String() string {
	return fmt.Sprintf("%s(pid=%d) cpu=%.2f%% mem=%.2fMB samples=%d",
		e.Name, e.PID, e.AvgCPU, e.AvgMemory/1024/1024, e.Samples)
}

// SnapshotStore keeps a bounded ring of recent snapshots.
type SnapshotStore struct {
	mu       sync.RWMutex
	buf      []Snapshot
	cap      int
	head     int
	count    int
}

// NewSnapshotStore creates a store that retains up to capacity snapshots.
func NewSnapshotStore(capacity int) *SnapshotStore {
	if capacity <= 0 {
		capacity = 10
	}
	return &SnapshotStore{
		buf: make([]Snapshot, capacity),
		cap: capacity,
	}
}

// Add inserts a new snapshot, evicting the oldest if at capacity.
func (s *SnapshotStore) Add(snap Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf[s.head] = snap
	s.head = (s.head + 1) % s.cap
	if s.count < s.cap {
		s.count++
	}
}

// Latest returns the most recently added snapshot and true, or false if empty.
func (s *SnapshotStore) Latest() (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.count == 0 {
		return Snapshot{}, false
	}
	idx := (s.head - 1 + s.cap) % s.cap
	return s.buf[idx], true
}

// All returns all stored snapshots in chronological order (oldest first).
func (s *SnapshotStore) All() []Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Snapshot, s.count)
	start := (s.head - s.count + s.cap) % s.cap
	for i := 0; i < s.count; i++ {
		result[i] = s.buf[(start+i)%s.cap]
	}
	return result
}

// Len returns the number of snapshots currently stored.
func (s *SnapshotStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}
