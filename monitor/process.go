package monitor

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/procwatch/config"
)

// ProcessStats holds the current resource usage of a monitored process.
type ProcessStats struct {
	PID        int
	Name       string
	CPUPercent float64
	MemoryMB   float64
}

// Alert represents a threshold violation for a process.
type Alert struct {
	Process string
	PID     int
	Kind    string // "cpu" or "memory"
	Value   float64
	Limit   float64
}

// Collector gathers stats for configured processes.
type Collector struct {
	cfg *config.Config
}

// NewCollector creates a new Collector from the given config.
func NewCollector(cfg *config.Config) *Collector {
	return &Collector{cfg: cfg}
}

// FindPID attempts to find a running PID for the given process name
// by scanning /proc on Linux.
func FindPID(name string) (int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, fmt.Errorf("read /proc: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		commData, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(commData)) == name {
			return pid, nil
		}
	}
	return 0, fmt.Errorf("process %q not found", name)
}

// CheckAlerts compares the given stats against configured thresholds
// and returns any alerts that were triggered.
func CheckAlerts(stats ProcessStats, proc config.Process) []Alert {
	var alerts []Alert
	if proc.CPUThreshold > 0 && stats.CPUPercent > proc.CPUThreshold {
		alerts = append(alerts, Alert{
			Process: proc.Name,
			PID:     stats.PID,
			Kind:    "cpu",
			Value:   stats.CPUPercent,
			Limit:   proc.CPUThreshold,
		})
	}
	if proc.MemThreshold > 0 && stats.MemoryMB > proc.MemThreshold {
		alerts = append(alerts, Alert{
			Process: proc.Name,
			PID:     stats.PID,
			Kind:    "memory",
			Value:   stats.MemoryMB,
			Limit:   proc.MemThreshold,
		})
	}
	return alerts
}
