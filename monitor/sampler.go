package monitor

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Sample holds a single point-in-time resource snapshot for a process.
type Sample struct {
	PID       int
	Name      string
	CPUPct    float64
	MemoryMB  float64
	Timestamp time.Time
}

// Sampler reads resource usage from /proc for a given PID.
type Sampler struct {
	procRoot string
}

// NewSampler creates a Sampler. procRoot is normally "/proc".
func NewSampler(procRoot string) *Sampler {
	if procRoot == "" {
		procRoot = "/proc"
	}
	return &Sampler{procRoot: procRoot}
}

// Collect reads CPU (as raw jiffies) and RSS memory for the given PID.
// CPU percentage calculation requires two samples; this returns raw ticks
// alongside memory so callers can derive a delta.
func (s *Sampler) Collect(pid int, name string) (Sample, error) {
	statPath := fmt.Sprintf("%s/%d/stat", s.procRoot, pid)
	statData, err := os.ReadFile(statPath)
	if err != nil {
		return Sample{}, fmt.Errorf("read stat for pid %d: %w", pid, err)
	}

	fields := strings.Fields(string(statData))
	if len(fields) < 24 {
		return Sample{}, fmt.Errorf("unexpected stat format for pid %d", pid)
	}

	utime, err := strconv.ParseFloat(fields[13], 64)
	if err != nil {
		return Sample{}, fmt.Errorf("parse utime: %w", err)
	}
	stime, err := strconv.ParseFloat(fields[14], 64)
	if err != nil {
		return Sample{}, fmt.Errorf("parse stime: %w", err)
	}
	rss, err := strconv.ParseInt(fields[23], 10, 64)
	if err != nil {
		return Sample{}, fmt.Errorf("parse rss: %w", err)
	}

	// Convert RSS pages to MB (assume 4 KB pages).
	memoryMB := float64(rss) * 4096.0 / (1024 * 1024)

	return Sample{
		PID:       pid,
		Name:      name,
		CPUPct:    utime + stime, // raw ticks; caller computes delta %
		MemoryMB:  memoryMB,
		Timestamp: time.Now(),
	}, nil
}
