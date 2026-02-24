package retry

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"
)

// WithBackoff retries fn up to maxRetries times with exponential backoff + jitter.
// Base delay is 1 second, max delay is 30 seconds.
func WithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	for attempt := range maxRetries {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == maxRetries-1 {
			break
		}

		delay := calculateDelay(attempt)
		slog.Warn("retry_backoff",
			"attempt", attempt+1,
			"max", maxRetries,
			"delay", delay,
			"error", lastErr,
		)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("all %d retries failed: %w", maxRetries, lastErr)
}

func calculateDelay(attempt int) time.Duration {
	base := math.Pow(2, float64(attempt)) * float64(time.Second)
	jitter := rand.Float64() * base * 0.5 // 0~50% of base delay
	delay := time.Duration(base) + time.Duration(jitter)

	maxDelay := 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}
