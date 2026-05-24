package alert

import (
	"fmt"
	"time"
)

// Severity represents the level of an alert.
type Severity string

const (
	SeverityWarn     Severity = "warn"
	SeverityCritical Severity = "critical"
)

// Event holds the details of a single threshold violation.
type Event struct {
	Timestamp   time.Time `json:"timestamp"`
	Process     string    `json:"process"`
	PID         int32     `json:"pid"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Severity    Severity  `json:"severity"`
	Message     string    `json:"message"`
}

// Handler is a function that receives alert events.
type Handler func(e Event)

// Manager dispatches alert events to registered handlers.
type Manager struct {
	handlers []Handler
}

// NewManager creates an empty Manager.
func NewManager() *Manager {
	return &Manager{}
}

// Register adds a handler to the manager.
func (m *Manager) Register(h Handler) {
	m.handlers = append(m.handlers, h)
}

// Emit builds an Event and forwards it to all registered handlers.
func (m *Manager) Emit(process string, pid int32, metric string, value, threshold float64, sev Severity) {
	e := Event{
		Timestamp: time.Now().UTC(),
		Process:   process,
		PID:       pid,
		Metric:    metric,
		Value:     value,
		Threshold: threshold,
		Severity:  sev,
		Message: fmt.Sprintf("%s: %s %.2f exceeds threshold %.2f",
			sev, metric, value, threshold),
	}
	for _, h := range m.handlers {
		h(e)
	}
}
