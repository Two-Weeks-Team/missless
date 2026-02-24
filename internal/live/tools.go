package live

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/util"
)

// ToolHandler executes server-side tools called by the Live API.
// Lock ordering: ToolHandler.mu is Level 3.
type ToolHandler struct {
	mu              sync.RWMutex
	resumptionToken string
	// sendEvent sends a JSON event to the browser (set by Proxy).
	sendEvent func(v any)
}

// NewToolHandler creates a new tool handler.
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// SetEventSender sets the callback for sending events to the browser.
func (h *ToolHandler) SetEventSender(fn func(v any)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sendEvent = fn
}

// Handle dispatches a tool call to the appropriate handler with panic recovery and latency logging.
func (h *ToolHandler) Handle(ctx context.Context, toolName string, args map[string]any) (result map[string]any, err error) {
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			slog.Error("tool_panic_recovered", "tool", toolName, "panic", r)
			result = map[string]any{"error": fmt.Sprintf("internal error in tool %s", toolName)}
			err = nil
		}
		slog.Info("tool_executed", "tool", toolName, "latency_ms", time.Since(start).Milliseconds())
	}()

	slog.Info("tool_call", "tool", toolName)

	switch toolName {
	case "generate_scene":
		return h.handleGenerateScene(ctx, args)
	case "generate_fast_scene":
		return h.handleGenerateFastScene(ctx, args)
	case "change_atmosphere":
		return h.handleChangeAtmosphere(ctx, args)
	case "recall_memory":
		return h.handleRecallMemory(ctx, args)
	case "analyze_user":
		return h.handleAnalyzeUser(ctx, args)
	case "end_reunion":
		return h.handleEndReunion(ctx, args)
	default:
		slog.Warn("unknown_tool", "tool", toolName)
		return map[string]any{"error": "unknown tool: " + toolName}, nil
	}
}

func (h *ToolHandler) handleGenerateScene(ctx context.Context, args map[string]any) (map[string]any, error) {
	prompt, _ := args["prompt"].(string)
	mood, _ := args["mood"].(string)

	// Launch async 2-stage progressive rendering via SafeGo.
	// Stage 1 (preview) runs immediately; Stage 2 (final) follows.
	util.SafeGo(func() {
		// Stage 1: fast preview → browser
		h.emitEvent(map[string]any{
			"type":   "scene_preview",
			"prompt": prompt,
			"mood":   mood,
			"status": "generating",
		})

		// TODO: T05 - gemini-2.5-flash-image (1-3s preview)
		// TODO: T05 - imagen-4.0-generate-001 (8-12s final)

		h.emitEvent(map[string]any{
			"type":   "scene_final",
			"prompt": prompt,
			"mood":   mood,
			"status": "complete",
		})
	})

	return map[string]any{"status": "scene generation started", "prompt": prompt, "mood": mood}, nil
}

func (h *ToolHandler) handleGenerateFastScene(ctx context.Context, args map[string]any) (map[string]any, error) {
	prompt, _ := args["prompt"].(string)
	mood, _ := args["mood"].(string)

	// Fast scene uses only the preview model (no final high-res pass).
	util.SafeGo(func() {
		h.emitEvent(map[string]any{
			"type":   "scene_preview",
			"prompt": prompt,
			"mood":   mood,
			"status": "generating",
		})

		// TODO: T05 - gemini-2.5-flash-image only (1-3s)

		h.emitEvent(map[string]any{
			"type":   "scene_preview",
			"prompt": prompt,
			"mood":   mood,
			"status": "complete",
		})
	})

	return map[string]any{"status": "fast scene started", "prompt": prompt, "mood": mood}, nil
}

func (h *ToolHandler) handleChangeAtmosphere(ctx context.Context, args map[string]any) (map[string]any, error) {
	mood, _ := args["mood"].(string)

	h.emitEvent(map[string]any{
		"type": "atmosphere_change",
		"mood": mood,
	})

	// TODO: T15 - Preset BGM selection + crossfade
	return map[string]any{"status": "atmosphere changed", "mood": mood}, nil
}

func (h *ToolHandler) handleRecallMemory(ctx context.Context, args map[string]any) (map[string]any, error) {
	query, _ := args["query"].(string)

	// TODO: T16 - Firestore memory search
	return map[string]any{"memories": []string{}, "query": query}, nil
}

func (h *ToolHandler) handleAnalyzeUser(ctx context.Context, args map[string]any) (map[string]any, error) {
	aspect, _ := args["aspect"].(string)

	// TODO: T17 - Flash Vision analysis
	return map[string]any{"observation": "user appears engaged", "aspect": aspect}, nil
}

func (h *ToolHandler) handleEndReunion(ctx context.Context, args map[string]any) (map[string]any, error) {
	reason, _ := args["reason"].(string)

	h.emitEvent(map[string]any{
		"type":   "reunion_ending",
		"reason": reason,
	})

	// TODO: T18 - Album generation
	return map[string]any{"status": "reunion ending", "reason": reason}, nil
}

// emitEvent sends an event to the browser if the sender is configured.
func (h *ToolHandler) emitEvent(v any) {
	h.mu.RLock()
	fn := h.sendEvent
	h.mu.RUnlock()

	if fn != nil {
		fn(v)
	}
}

// UpdateResumptionToken stores the session resumption handle.
func (h *ToolHandler) UpdateResumptionToken(token string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.resumptionToken = token
}
