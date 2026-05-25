package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// writeFakeStat creates a minimal /proc/<pid>/stat file under a temp root.
func writeFakeStat(t *testing.T, root string, pid int, utime, stime float64, rss int64) {
	t.Helper()
	dir := filepath.Join(root, fmt.Sprintf("%d", pid))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Fields 1-24; we only care about indices 13 (utime), 14 (stime), 23 (rss).
	fields := make([]string, 24)
	for i := range fields {
		fields[i] = "0"
	}
	fields[0] = fmt.Sprintf("%d", pid)
	fields[1] = "(fake)"
	fields[13] = fmt.Sprintf("%.0f", utime)
	fields[14] = fmt.Sprintf("%.0f", stime)
	fields[23] = fmt.Sprintf("%d", rss)

	line := ""
	for i, f := range fields {
		if i > 0 {
			line += " "
		}
		line += f
	}
	if err := os.WriteFile(filepath.Join(dir, "stat"), []byte(line), 0644); err != nil {
		t.Fatalf("write stat: %v", err)
	}
}

func TestSampler_Collect_Success(t *testing.T) {
	root := t.TempDir()
	writeFakeStat(t, root, 42, 100, 50, 2048)

	s := NewSampler(root)
	sample, err := s.Collect(42, "myproc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sample.PID != 42 {
		t.Errorf("PID: got %d, want 42", sample.PID)
	}
	if sample.Name != "myproc" {
		t.Errorf("Name: got %s, want myproc", sample.Name)
	}
	// utime+stime = 150 ticks
	if sample.CPUPct != 150 {
		t.Errorf("CPUPct ticks: got %f, want 150", sample.CPUPct)
	}
	// 2048 pages * 4096 bytes / 1MB
	expectedMB := float64(2048) * 4096.0 / (1024 * 1024)
	if sample.MemoryMB != expectedMB {
		t.Errorf("MemoryMB: got %f, want %f", sample.MemoryMB, expectedMB)
	}
	if sample.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestSampler_Collect_MissingPID(t *testing.T) {
	root := t.TempDir()
	s := NewSampler(root)
	_, err := s.Collect(9999, "ghost")
	if err == nil {
		t.Fatal("expected error for missing pid, got nil")
	}
}

func TestNewSampler_DefaultRoot(t *testing.T) {
	s := NewSampler("")
	if s.procRoot != "/proc" {
		t.Errorf("procRoot: got %s, want /proc", s.procRoot)
	}
}
