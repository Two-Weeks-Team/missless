package live

import (
	"context"
	"log/slog"
	"sync"
)

// ToolHandler executes server-side tools called by the Live API.
// Lock ordering: ToolHandler.mu is Level 3.
type ToolHandler struct {
	mu              sync.RWMutex
	resumptionToken string
}

// NewToolHandler creates a new tool handler.
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// Handle dispatches a tool call to the appropriate handler.
func (h *ToolHandler) Handle(ctx context.Context, toolName string, args map[string]any) (map[string]any, error) {
	slog.Info("tool_call", "tool", toolName)

	switch toolName {
	case "generate_scene":
		return h.handleGenerateScene(ctx, args)
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
		return map[string]any{"error": "unknown tool"}, nil
	}
}

func (h *ToolHandler) handleGenerateScene(ctx context.Context, args map[string]any) (map[string]any, error) {
	// TODO: T05 - 2-stage progressive rendering
	// Stage 1: gemini-2.5-flash-image (1-3s preview) → scene_preview event
	// Stage 2: imagen-4.0-generate-001 (8-12s final) → scene_final event
	return map[string]any{"status": "scene generation started"}, nil
}

func (h *ToolHandler) handleChangeAtmosphere(ctx context.Context, args map[string]any) (map[string]any, error) {
	// TODO: T15 - Preset BGM selection + crossfade
	return map[string]any{"status": "atmosphere changed"}, nil
}

func (h *ToolHandler) handleRecallMemory(ctx context.Context, args map[string]any) (map[string]any, error) {
	// TODO: T16 - Firestore memory search
	return map[string]any{"memories": []string{}}, nil
}

func (h *ToolHandler) handleAnalyzeUser(ctx context.Context, args map[string]any) (map[string]any, error) {
	// TODO: T17 - Flash Vision analysis
	return map[string]any{"observation": "user appears engaged"}, nil
}

func (h *ToolHandler) handleEndReunion(ctx context.Context, args map[string]any) (map[string]any, error) {
	// TODO: T18 - Album generation
	return map[string]any{"status": "album generation started"}, nil
}

// UpdateResumptionToken stores the session resumption handle.
func (h *ToolHandler) UpdateResumptionToken(token string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.resumptionToken = token
}
