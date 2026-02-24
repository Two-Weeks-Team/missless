package scene

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestAlbumGenerator_RecordScene(t *testing.T) {
	ag := NewAlbumGenerator()

	if ag.SceneCount() != 0 {
		t.Fatalf("expected 0 scenes initially, got %d", ag.SceneCount())
	}

	ag.RecordScene("img1-base64", "First scene")
	ag.RecordScene("img2-base64", "Second scene")
	ag.RecordScene("img3-base64", "Third scene")

	if ag.SceneCount() != 3 {
		t.Fatalf("expected 3 scenes, got %d", ag.SceneCount())
	}
}

func TestAlbumGenerator_CreateAlbum_Empty(t *testing.T) {
	ag := NewAlbumGenerator()
	ag.SetPersona("Mom")

	album, err := ag.CreateAlbum(context.Background(), "Test summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if album == nil {
		t.Fatal("expected non-nil album")
	}
	if len(album.Scenes) != 0 {
		t.Fatalf("expected 0 scenes in empty album, got %d", len(album.Scenes))
	}
	if album.Summary != "Test summary" {
		t.Fatalf("expected summary 'Test summary', got %q", album.Summary)
	}
	if album.Persona != "Mom" {
		t.Fatalf("expected persona 'Mom', got %q", album.Persona)
	}
	if album.ID == "" {
		t.Fatal("expected non-empty album ID")
	}
}

func TestAlbumGenerator_CreateAlbum_WithScenes(t *testing.T) {
	ag := NewAlbumGenerator()
	ag.SetPersona("Dad")

	ag.RecordScene("scene1", "At the park")
	ag.RecordScene("scene2", "At the cafe")

	album, err := ag.CreateAlbum(context.Background(), "Reunion with Dad")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(album.Scenes) != 2 {
		t.Fatalf("expected 2 scenes, got %d", len(album.Scenes))
	}
	if album.ShareURL == "" {
		t.Fatal("expected non-empty share URL")
	}
	if album.Persona != "Dad" {
		t.Fatalf("expected persona 'Dad', got %q", album.Persona)
	}
}

func TestAlbumGenerator_CreateAlbum_UploadFail(t *testing.T) {
	ag := NewAlbumGenerator()
	ag.SetPersona("Mom")

	// Upload function that fails for the second scene.
	callCount := 0
	ag.SetUploadFunc(func(ctx context.Context, imageData, filename string) (string, error) {
		callCount++
		if callCount == 2 {
			return "", fmt.Errorf("upload failed")
		}
		return "https://storage.example.com/" + filename, nil
	})

	ag.RecordScene("scene1", "First")
	ag.RecordScene("scene2", "Second (fails)")
	ag.RecordScene("scene3", "Third")

	album, err := ag.CreateAlbum(context.Background(), "Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All 3 scenes should be present (upload failure skips, doesn't fail).
	if len(album.Scenes) != 3 {
		t.Fatalf("expected 3 scenes despite upload failure, got %d", len(album.Scenes))
	}

	// First and third scenes should have uploaded URLs.
	if album.Scenes[0].ImageURL == "scene1" {
		t.Fatal("expected first scene to have uploaded URL")
	}
	// Second scene should keep original URL (upload failed).
	if album.Scenes[1].ImageURL != "scene2" {
		t.Fatalf("expected second scene to keep original URL, got %q", album.Scenes[1].ImageURL)
	}
}

func TestAlbumGenerator_Concurrency(t *testing.T) {
	ag := NewAlbumGenerator()
	ag.SetPersona("Friend")

	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ag.RecordScene(fmt.Sprintf("scene-%d", n), fmt.Sprintf("Caption %d", n))
			ag.SceneCount()
		}(i)
	}

	wg.Wait()

	if ag.SceneCount() != goroutines {
		t.Fatalf("expected %d scenes, got %d", goroutines, ag.SceneCount())
	}

	// CreateAlbum should also be safe.
	album, err := ag.CreateAlbum(context.Background(), "Concurrent test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(album.Scenes) != goroutines {
		t.Fatalf("expected %d scenes in album, got %d", goroutines, len(album.Scenes))
	}
}

func TestAlbumGenerator_AlbumID_Unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateAlbumID()
		if ids[id] {
			t.Fatalf("duplicate album ID: %s", id)
		}
		ids[id] = true
	}
}
