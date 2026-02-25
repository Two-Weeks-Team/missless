package scene

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/util"
	"google.golang.org/genai"
)

const (
	// FlashModel is the fast preview image generation model.
	FlashModel = "gemini-2.5-flash-image"

	// ProModel is the high-quality final image generation model.
	ProModel = "imagen-4.0-generate-001"

	// flashTimeout is the max wait for the flash preview stage.
	flashTimeout = 5 * time.Second

	// proTimeout is the max wait for the pro final stage.
	proTimeout = 20 * time.Second

	// storyPageTimeout is the max wait for interleaved story page generation.
	storyPageTimeout = 10 * time.Second
)

// InterleavedResult holds both text and image from a single Gemini response.
// This demonstrates native interleaved output — text + image generated together
// in a single GenerateContent call, as required by the Creative Storyteller track.
type InterleavedResult struct {
	Text        string `json:"text"`
	ImageBase64 string `json:"imageBase64"`
}

// Generator handles 2-stage progressive image generation.
// Stage 1: Flash preview (1-3s)
// Stage 2: Imagen 4 final (8-12s)
type Generator struct {
	client *genai.Client
	anchor *CharacterAnchor
	mu     sync.RWMutex
}

// NewGenerator creates a new image generator with a genai client.
func NewGenerator(client *genai.Client) *Generator {
	return &Generator{
		client: client,
		anchor: NewCharacterAnchor(),
	}
}

// SetAnchor updates the character anchor used for consistency.
func (g *Generator) SetAnchor(anchor *CharacterAnchor) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.anchor = anchor
}

// getAnchor returns the current character anchor under lock.
func (g *Generator) getAnchor() *CharacterAnchor {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.anchor
}

// GenerateProgressive runs both stages and sends events via callback.
// Stage 1 (flash) sends a quick preview; Stage 2 (pro) follows with high quality.
func (g *Generator) GenerateProgressive(ctx context.Context, prompt, mood, characters string, onEvent func(eventType string, data string)) {
	flashCtx, flashCancel := context.WithTimeout(ctx, flashTimeout)
	proCtx, proCancel := context.WithTimeout(ctx, proTimeout)

	flashDone := make(chan struct{}, 1)

	// Stage 1: Flash preview
	util.SafeGo(func() {
		defer flashCancel()
		defer func() { flashDone <- struct{}{} }()

		img, err := g.generateFlash(flashCtx, prompt, mood, characters)
		if err != nil {
			slog.Warn("flash_generate_failed", "error", err)
			return
		}
		onEvent("scene_preview", img)
	})

	// Stage 2: Pro final (waits for flash to complete)
	util.SafeGo(func() {
		defer proCancel()

		select {
		case <-flashDone:
		case <-proCtx.Done():
			return
		}

		img, err := g.generatePro(proCtx, prompt, mood, characters)
		if err != nil {
			slog.Warn("pro_generate_failed", "error", err)
			// Flash fallback: the preview is already sent
			return
		}
		onEvent("scene_final", img)

		anchor := g.getAnchor()
		if anchor != nil {
			anchor.UpdateLastScene(img)
		}
	})
}

// GeneratePreviewOnly runs only the flash stage for quick feedback.
func (g *Generator) GeneratePreviewOnly(ctx context.Context, prompt, mood, characters string, onEvent func(eventType string, data string)) {
	flashCtx, flashCancel := context.WithTimeout(ctx, flashTimeout)

	util.SafeGo(func() {
		defer flashCancel()

		img, err := g.generateFlash(flashCtx, prompt, mood, characters)
		if err != nil {
			slog.Warn("flash_preview_only_failed", "error", err)
			return
		}
		onEvent("scene_preview", img)
	})
}

// generateFlash generates a quick preview image using the Flash model.
func (g *Generator) generateFlash(ctx context.Context, prompt, mood, characters string) (string, error) {
	fullPrompt := g.buildPrompt(prompt, mood, characters)

	slog.Info("flash_generate_start", "prompt_len", len(fullPrompt))
	start := time.Now()

	resp, err := g.client.Models.GenerateContent(ctx, FlashModel, genai.Text(fullPrompt), &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	})
	if err != nil {
		return "", fmt.Errorf("flash generate: %w", err)
	}

	slog.Info("flash_generate_done", "latency_ms", time.Since(start).Milliseconds())

	return extractImageBase64(resp)
}

// generatePro generates a high-quality final image using Imagen 4.
func (g *Generator) generatePro(ctx context.Context, prompt, mood, characters string) (string, error) {
	fullPrompt := g.buildPrompt(prompt, mood, characters)

	slog.Info("pro_generate_start", "prompt_len", len(fullPrompt))
	start := time.Now()

	numImages := int32(1)
	resp, err := g.client.Models.GenerateImages(ctx, ProModel, fullPrompt, &genai.GenerateImagesConfig{
		NumberOfImages:   numImages,
		AspectRatio:      "16:9",
		OutputMIMEType:   "image/jpeg",
		PersonGeneration: genai.PersonGenerationAllowAll,
	})
	if err != nil {
		return "", fmt.Errorf("pro generate: %w", err)
	}

	slog.Info("pro_generate_done", "latency_ms", time.Since(start).Milliseconds())

	if len(resp.GeneratedImages) == 0 {
		return "", fmt.Errorf("pro generate: no images returned")
	}

	img := resp.GeneratedImages[0]
	if img.Image == nil || len(img.Image.ImageBytes) == 0 {
		return "", fmt.Errorf("pro generate: empty image data")
	}

	return base64.StdEncoding.EncodeToString(img.Image.ImageBytes), nil
}

// buildPrompt constructs the full generation prompt from scene parts.
func (g *Generator) buildPrompt(prompt, mood, characters string) string {
	var parts []string

	parts = append(parts, prompt)

	if mood != "" {
		parts = append(parts, fmt.Sprintf("Mood and atmosphere: %s", mood))
	}

	if characters != "" {
		parts = append(parts, fmt.Sprintf("Characters: %s", characters))
	}

	// Add style guide for consistency
	parts = append(parts, "Style: warm illustration, soft lighting, emotional, nostalgic reunion scene")

	anchor := g.getAnchor()
	if anchor != nil {
		// Add silhouette/back-view guide for character consistency.
		if len(anchor.GetRefImages()) > 0 {
			parts = append(parts, anchor.SilhouetteGuide())
		}

		// Add last scene reference for continuity if available.
		lastScene := anchor.GetLastScene()
		if lastScene != "" {
			parts = append(parts, "Maintain visual continuity with previous scene")
		}
	}

	return strings.Join(parts, ". ")
}

// GenerateStoryPage generates a narrated story page with both text and illustration
// from a single GenerateContent call. This is native interleaved output — the model
// produces text narration and an accompanying image in one response.
func (g *Generator) GenerateStoryPage(ctx context.Context, prompt, mood, characters string) (*InterleavedResult, error) {
	storyCtx, cancel := context.WithTimeout(ctx, storyPageTimeout)
	defer cancel()

	storyPrompt := g.buildStoryPagePrompt(prompt, mood, characters)

	slog.Info("story_page_generate_start", "prompt_len", len(storyPrompt))
	start := time.Now()

	resp, err := g.client.Models.GenerateContent(storyCtx, FlashModel, genai.Text(storyPrompt), &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
	})
	if err != nil {
		return nil, fmt.Errorf("story page generate: %w", err)
	}

	slog.Info("story_page_generate_done", "latency_ms", time.Since(start).Milliseconds())

	result, err := extractInterleaved(resp)
	if err != nil {
		return nil, fmt.Errorf("story page extract: %w", err)
	}

	anchor := g.getAnchor()
	if anchor != nil && result.ImageBase64 != "" {
		anchor.UpdateLastScene(result.ImageBase64)
	}

	return result, nil
}

// buildStoryPagePrompt creates a prompt that asks for both narration text and an illustration.
func (g *Generator) buildStoryPagePrompt(prompt, mood, characters string) string {
	var parts []string

	parts = append(parts, "Create a story page for a virtual reunion keepsake album.")
	parts = append(parts, fmt.Sprintf(`Scene: """%s"""`, prompt))

	if mood != "" {
		parts = append(parts, fmt.Sprintf(`Mood: """%s"""`, mood))
	}
	if characters != "" {
		parts = append(parts, fmt.Sprintf(`Characters: """%s"""`, characters))
	}

	parts = append(parts, "Write a short, heartfelt narration (2-3 sentences) describing this reunion moment.")
	parts = append(parts, "Then generate a warm, nostalgic illustration that matches the narration.")
	parts = append(parts, "Style: soft watercolor, warm lighting, emotional reunion scene.")

	anchor := g.getAnchor()
	if anchor != nil {
		if len(anchor.GetRefImages()) > 0 {
			parts = append(parts, anchor.SilhouetteGuide())
		}
		if anchor.GetLastScene() != "" {
			parts = append(parts, "Maintain visual continuity with previous scenes.")
		}
	}

	return strings.Join(parts, " ")
}

// extractImageBase64 extracts the first image from a GenerateContentResponse as base64.
func extractImageBase64(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		if part.InlineData != nil && len(part.InlineData.Data) > 0 {
			return base64.StdEncoding.EncodeToString(part.InlineData.Data), nil
		}
	}

	return "", fmt.Errorf("no image data in response")
}

// extractInterleaved extracts both text and image from a single GenerateContentResponse.
// This handles native interleaved output where the model returns text + image parts together.
func extractInterleaved(resp *genai.GenerateContentResponse) (*InterleavedResult, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	content := resp.Candidates[0].Content
	if content == nil {
		return nil, fmt.Errorf("nil content in candidate")
	}

	result := &InterleavedResult{}
	var textParts []string

	for _, part := range content.Parts {
		if part.Text != "" {
			textParts = append(textParts, part.Text)
		}
		// Keep only the first image, consistent with extractImageBase64.
		if result.ImageBase64 == "" && part.InlineData != nil && len(part.InlineData.Data) > 0 {
			result.ImageBase64 = base64.StdEncoding.EncodeToString(part.InlineData.Data)
		}
	}

	result.Text = strings.Join(textParts, " ")

	if result.Text == "" && result.ImageBase64 == "" {
		return nil, fmt.Errorf("no text or image data in response")
	}

	return result, nil
}
