package monitor

import (
	"testing"
	"time"
)

func TestThrottle_AllowsFirstAlert(t *testing.T) {
	th := NewThrottle(5 * time.Second)
	if !th.Allow("nginx") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestThrottle_SuppressesWithinCooldown(t *testing.T) {
	th := NewThrottle(5 * time.Second)
	th.Allow("nginx") // first — allowed
	if th.Allow("nginx") {
		t.Fatal("expected second alert within cooldown to be suppressed")
	}
}

func TestThrottle_AllowsAfterCooldown(t *testing.T) {
	th := NewThrottle(10 * time.Millisecond)
	th.Allow("nginx")
	time.Sleep(20 * time.Millisecond)
	if !th.Allow("nginx") {
		t.Fatal("expected alert to be allowed after cooldown expires")
	}
}

func TestThrottle_IndependentKeys(t *testing.T) {
	th := NewThrottle(5 * time.Second)
	th.Allow("nginx")
	if !th.Allow("redis") {
		t.Fatal("expected independent key to be allowed")
	}
}

func TestThrottle_Reset(t *testing.T) {
	th := NewThrottle(5 * time.Second)
	th.Allow("nginx")
	th.Reset("nginx")
	if !th.Allow("nginx") {
		t.Fatal("expected alert to be allowed after explicit reset")
	}
}

func TestThrottle_ResetAll(t *testing.T) {
	th := NewThrottle(5 * time.Second)
	th.Allow("nginx")
	th.Allow("redis")
	th.ResetAll()
	if !th.Allow("nginx") || !th.Allow("redis") {
		t.Fatal("expected all keys to be allowed after ResetAll")
	}
}

func TestNewThrottle_DefaultCooldown(t *testing.T) {
	th := NewThrottle(0)
	if th.cooldown != 60*time.Second {
		t.Fatalf("expected default cooldown 60s, got %v", th.cooldown)
	}
}
