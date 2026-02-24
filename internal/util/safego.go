package util

import (
	"context"
	"log/slog"
	"runtime/debug"
)

// SafeGo wraps a goroutine with panic recovery.
// Use this instead of bare `go func()` everywhere.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("GOROUTINE_PANIC",
					"error", r,
					"stack", string(debug.Stack()),
				)
			}
		}()
		fn()
	}()
}

// SafeGoWithContext wraps a goroutine with panic recovery and context cancellation.
// On panic, cancel() is called to propagate failure to parent context.
func SafeGoWithContext(cancel context.CancelFunc, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("GOROUTINE_PANIC_WITH_CANCEL",
					"error", r,
					"stack", string(debug.Stack()),
				)
				cancel()
			}
		}()
		fn()
	}()
}
