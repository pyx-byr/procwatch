package monitor

import (
	"testing"
	"time"
)

func TestHealthChecker_DefaultStaleness(t *testing.T) {
	h := NewHealthChecker(0)
	if h.staleness != 30*time.Second {
		t.Fatalf("expected 30s staleness, got %v", h.staleness)
	}
}

func TestHealthChecker_UpdateAndGet_Alive(t *testing.T) {
	h := NewHealthChecker(5 * time.Second)
	h.Update("nginx", 1234)

	s, ok := h.Get("nginx")
	if !ok {
		t.Fatal("expected status to exist")
	}
	if !s.Alive {
		t.Error("expected process to be alive")
	}
	if s.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", s.PID)
	}
	if s.Restarts != 0 {
		t.Errorf("expected 0 restarts, got %d", s.Restarts)
	}
}

func TestHealthChecker_Get_Missing(t *testing.T) {
	h := NewHealthChecker(5 * time.Second)
	_, ok := h.Get("unknown")
	if ok {
		t.Fatal("expected missing process to return false")
	}
}

func TestHealthChecker_DetectsRestart(t *testing.T) {
	h := NewHealthChecker(5 * time.Second)
	h.Update("nginx", 1000)
	h.Update("nginx", 2000)

	s, _ := h.Get("nginx")
	if s.Restarts != 1 {
		t.Errorf("expected 1 restart, got %d", s.Restarts)
	}
	if s.PID != 2000 {
		t.Errorf("expected PID 2000, got %d", s.PID)
	}
}

func TestHealthChecker_StalenessMarksNotAlive(t *testing.T) {
	h := NewHealthChecker(1 * time.Millisecond)
	h.Update("nginx", 1234)

	time.Sleep(5 * time.Millisecond)

	s, ok := h.Get("nginx")
	if !ok {
		t.Fatal("expected status to exist")
	}
	if s.Alive {
		t.Error("expected stale process to be marked not alive")
	}
}

func TestHealthChecker_All(t *testing.T) {
	h := NewHealthChecker(5 * time.Second)
	h.Update("nginx", 1)
	h.Update("redis", 2)
	h.Update("postgres", 3)

	all := h.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 statuses, got %d", len(all))
	}
}

func TestHealthChecker_SamePIDNoRestart(t *testing.T) {
	h := NewHealthChecker(5 * time.Second)
	h.Update("nginx", 1234)
	h.Update("nginx", 1234)
	h.Update("nginx", 1234)

	s, _ := h.Get("nginx")
	if s.Restarts != 0 {
		t.Errorf("expected 0 restarts for same PID, got %d", s.Restarts)
	}
}
