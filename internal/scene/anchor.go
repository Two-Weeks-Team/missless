package scene

import (
	"sync"
)

const maxRefImages = 5

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
