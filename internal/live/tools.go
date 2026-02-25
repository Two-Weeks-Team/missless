package live

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/memory"
	"github.com/Two-Weeks-Team/missless/internal/scene"
	"google.golang.org/genai"
)

// maxTranscriptBuffer is the maximum number of recent transcript entries to keep.
const maxTranscriptBuffer = 30

// transcriptEntry stores a single conversation turn.
type transcriptEntry struct {
	Role string
	Text string
}

// analysisModel is the Gemini model used for user analysis.
const analysisModel = "gemini-2.5-flash"

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
	// genaiClient is used for Flash-based user analysis.
	genaiClient *genai.Client
	// transcripts stores recent conversation turns for analysis context.
	transcripts []transcriptEntry
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

// SetGenaiClient sets the genai client for Flash-based analysis.
func (h *ToolHandler) SetGenaiClient(client *genai.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.genaiClient = client
}

// AddTranscript appends a conversation turn to the buffer (capped at maxTranscriptBuffer).
func (h *ToolHandler) AddTranscript(role, text string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.transcripts = append(h.transcripts, transcriptEntry{Role: role, Text: text})
	if len(h.transcripts) > maxTranscriptBuffer {
		h.transcripts = h.transcripts[len(h.transcripts)-maxTranscriptBuffer:]
	}
}

// getTranscriptContext returns the recent transcript as a formatted string.
func (h *ToolHandler) getTranscriptContext() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.transcripts) == 0 {
		return "(no conversation yet)"
	}
	var sb strings.Builder
	for _, t := range h.transcripts {
		sb.WriteString(t.Role)
		sb.WriteString(": ")
		sb.WriteString(t.Text)
		sb.WriteString("\n")
	}
	return sb.String()
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

	h.mu.RLock()
	client := h.genaiClient
	h.mu.RUnlock()

	if client == nil {
		slog.Warn("analyze_user_no_client", "aspect", aspect)
		return map[string]any{"observation": "unable to analyze — client not configured", "aspect": aspect}, nil
	}

	transcript := h.getTranscriptContext()
	prompt := fmt.Sprintf(`Analyze the following conversation and provide a brief observation about the user's %s.

Recent conversation:
%s

Respond with ONLY a JSON object (no markdown, no code fences):
{"observation": "<one sentence observation about user's %s>", "confidence": "<low|medium|high>", "suggestion": "<one sentence suggestion for how to respond>"}`, aspect, transcript, aspect)

	analysisCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := client.Models.GenerateContent(analysisCtx, analysisModel, []*genai.Content{
		genai.NewContentFromText(prompt, "user"),
	}, nil)
	if err != nil {
		slog.Warn("analyze_user_flash_error", "error", err, "aspect", aspect)
		return map[string]any{"observation": "analysis temporarily unavailable", "aspect": aspect}, nil
	}

	// Extract text from response.
	if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				slog.Info("analyze_user_result", "aspect", aspect, "result", part.Text)
				return map[string]any{"analysis": part.Text, "aspect": aspect}, nil
			}
		}
	}

	return map[string]any{"observation": "no analysis available", "aspect": aspect}, nil
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
