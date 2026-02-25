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
		name      string
		args      map[string]any
		expectErr bool // tools that require a generator return an error without one
	}{
		{"generate_scene", map[string]any{"prompt": "sunset", "mood": "warm"}, false},
		{"generate_fast_scene", map[string]any{"prompt": "ocean", "mood": "calm"}, false},
		{"generate_story_page", map[string]any{"prompt": "reunion", "mood": "warm"}, true},
		{"change_atmosphere", map[string]any{"mood": "nostalgic"}, false},
		{"recall_memory", map[string]any{"query": "childhood"}, false},
		{"analyze_user", map[string]any{"aspect": "emotion"}, false},
		{"end_reunion", map[string]any{"reason": "user_request"}, false},
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
			if _, hasErr := result["error"]; hasErr && !tc.expectErr {
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

func TestToolHandler_SceneFallbackEvent(t *testing.T) {
	h := NewToolHandler()
	// No generator set — should emit a fallback event.

	var wg sync.WaitGroup
	wg.Add(1)

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

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for fallback event")
	}

	mu.Lock()
	count := len(events)
	mu.Unlock()

	if count != 1 {
		t.Fatalf("expected 1 fallback event, got %d", count)
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

func TestGetPresetBGMURL_AllMoods(t *testing.T) {
	moods := []string{"warm", "romantic", "nostalgic", "playful", "emotional", "farewell"}
	for _, mood := range moods {
		bgm := GetPresetBGMURL(mood)
		if bgm.Mood != mood {
			t.Fatalf("expected mood %q, got %q", mood, bgm.Mood)
		}
		if bgm.URL == "" {
			t.Fatalf("expected non-empty URL for mood %q", mood)
		}
	}
}

func TestGetPresetBGMURL_UnknownMood(t *testing.T) {
	bgm := GetPresetBGMURL("unknown_mood")
	if bgm.Mood != "nostalgic" {
		t.Fatalf("expected fallback to 'nostalgic', got %q", bgm.Mood)
	}
	if bgm.URL == "" {
		t.Fatal("expected non-empty fallback URL")
	}
}

func TestHandleChangeAtmosphere_WriteJSON(t *testing.T) {
	h := NewToolHandler()

	var mu sync.Mutex
	var events []map[string]any

	h.SetEventSender(func(v any) {
		mu.Lock()
		defer mu.Unlock()
		if m, ok := v.(map[string]any); ok {
			events = append(events, m)
		}
	})

	result, err := h.Handle(context.Background(), "change_atmosphere", map[string]any{"mood": "warm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check result contains bgm_url.
	if result["bgm_url"] == nil || result["bgm_url"] == "" {
		t.Fatal("expected bgm_url in result")
	}

	// Check event contains bgm_url and mood.
	mu.Lock()
	defer mu.Unlock()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev["type"] != "atmosphere_change" {
		t.Fatalf("expected type 'atmosphere_change', got %q", ev["type"])
	}
	if ev["bgm_url"] == nil || ev["bgm_url"] == "" {
		t.Fatal("expected bgm_url in event")
	}
	if ev["mood"] != "warm" {
		t.Fatalf("expected mood 'warm', got %q", ev["mood"])
	}
}

func TestAllPresetMoods(t *testing.T) {
	moods := AllPresetMoods()
	if len(moods) != 6 {
		t.Fatalf("expected 6 preset moods, got %d", len(moods))
	}

	expected := map[string]bool{
		"warm": false, "romantic": false, "nostalgic": false,
		"playful": false, "emotional": false, "farewell": false,
	}
	for _, m := range moods {
		if _, ok := expected[m]; !ok {
			t.Fatalf("unexpected mood %q", m)
		}
		expected[m] = true
	}
	for m, found := range expected {
		if !found {
			t.Fatalf("missing mood %q", m)
		}
	}
}
