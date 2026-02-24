package live

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestToolHandler_UnknownTool(t *testing.T) {
	h := NewToolHandler()
	result, err := h.Handle(context.Background(), "nonexistent_tool", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	errMsg, ok := result["error"].(string)
	if !ok {
		t.Fatal("expected error field in result")
	}
	if errMsg != "unknown tool: nonexistent_tool" {
		t.Fatalf("expected 'unknown tool: nonexistent_tool', got %q", errMsg)
	}
}

func TestToolHandler_PanicRecovery(t *testing.T) {
	h := NewToolHandler()

	// Set an event sender that panics to test recovery.
	h.SetEventSender(func(v any) {
		panic("test panic in event sender")
	})

	// change_atmosphere calls emitEvent synchronously, which will panic.
	result, err := h.Handle(context.Background(), "change_atmosphere", map[string]any{"mood": "warm"})
	if err != nil {
		t.Fatalf("expected nil error after panic recovery, got %v", err)
	}
	errMsg, ok := result["error"].(string)
	if !ok {
		t.Fatal("expected error field in result after panic recovery")
	}
	if errMsg == "" {
		t.Fatal("expected non-empty error message after panic recovery")
	}
}

func TestToolHandler_MissingArgs(t *testing.T) {
	h := NewToolHandler()
	ctx := context.Background()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"generate_scene", nil},
		{"generate_scene", map[string]any{"prompt": "test"}},
		{"generate_fast_scene", map[string]any{}},
		{"change_atmosphere", map[string]any{}},
		{"recall_memory", map[string]any{}},
		{"analyze_user", map[string]any{}},
		{"end_reunion", map[string]any{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := h.Handle(ctx, tc.name, tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, hasErr := result["error"]; !hasErr {
				t.Fatalf("expected error for missing args, got: %v", result)
			}
		})
	}
}

func TestToolHandler_Latency(t *testing.T) {
	h := NewToolHandler()

	start := time.Now()
	_, err := h.Handle(context.Background(), "generate_scene", map[string]any{
		"prompt": "test",
		"mood":   "warm",
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Tool should execute quickly (stubs); verify latency logging doesn't block.
	if elapsed > 1*time.Second {
		t.Fatalf("tool execution too slow: %v", elapsed)
	}
}

func TestToolHandler_AllTools(t *testing.T) {
	h := NewToolHandler()
	ctx := context.Background()

	tools := []struct {
		name string
		args map[string]any
	}{
		{"generate_scene", map[string]any{"prompt": "sunset", "mood": "warm"}},
		{"generate_fast_scene", map[string]any{"prompt": "ocean", "mood": "calm"}},
		{"change_atmosphere", map[string]any{"mood": "nostalgic"}},
		{"recall_memory", map[string]any{"query": "childhood"}},
		{"analyze_user", map[string]any{"aspect": "emotion"}},
		{"end_reunion", map[string]any{"reason": "user_request"}},
	}

	for _, tc := range tools {
		t.Run(tc.name, func(t *testing.T) {
			result, err := h.Handle(ctx, tc.name, tc.args)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tc.name, err)
			}
			if result == nil {
				t.Fatalf("%s: expected non-nil result", tc.name)
			}
			if _, hasErr := result["error"]; hasErr {
				t.Fatalf("%s: got error in result: %v", tc.name, result["error"])
			}
		})
	}
}

func TestToolHandler_EventSender(t *testing.T) {
	h := NewToolHandler()

	var mu sync.Mutex
	var events []any

	h.SetEventSender(func(v any) {
		mu.Lock()
		events = append(events, v)
		mu.Unlock()
	})

	// change_atmosphere emits an event synchronously
	_, err := h.Handle(context.Background(), "change_atmosphere", map[string]any{"mood": "joyful"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	count := len(events)
	mu.Unlock()

	if count != 1 {
		t.Fatalf("expected 1 event, got %d", count)
	}
}

func TestToolHandler_AsyncSceneEvents(t *testing.T) {
	h := NewToolHandler()

	var wg sync.WaitGroup
	wg.Add(2) // Expect 2 events: scene_preview + scene_final

	var mu sync.Mutex
	var events []map[string]any

	h.SetEventSender(func(v any) {
		mu.Lock()
		if m, ok := v.(map[string]any); ok {
			events = append(events, m)
		}
		mu.Unlock()
		wg.Done()
	})

	_, err := h.Handle(context.Background(), "generate_scene", map[string]any{
		"prompt": "sunset beach",
		"mood":   "warm",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for async goroutine with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for async events")
	}

	mu.Lock()
	count := len(events)
	mu.Unlock()

	if count < 2 {
		t.Fatalf("expected at least 2 async events (preview + final), got %d", count)
	}
}

func TestToolHandler_ResumptionToken(t *testing.T) {
	h := NewToolHandler()

	h.UpdateResumptionToken("test-token-123")

	h.mu.RLock()
	token := h.resumptionToken
	h.mu.RUnlock()

	if token != "test-token-123" {
		t.Fatalf("expected token 'test-token-123', got %q", token)
	}
}
