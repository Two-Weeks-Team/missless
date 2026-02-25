package scene

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AlbumScene represents a single scene in the album.
type AlbumScene struct {
	ImageURL  string    `json:"imageUrl"`
	Caption   string    `json:"caption"`
	Narration string    `json:"narration,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Album is the final compiled album data.
type Album struct {
	ID        string       `json:"id"`
	Scenes    []AlbumScene `json:"scenes"`
	Summary   string       `json:"summary"`
	Persona   string       `json:"persona"`
	CreatedAt time.Time    `json:"createdAt"`
	ShareURL  string       `json:"shareUrl"`
}

// UploadFunc uploads an image and returns its public URL.
// For hackathon, this is a no-op that returns the input URL.
type UploadFunc func(ctx context.Context, imageData, filename string) (string, error)

// AlbumGenerator compiles reunion scenes into a shareable album.
// Lock ordering: AlbumGenerator.mu is Level 5.
type AlbumGenerator struct {
	mu       sync.Mutex
	scenes   []AlbumScene
	persona  string
	uploadFn UploadFunc
}

// NewAlbumGenerator creates a new album generator.
func NewAlbumGenerator() *AlbumGenerator {
	return &AlbumGenerator{
		scenes: make([]AlbumScene, 0),
	}
}

// SetPersona sets the persona name for the album.
func (ag *AlbumGenerator) SetPersona(name string) {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	ag.persona = name
}

// SetUploadFunc sets the upload function for Cloud Storage.
func (ag *AlbumGenerator) SetUploadFunc(fn UploadFunc) {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	ag.uploadFn = fn
}

// RecordScene adds a scene to the album.
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

// RecordStoryPage adds a scene with narration from interleaved generation.
func (ag *AlbumGenerator) RecordStoryPage(imageURL, caption, narration string) {
	ag.mu.Lock()
	defer ag.mu.Unlock()

	ag.scenes = append(ag.scenes, AlbumScene{
		ImageURL:  imageURL,
		Caption:   caption,
		Narration: narration,
		Timestamp: time.Now(),
	})

	slog.Info("album_story_page_recorded", "total", len(ag.scenes))
}

// SceneCount returns the number of recorded scenes.
func (ag *AlbumGenerator) SceneCount() int {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	return len(ag.scenes)
}

// CreateAlbum compiles recorded scenes into a shareable album.
// Upload failures for individual scenes are skipped (best-effort).
func (ag *AlbumGenerator) CreateAlbum(ctx context.Context, summary string) (*Album, error) {
	ag.mu.Lock()
	scenes := make([]AlbumScene, len(ag.scenes))
	copy(scenes, ag.scenes)
	persona := ag.persona
	uploadFn := ag.uploadFn
	ag.mu.Unlock()

	if len(scenes) == 0 {
		// Empty album is valid — just no scenes.
		return &Album{
			ID:        generateAlbumID(),
			Scenes:    scenes,
			Summary:   summary,
			Persona:   persona,
			CreatedAt: time.Now(),
		}, nil
	}

	albumID := generateAlbumID()

	// Upload scenes (best-effort: skip failures).
	// Use albumID in path to prevent cross-album overwrites.
	if uploadFn != nil {
		safeName := sanitizePathComponent(persona)
		for i, s := range scenes {
			filename := fmt.Sprintf("albums/%s/%s/scene_%d.jpg", safeName, albumID, i)
			url, err := uploadFn(ctx, s.ImageURL, filename)
			if err != nil {
				slog.Warn("album_upload_skipped", "scene", i, "error", err)
				continue
			}
			scenes[i].ImageURL = url
		}
	}
	album := &Album{
		ID:        albumID,
		Scenes:    scenes,
		Summary:   summary,
		Persona:   persona,
		CreatedAt: time.Now(),
		ShareURL:  fmt.Sprintf("/album/%s", albumID),
	}

	slog.Info("album_created", "id", albumID, "scenes", len(scenes), "persona", persona)
	return album, nil
}

// sanitizePathComponent strips directory traversal from a user-supplied name,
// keeping only the base filename and replacing path separators.
func sanitizePathComponent(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	if name == "" || name == "." {
		name = "unknown"
	}
	return name
}

// generateAlbumID creates a random 8-byte hex album ID.
func generateAlbumID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("album-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
