package scene

import (
	"net/http"
	"strings"
	"sync"

	"google.golang.org/genai"
)

const maxRefImages = 5

// silhouetteGuide is the prompt snippet for >70% silhouette/back-view rendering.
const silhouetteGuide = `Character rendering guideline: Show the character predominantly from behind,
as a silhouette, or in a three-quarter back view (>70% of the time).
Use soft backlighting to create a recognizable outline.
Avoid detailed frontal face rendering — instead convey identity through posture,
hair, clothing silhouette, and body language.
This maintains visual consistency while preserving emotional mystery.`

// CharacterAnchor maintains visual consistency across scenes.
// Lock ordering: CharacterAnchor.mu is Level 4.
type CharacterAnchor struct {
	mu           sync.RWMutex
	RefImages    [][]byte // Initial person crop images (FIFO, max 5)
	LastSceneB64 string   // Last generated scene for continuity
}

// NewCharacterAnchor creates a new anchor.
func NewCharacterAnchor() *CharacterAnchor {
	return &CharacterAnchor{
		RefImages: make([][]byte, 0, maxRefImages),
	}
}

// AddRefImage adds a reference image crop (FIFO, max 5).
func (a *CharacterAnchor) AddRefImage(crop []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.RefImages = append(a.RefImages, crop)
	if len(a.RefImages) > maxRefImages {
		a.RefImages = a.RefImages[len(a.RefImages)-maxRefImages:]
	}
}

// GetRefImages returns a copy of reference images (safe for concurrent use).
func (a *CharacterAnchor) GetRefImages() [][]byte {
	a.mu.RLock()
	defer a.mu.RUnlock()

	images := make([][]byte, len(a.RefImages))
	copy(images, a.RefImages)
	return images
}

// UpdateLastScene updates the last generated scene.
func (a *CharacterAnchor) UpdateLastScene(base64Img string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.LastSceneB64 = base64Img
}

// GetLastScene returns the last scene (safe for concurrent use).
func (a *CharacterAnchor) GetLastScene() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.LastSceneB64
}

// GetRefParts converts reference images into genai.Parts for multimodal prompts.
// Returns nil if no reference images are stored.
func (a *CharacterAnchor) GetRefParts() []*genai.Part {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.RefImages) == 0 {
		return nil
	}

	parts := make([]*genai.Part, 0, len(a.RefImages))
	for _, img := range a.RefImages {
		mime := http.DetectContentType(img)
		if !strings.HasPrefix(mime, "image/") {
			mime = "image/jpeg"
		}
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: mime,
				Data:     img,
			},
		})
	}
	return parts
}

// SilhouetteGuide returns the prompt snippet for silhouette/back-view rendering.
func (a *CharacterAnchor) SilhouetteGuide() string {
	return silhouetteGuide
}
