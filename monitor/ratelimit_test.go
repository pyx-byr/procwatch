package monitor

import (
	"testing"
	"time"
)

func TestRateLimit_AllowsUpToBurst(t *testing.T) {
	rl := NewRateLimit(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !rl.Allow("nginx") {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}
}

func TestRateLimit_BlocksAfterBurst(t *testing.T) {
	rl := NewRateLimit(2, time.Minute)
	rl.Allow("nginx")
	rl.Allow("nginx")
	if rl.Allow("nginx") {
		t.Fatal("expected Allow to return false after burst exhausted")
	}
}

func TestRateLimit_IndependentProcesses(t *testing.T) {
	rl := NewRateLimit(1, time.Minute)
	if !rl.Allow("nginx") {
		t.Fatal("expected nginx to be allowed")
	}
	if !rl.Allow("redis") {
		t.Fatal("expected redis to be allowed independently")
	}
	if rl.Allow("nginx") {
		t.Fatal("expected nginx to be blocked after burst")
	}
}

func TestRateLimit_ResetsAfterWindow(t *testing.T) {
	rl := NewRateLimit(1, 50*time.Millisecond)
	if !rl.Allow("svc") {
		t.Fatal("first allow should succeed")
	}
	if rl.Allow("svc") {
		t.Fatal("second allow within window should be blocked")
	}
	time.Sleep(60 * time.Millisecond)
	if !rl.Allow("svc") {
		t.Fatal("allow after window expiry should succeed")
	}
}

func TestRateLimit_Remaining(t *testing.T) {
	rl := NewRateLimit(3, time.Minute)
	if got := rl.Remaining("svc"); got != 3 {
		t.Fatalf("expected 3 remaining before any calls, got %d", got)
	}
	rl.Allow("svc")
	if got := rl.Remaining("svc"); got != 2 {
		t.Fatalf("expected 2 remaining after one allow, got %d", got)
	}
}

func TestRateLimit_Reset(t *testing.T) {
	rl := NewRateLimit(1, time.Minute)
	rl.Allow("svc")
	if rl.Allow("svc") {
		t.Fatal("expected block before reset")
	}
	rl.Reset("svc")
	if !rl.Allow("svc") {
		t.Fatal("expected allow after reset")
	}
}

func TestNewRateLimit_Defaults(t *testing.T) {
	rl := NewRateLimit(0, 0)
	if rl.maxBurst != 3 {
		t.Fatalf("expected default maxBurst 3, got %d", rl.maxBurst)
	}
	if rl.window != time.Minute {
		t.Fatalf("expected default window 1m, got %v", rl.window)
	}
}
