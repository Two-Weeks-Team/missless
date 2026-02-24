package scene

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// AlbumScene represents a single scene in the album.
type AlbumScene struct {
	ImageURL  string    `json:"imageUrl"`
	Caption   string    `json:"caption"`
	Timestamp time.Time `json:"timestamp"`
}

// AlbumGenerator compiles reunion scenes into a shareable album.
// Lock ordering: AlbumGenerator.mu is Level 5.
type AlbumGenerator struct {
	mu     sync.Mutex
	scenes []AlbumScene
}

// NewAlbumGenerator creates a new album generator.
func NewAlbumGenerator() *AlbumGenerator {
	return &AlbumGenerator{
		scenes: make([]AlbumScene, 0),
	}
}

// RecordScene adds a scene to the album.
// I/O (Cloud Storage upload) must happen OUTSIDE the lock.
func (ag *AlbumGenerator) RecordScene(imageURL, caption string) {
	ag.mu.Lock()
	defer ag.mu.Unlock()

	ag.scenes = append(ag.scenes, AlbumScene{
		ImageURL:  imageURL,
		Caption:   caption,
		Timestamp: time.Now(),
	})

	slog.Info("album_scene_recorded", "total", len(ag.scenes))
}

// Generate creates the final album.
func (ag *AlbumGenerator) Generate(ctx context.Context, summary string) (string, error) {
	ag.mu.Lock()
	scenes := make([]AlbumScene, len(ag.scenes))
	copy(scenes, ag.scenes)
	ag.mu.Unlock()

	if len(scenes) == 0 {
		return "", fmt.Errorf("no scenes to generate album")
	}

	// TODO: T18 - Generate album page + OG card
	// 1. Compile scenes with captions
	// 2. Generate summary card
	// 3. Upload to Cloud Storage (albums/ prefix)
	// 4. Return shareable URL

	slog.Info("album_generated", "scenes", len(scenes), "summary", summary)
	return "", fmt.Errorf("not yet implemented")
}
