package scene

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Two-Weeks-Team/missless/internal/util"
)

// Generator handles 2-stage progressive image generation.
// Stage 1: gemini-2.5-flash-image (1-3s preview)
// Stage 2: imagen-4.0-generate-001 (8-12s high-quality, Imagen 4)
type Generator struct {
	anchor *CharacterAnchor
	mu     sync.RWMutex
	// TODO: T05 - Add genai.Client field
}

// GenerateProgressive runs both stages and sends events via callback.
func (g *Generator) GenerateProgressive(ctx context.Context, prompt, mood, characters string, onEvent func(eventType string, data string)) {
	flashCtx, flashCancel := context.WithTimeout(ctx, 5*secondDuration)
	defer flashCancel()

	proCtx, proCancel := context.WithTimeout(ctx, 20*secondDuration)
	defer proCancel()

	flashDone := make(chan struct{}, 1)

	// Stage 1: Flash preview
	util.SafeGo(func() {
		defer func() { flashDone <- struct{}{} }()
		img, err := g.generateFlash(flashCtx, prompt, mood, characters)
		if err != nil {
			slog.Warn("flash_generate_failed", "error", err)
			return
		}
		onEvent("scene_preview", img)
	})

	// Stage 2: Pro final (waits for flash)
	util.SafeGo(func() {
		select {
		case <-flashDone:
		case <-proCtx.Done():
			return
		}

		img, err := g.generatePro(proCtx, prompt, mood, characters)
		if err != nil {
			slog.Warn("pro_generate_failed", "error", err)
			return
		}
		onEvent("scene_final", img)
		g.anchor.UpdateLastScene(img)
	})
}

// SetAnchor updates the character anchor used for consistency.
func (g *Generator) SetAnchor(anchor *CharacterAnchor) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.anchor = anchor
}

func (g *Generator) generateFlash(ctx context.Context, prompt, mood, characters string) (string, error) {
	// TODO: T05 - Call gemini-2.5-flash-image
	return "", fmt.Errorf("not yet implemented")
}

func (g *Generator) generatePro(ctx context.Context, prompt, mood, characters string) (string, error) {
	// TODO: T05 - Call imagen-4.0-generate-001 (Imagen 4)
	return "", fmt.Errorf("not yet implemented")
}

const secondDuration = 1_000_000_000 // time.Second as int64 for multiplication
