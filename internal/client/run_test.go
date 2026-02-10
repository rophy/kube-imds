package client

import (
	"context"
	"testing"
	"time"
)

func TestRenewalDuration(t *testing.T) {
	exp := time.Now().Add(1 * time.Hour)
	d := renewalDuration(exp)

	// 80% of 1 hour = 48 minutes
	expected := 48 * time.Minute
	tolerance := 5 * time.Second
	if d < expected-tolerance || d > expected+tolerance {
		t.Errorf("expected ~%s, got %s", expected, d)
	}
}

func TestRenewalDuration_AlreadyExpired(t *testing.T) {
	exp := time.Now().Add(-1 * time.Minute)
	d := renewalDuration(exp)

	if d != 1*time.Second {
		t.Errorf("expected 1s floor, got %s", d)
	}
}

func TestRenewalDuration_ShortLifetime(t *testing.T) {
	exp := time.Now().Add(500 * time.Millisecond)
	d := renewalDuration(exp)

	if d != 1*time.Second {
		t.Errorf("expected 1s floor for short lifetime, got %s", d)
	}
}

func TestNextBackoff(t *testing.T) {
	b := initialBackoff // 1s
	b = nextBackoff(b)  // 2s
	if b != 2*time.Second {
		t.Errorf("expected 2s, got %s", b)
	}
	b = nextBackoff(b) // 4s
	if b != 4*time.Second {
		t.Errorf("expected 4s, got %s", b)
	}

	// Run until capped
	for i := 0; i < 20; i++ {
		b = nextBackoff(b)
	}
	if b != maxBackoff {
		t.Errorf("expected cap at %s, got %s", maxBackoff, b)
	}
}

func TestSleep_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	ok := sleep(ctx, 1*time.Hour)
	elapsed := time.Since(start)

	if ok {
		t.Error("expected false (context cancelled)")
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("sleep should have returned immediately, took %s", elapsed)
	}
}

func TestSleep_Completes(t *testing.T) {
	ctx := context.Background()

	start := time.Now()
	ok := sleep(ctx, 50*time.Millisecond)
	elapsed := time.Since(start)

	if !ok {
		t.Error("expected true (duration elapsed)")
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("sleep returned too early: %s", elapsed)
	}
}
