package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ProcessConfig defines monitoring settings for a single process.
type ProcessConfig struct {
	Name        string  `json:"name"`
	PIDFile     string  `json:"pid_file,omitempty"`
	CPUThreshold float64 `json:"cpu_threshold_percent"`
	MemThreshold uint64  `json:"mem_threshold_mb"`
}

// Config is the top-level configuration for procwatch.
type Config struct {
	PollInterval time.Duration    `json:"-"`
	PollIntervalSec int           `json:"poll_interval_seconds"`
	LogLevel     string           `json:"log_level"`
	LogFormat    string           `json:"log_format"` // "json" or "text"
	Processes    []ProcessConfig  `json:"processes"`
}

// Load reads and parses a JSON config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.PollIntervalSec <= 0 {
		cfg.PollIntervalSec = 5
	}
	cfg.PollInterval = time.Duration(cfg.PollIntervalSec) * time.Second

	if cfg.LogFormat == "" {
		cfg.LogFormat = "json"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Processes) == 0 {
		return fmt.Errorf("at least one process must be configured")
	}
	for i, p := range c.Processes {
		if p.Name == "" {
			return fmt.Errorf("process[%d]: name is required", i)
		}
		if p.CPUThreshold < 0 || p.CPUThreshold > 100 {
			return fmt.Errorf("process %q: cpu_threshold_percent must be 0-100", p.Name)
		}
	}
	return nil
}
