package scene

import (
	"testing"

	"google.golang.org/genai"
)

func TestGenerator_BuildPrompt_WithRef(t *testing.T) {
	gen := &Generator{anchor: NewCharacterAnchor()}
	gen.anchor.UpdateLastScene("prevSceneBase64")

	prompt := gen.buildPrompt("sunset on the beach", "warm", "mother and daughter")

	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	// Should contain all parts
	for _, want := range []string{"sunset on the beach", "warm", "mother and daughter", "continuity"} {
		if !contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q, got: %s", want, prompt)
		}
	}
}

func TestGenerator_BuildPrompt_NoRef(t *testing.T) {
	gen := &Generator{anchor: NewCharacterAnchor()}

	prompt := gen.buildPrompt("rainy day", "nostalgic", "")

	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	// Should contain mood but not continuity (no last scene)
	if !contains(prompt, "nostalgic") {
		t.Fatalf("expected prompt to contain mood, got: %s", prompt)
	}
	if contains(prompt, "continuity") {
		t.Fatalf("expected no continuity reference without last scene, got: %s", prompt)
	}
}

func TestGenerator_BuildPrompt_NilAnchor(t *testing.T) {
	gen := &Generator{}

	prompt := gen.buildPrompt("test prompt", "happy", "")

	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	if !contains(prompt, "test prompt") {
		t.Fatalf("expected prompt to contain base text, got: %s", prompt)
	}
}

func TestExtractImageBase64_NoImage(t *testing.T) {
	// Empty response
	_, err := extractImageBase64(nil)
	if err == nil {
		t.Fatal("expected error for nil response")
	}

	// Response with no candidates
	_, err = extractImageBase64(&genai.GenerateContentResponse{})
	if err == nil {
		t.Fatal("expected error for empty candidates")
	}

	// Response with candidate but no image
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "some text, no image"},
					},
				},
			},
		},
	}
	_, err = extractImageBase64(resp)
	if err == nil {
		t.Fatal("expected error for text-only response")
	}
}

func TestExtractImageBase64_WithImage(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{
							InlineData: &genai.Blob{
								MIMEType: "image/jpeg",
								Data:     []byte{0xFF, 0xD8, 0xFF, 0xE0},
							},
						},
					},
				},
			},
		},
	}

	b64, err := extractImageBase64(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b64 == "" {
		t.Fatal("expected non-empty base64 string")
	}
}

func TestCharacterAnchor_RefImages(t *testing.T) {
	anchor := NewCharacterAnchor()

	// Add images up to max
	for i := 0; i < 7; i++ {
		anchor.AddRefImage([]byte{byte(i)})
	}

	images := anchor.GetRefImages()
	if len(images) != maxRefImages {
		t.Fatalf("expected %d ref images, got %d", maxRefImages, len(images))
	}

	// Should keep the most recent 5 (indices 2-6)
	if images[0][0] != 2 {
		t.Fatalf("expected first image byte 2 (FIFO eviction), got %d", images[0][0])
	}
}

func TestCharacterAnchor_LastScene(t *testing.T) {
	anchor := NewCharacterAnchor()

	if anchor.GetLastScene() != "" {
		t.Fatal("expected empty last scene initially")
	}

	anchor.UpdateLastScene("test-scene-base64")
	if anchor.GetLastScene() != "test-scene-base64" {
		t.Fatalf("expected 'test-scene-base64', got %q", anchor.GetLastScene())
	}
}

func TestNewGenerator(t *testing.T) {
	// NewGenerator requires a non-nil client in production, but for unit tests
	// we verify the struct is properly initialized.
	gen := &Generator{anchor: NewCharacterAnchor()}

	if gen.anchor == nil {
		t.Fatal("expected anchor to be set")
	}

	gen.SetAnchor(NewCharacterAnchor())
	anchor := gen.getAnchor()
	if anchor == nil {
		t.Fatal("expected anchor after SetAnchor")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
