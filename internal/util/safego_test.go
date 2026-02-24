package util

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSafeGo_NormalExecution(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	done := false
	SafeGo(func() {
		defer wg.Done()
		done = true
	})

	wg.Wait()
	if !done {
		t.Fatal("expected goroutine to execute")
	}
}

func TestSafeGo_PanicRecovery(t *testing.T) {
	done := make(chan struct{})
	SafeGo(func() {
		defer close(done)
		panic("test panic in SafeGo")
	})

	select {
	case <-done:
		// Goroutine completed (recovered from panic)
	case <-time.After(2 * time.Second):
		t.Fatal("SafeGo goroutine timed out")
	}
}

func TestSafeGoWithContext_NormalExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	done := false
	SafeGoWithContext(cancel, func() {
		defer wg.Done()
		done = true
	})

	wg.Wait()
	if !done {
		t.Fatal("expected goroutine to execute")
	}
	if ctx.Err() != nil {
		t.Fatal("context should not be cancelled on normal execution")
	}
}

func TestSafeGoWithContext_PanicCancelsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	SafeGoWithContext(cancel, func() {
		panic("test panic cancels context")
	})

	// Wait for context to be cancelled by the panic recovery
	select {
	case <-ctx.Done():
		// Context was cancelled — panic recovery called cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("expected context to be cancelled after panic, timed out")
	}
}
