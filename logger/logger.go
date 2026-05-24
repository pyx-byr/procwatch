package logger

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// Level represents log severity.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelAlert Level = "ALERT"
	LevelError Level = "ERROR"
)

// Entry is a structured log record emitted as JSON.
type Entry struct {
	Timestamp string            `json:"timestamp"`
	Level     Level             `json:"level"`
	Message   string            `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger writes structured JSON log entries to an io.Writer.
type Logger struct {
	out io.Writer
}

// New creates a Logger that writes to the given writer.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{out: w}
}

// log serialises and writes a single Entry.
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	e := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	data = append(data, '\n')
	_, _ = l.out.Write(data)
}

// Info emits an informational log entry.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(LevelInfo, msg, fields)
}

// Warn emits a warning log entry.
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(LevelWarn, msg, fields)
}

// Alert emits an alert log entry (threshold breach).
func (l *Logger) Alert(msg string, fields map[string]interface{}) {
	l.log(LevelAlert, msg, fields)
}

// Error emits an error log entry.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(LevelError, msg, fields)
}
