package monitor

import (
	"testing"
	"time"
)

func makeSnapshot(ts time.Time, names ...string) Snapshot {
	entries := make(map[string]SnapshotEntry, len(names))
	for i, n := range names {
		entries[n] = SnapshotEntry{
			Name:      n,
			PID:       1000 + i,
			AvgCPU:    float64(i) * 1.5,
			AvgMemory: float64(i) * 1024 * 1024,
			Samples:   3,
		}
	}
	return Snapshot{Timestamp: ts, Entries: entries}
}

func TestSnapshotStore_DefaultCapacity(t *testing.T) {
	s := NewSnapshotStore(0)
	if s.cap != 10 {
		t.Fatalf("expected default cap 10, got %d", s.cap)
	}
}

func TestSnapshotStore_EmptyLatest(t *testing.T) {
	s := NewSnapshotStore(5)
	_, ok := s.Latest()
	if ok {
		t.Fatal("expected false for empty store")
	}
}

func TestSnapshotStore_AddAndLatest(t *testing.T) {
	s := NewSnapshotStore(5)
	now := time.Now()
	snap := makeSnapshot(now, "nginx")
	s.Add(snap)

	got, ok := s.Latest()
	if !ok {
		t.Fatal("expected snapshot, got none")
	}
	if !got.Timestamp.Equal(now) {
		t.Errorf("timestamp mismatch: got %v want %v", got.Timestamp, now)
	}
	if _, exists := got.Entries["nginx"]; !exists {
		t.Error("expected nginx entry in snapshot")
	}
}

func TestSnapshotStore_RingEviction(t *testing.T) {
	s := NewSnapshotStore(3)
	base := time.Now()
	for i := 0; i < 5; i++ {
		s.Add(makeSnapshot(base.Add(time.Duration(i)*time.Second), "proc"))
	}
	if s.Len() != 3 {
		t.Fatalf("expected 3 stored snapshots, got %d", s.Len())
	}
	latest, _ := s.Latest()
	expected := base.Add(4 * time.Second)
	if !latest.Timestamp.Equal(expected) {
		t.Errorf("latest ts mismatch: got %v want %v", latest.Timestamp, expected)
	}
}

func TestSnapshotStore_AllOrder(t *testing.T) {
	s := NewSnapshotStore(4)
	base := time.Now()
	times := []time.Time{
		base, base.Add(time.Second), base.Add(2 * time.Second),
	}
	for _, ts := range times {
		s.Add(makeSnapshot(ts, "app"))
	}
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	for i, snap := range all {
		if !snap.Timestamp.Equal(times[i]) {
			t.Errorf("index %d: got %v want %v", i, snap.Timestamp, times[i])
		}
	}
}

func TestSnapshotEntry_String(t *testing.T) {
	e := SnapshotEntry{Name: "nginx", PID: 42, AvgCPU: 12.5, AvgMemory: 52428800, Samples: 5}
	got := e.String()
	if got == "" {
		t.Error("expected non-empty string from SnapshotEntry.String()")
	}
}
