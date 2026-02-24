package retry

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestWithBackoff_ImmediateSuccess(t *testing.T) {
	calls := 0
	err := WithBackoff(context.Background(), 3, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestWithBackoff_RetrySuccess(t *testing.T) {
	calls := 0
	err := WithBackoff(context.Background(), 3, func() error {
		calls++
		if calls < 2 {
			return errors.New("temporary error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error after retry, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestWithBackoff_MaxRetriesExceeded(t *testing.T) {
	calls := 0
	err := WithBackoff(context.Background(), 3, func() error {
		calls++
		return errors.New("persistent error")
	})
	if err == nil {
		t.Fatal("expected error after max retries")
	}
	if !strings.Contains(err.Error(), "all 3 retries failed") {
		t.Fatalf("expected 'all 3 retries failed' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "persistent error") {
		t.Fatalf("expected wrapped original error, got: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestWithBackoff_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := WithBackoff(ctx, 10, func() error {
		calls++
		return errors.New("always fail")
	})

	if err == nil {
		t.Fatal("expected error on context cancel")
	}
	// Should have been cancelled before completing all 10 retries
	if calls >= 10 {
		t.Fatalf("expected fewer than 10 calls due to cancel, got %d", calls)
	}
}

func TestCalculateDelay_Bounds(t *testing.T) {
	for attempt := 0; attempt < 20; attempt++ {
		d := calculateDelay(attempt)
		if d > 30*time.Second {
			t.Fatalf("delay for attempt %d exceeded max: %v", attempt, d)
		}
		if d < 0 {
			t.Fatalf("delay for attempt %d is negative: %v", attempt, d)
		}
	}
}

func TestCalculateDelay_JitterRange(t *testing.T) {
	// For attempt 0: base = 1s, jitter = 0~0.5s → total = 1s~1.5s
	// For attempt 1: base = 2s, jitter = 0~1s   → total = 2s~3s
	// For attempt 2: base = 4s, jitter = 0~2s   → total = 4s~6s
	cases := []struct {
		attempt int
		minD    time.Duration
		maxD    time.Duration
	}{
		{0, 1 * time.Second, 1500 * time.Millisecond},
		{1, 2 * time.Second, 3 * time.Second},
		{2, 4 * time.Second, 6 * time.Second},
	}

	for _, tc := range cases {
		for range 100 {
			d := calculateDelay(tc.attempt)
			if d < tc.minD || d > tc.maxD {
				t.Fatalf("attempt %d: delay %v out of range [%v, %v]",
					tc.attempt, d, tc.minD, tc.maxD)
			}
		}
	}
}
