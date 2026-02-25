package live

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/memory"
	"github.com/Two-Weeks-Team/missless/internal/scene"
)

// ToolHandler executes server-side tools called by the Live API.
// Lock ordering: ToolHandler.mu is Level 3.
type ToolHandler struct {
	mu              sync.RWMutex
	resumptionToken string
	// sendEvent sends a JSON event to the browser (set by Proxy).
	sendEvent func(v any)
	// generator handles image generation (nil until SetGenerator is called).
	generator *scene.Generator
	// albumGen compiles scenes into a shareable album.
	albumGen *scene.AlbumGenerator
	// memoryStore handles memory search for recall_memory tool.
	memoryStore *memory.Store
	// personaID is the current persona identifier for memory lookups.
	personaID string
}

// NewToolHandler creates a new tool handler.
func NewToolHandler() *ToolHandler {
	return &ToolHandler{}
}

// SetGenerator sets the image generator for scene tools.
func (h *ToolHandler) SetGenerator(gen *scene.Generator) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.generator = gen
}

// SetAlbumGenerator sets the album generator for end_reunion.
func (h *ToolHandler) SetAlbumGenerator(ag *scene.AlbumGenerator) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.albumGen = ag
}

// SetMemoryStore sets the memory store for recall_memory.
func (h *ToolHandler) SetMemoryStore(store *memory.Store, personaID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.memoryStore = store
	h.personaID = personaID
}

// getGenerator returns the current generator under lock.
func (h *ToolHandler) getGenerator() *scene.Generator {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.generator
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

// requireStringArg extracts a required string argument from args.
func requireStringArg(args map[string]any, key string) (string, map[string]any) {
	v, ok := args[key].(string)
	if !ok || v == "" {
		return "", map[string]any{"error": "missing or invalid '" + key + "' argument"}
	}
	return v, nil
}

func (h *ToolHandler) handleGenerateScene(ctx context.Context, args map[string]any) (map[string]any, error) {
	prompt, errResp := requireStringArg(args, "prompt")
	if errResp != nil {
		return errResp, nil
	}
	mood, errResp := requireStringArg(args, "mood")
	if errResp != nil {
		return errResp, nil
	}
	characters, _ := args["characters"].(string)

	gen := h.getGenerator()
	if gen != nil {
		// Use the real 2-stage progressive generator.
		gen.GenerateProgressive(ctx, prompt, mood, characters, func(eventType, data string) {
			h.emitEvent(map[string]any{
				"type":  eventType,
				"image": data,
			})
		})
	} else {
		// Fallback: notify browser without actual image data.
		h.emitEvent(map[string]any{"type": "scene_preview", "status": "generating"})
	}

	return map[string]any{"status": "scene generation started", "prompt": prompt, "mood": mood}, nil
}

func (h *ToolHandler) handleGenerateFastScene(ctx context.Context, args map[string]any) (map[string]any, error) {
	prompt, errResp := requireStringArg(args, "prompt")
	if errResp != nil {
		return errResp, nil
	}
	mood, errResp := requireStringArg(args, "mood")
	if errResp != nil {
		return errResp, nil
	}

	gen := h.getGenerator()
	if gen != nil {
		// Preview-only mode: fast image without the high-quality pass.
		gen.GeneratePreviewOnly(ctx, prompt, mood, "", func(eventType, data string) {
			h.emitEvent(map[string]any{
				"type":  eventType,
				"image": data,
			})
		})
	} else {
		h.emitEvent(map[string]any{"type": "scene_preview", "status": "generating"})
	}

	return map[string]any{"status": "fast scene started", "prompt": prompt, "mood": mood}, nil
}

func (h *ToolHandler) handleChangeAtmosphere(ctx context.Context, args map[string]any) (map[string]any, error) {
	mood, errResp := requireStringArg(args, "mood")
	if errResp != nil {
		return errResp, nil
	}

	bgm := GetPresetBGMURL(mood)
	h.emitEvent(map[string]any{
		"type":    "atmosphere_change",
		"mood":    bgm.Mood,
		"bgm_url": bgm.URL,
	})

	return map[string]any{"status": "atmosphere changed", "mood": bgm.Mood, "bgm_url": bgm.URL}, nil
}

func (h *ToolHandler) handleRecallMemory(ctx context.Context, args map[string]any) (map[string]any, error) {
	query, errResp := requireStringArg(args, "query")
	if errResp != nil {
		return errResp, nil
	}

	h.mu.RLock()
	store := h.memoryStore
	pid := h.personaID
	h.mu.RUnlock()

	if store == nil || pid == "" {
		return map[string]any{"memories": []any{}, "query": query}, nil
	}

	memories, err := store.Search(ctx, pid, query)
	if err != nil {
		slog.Warn("memory_search_failed", "error", err)
		return map[string]any{"memories": []any{}, "query": query}, nil
	}

	// Format memories for AI consumption.
	results := make([]map[string]string, 0, len(memories))
	for _, m := range memories {
		results = append(results, map[string]string{
			"timestamp":   m.Timestamp,
			"description": m.Description,
			"expression":  m.Expression,
		})
	}

	return map[string]any{"memories": results, "query": query, "count": len(results)}, nil
}

func (h *ToolHandler) handleAnalyzeUser(ctx context.Context, args map[string]any) (map[string]any, error) {
	aspect, errResp := requireStringArg(args, "aspect")
	if errResp != nil {
		return errResp, nil
	}

	return map[string]any{"observation": "user appears engaged", "aspect": aspect}, nil
}

func (h *ToolHandler) handleEndReunion(ctx context.Context, args map[string]any) (map[string]any, error) {
	reason, errResp := requireStringArg(args, "reason")
	if errResp != nil {
		return errResp, nil
	}

	h.emitEvent(map[string]any{
		"type":   "reunion_ending",
		"reason": reason,
	})

	// Generate album from recorded scenes.
	h.mu.RLock()
	ag := h.albumGen
	h.mu.RUnlock()

	if ag != nil {
		album, err := ag.CreateAlbum(ctx, "Reunion memories")
		if err != nil {
			slog.Warn("album_generation_failed", "error", err)
		} else {
			h.emitEvent(map[string]any{
				"type":     "album_created",
				"albumId":  album.ID,
				"shareUrl": album.ShareURL,
				"scenes":   len(album.Scenes),
			})
		}
	}

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
