package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "procwatch-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, `{
		"poll_interval_seconds": 10,
		"log_level": "debug",
		"log_format": "json",
		"processes": [
			{"name": "nginx", "cpu_threshold_percent": 80, "mem_threshold_mb": 512}
		]
	}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 10*time.Second {
		t.Errorf("expected 10s poll interval, got %v", cfg.PollInterval)
	}
	if len(cfg.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(cfg.Processes))
	}
	if cfg.Processes[0].Name != "nginx" {
		t.Errorf("expected process name 'nginx', got %q", cfg.Processes[0].Name)
	}
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTempConfig(t, `{"processes": [{"name": "app", "cpu_threshold_percent": 50}]}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Errorf("expected default 5s poll interval, got %v", cfg.PollInterval)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected default log format 'json', got %q", cfg.LogFormat)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got %q", cfg.LogLevel)
	}
}

func TestLoad_NoProcesses(t *testing.T) {
	path := writeTempConfig(t, `{"poll_interval_seconds": 5, "processes": []}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty processes, got nil")
	}
}

func TestLoad_InvalidCPUThreshold(t *testing.T) {
	path := writeTempConfig(t, `{"processes": [{"name": "bad", "cpu_threshold_percent": 150}]}`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid cpu threshold, got nil")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
