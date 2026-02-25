package scene

import (
	"sync"
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

func TestCharacterAnchor_UpdateLastScene(t *testing.T) {
	anchor := NewCharacterAnchor()

	// Initially empty.
	if anchor.GetLastScene() != "" {
		t.Fatal("expected empty last scene initially")
	}

	// Set scene.
	anchor.UpdateLastScene("scene-1-base64")
	if anchor.GetLastScene() != "scene-1-base64" {
		t.Fatalf("expected 'scene-1-base64', got %q", anchor.GetLastScene())
	}

	// Overwrite with new scene.
	anchor.UpdateLastScene("scene-2-base64")
	if anchor.GetLastScene() != "scene-2-base64" {
		t.Fatalf("expected 'scene-2-base64', got %q", anchor.GetLastScene())
	}
}

func TestCharacterAnchor_GetRefParts_WithImages(t *testing.T) {
	anchor := NewCharacterAnchor()
	anchor.AddRefImage([]byte{0xFF, 0xD8, 0x01})
	anchor.AddRefImage([]byte{0xFF, 0xD8, 0x02})
	anchor.AddRefImage([]byte{0xFF, 0xD8, 0x03})

	parts := anchor.GetRefParts()
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}

	for i, part := range parts {
		if part.InlineData == nil {
			t.Fatalf("part[%d]: expected InlineData", i)
		}
		if part.InlineData.MIMEType != "image/jpeg" {
			t.Fatalf("part[%d]: expected MIME 'image/jpeg', got %q", i, part.InlineData.MIMEType)
		}
		if len(part.InlineData.Data) == 0 {
			t.Fatalf("part[%d]: expected non-empty data", i)
		}
	}

	// Verify data matches the original images.
	if parts[0].InlineData.Data[2] != 0x01 {
		t.Fatalf("expected first image data byte 0x01, got %x", parts[0].InlineData.Data[2])
	}
	if parts[2].InlineData.Data[2] != 0x03 {
		t.Fatalf("expected third image data byte 0x03, got %x", parts[2].InlineData.Data[2])
	}
}

func TestCharacterAnchor_GetRefParts_Empty(t *testing.T) {
	anchor := NewCharacterAnchor()

	parts := anchor.GetRefParts()
	if parts != nil {
		t.Fatalf("expected nil parts for empty anchor, got %d parts", len(parts))
	}
}

func TestCharacterAnchor_Concurrency(t *testing.T) {
	anchor := NewCharacterAnchor()

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent writes and reads — should not race.
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			anchor.AddRefImage([]byte{byte(n)})
			anchor.GetRefImages()
			anchor.GetRefParts()
			anchor.UpdateLastScene("scene-from-goroutine")
			anchor.GetLastScene()
			anchor.SilhouetteGuide()
		}(i)
	}

	wg.Wait()

	// Verify FIFO: at most maxRefImages.
	images := anchor.GetRefImages()
	if len(images) > maxRefImages {
		t.Fatalf("expected at most %d images, got %d", maxRefImages, len(images))
	}
}

func TestCharacterAnchor_SilhouetteGuide(t *testing.T) {
	anchor := NewCharacterAnchor()
	guide := anchor.SilhouetteGuide()

	if guide == "" {
		t.Fatal("expected non-empty silhouette guide")
	}
	if !contains(guide, "silhouette") {
		t.Fatal("expected 'silhouette' in guide")
	}
	if !contains(guide, "back") {
		t.Fatal("expected 'back' in guide")
	}
}

func TestBuildPrompt_WithSilhouetteGuide(t *testing.T) {
	gen := &Generator{anchor: NewCharacterAnchor()}
	gen.anchor.AddRefImage([]byte{0xFF, 0xD8, 0x01})

	prompt := gen.buildPrompt("park scene", "warm", "family")

	if !contains(prompt, "silhouette") {
		t.Fatalf("expected silhouette guide in prompt when ref images exist, got: %s", prompt)
	}
}

func TestBuildPrompt_NoSilhouetteWithoutImages(t *testing.T) {
	gen := &Generator{anchor: NewCharacterAnchor()}

	prompt := gen.buildPrompt("park scene", "warm", "family")

	if contains(prompt, "silhouette") {
		t.Fatalf("expected no silhouette guide without ref images, got: %s", prompt)
	}
}

func TestExtractInterleaved_TextAndImage(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "A warm reunion scene."},
						{
							InlineData: &genai.Blob{
								MIMEType: "image/jpeg",
								Data:     []byte{0xFF, 0xD8, 0xFF, 0xE0},
							},
						},
						{Text: " The sun sets gently."},
					},
				},
			},
		},
	}

	result, err := extractInterleaved(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text == "" {
		t.Fatal("expected non-empty text")
	}
	if result.ImageBase64 == "" {
		t.Fatal("expected non-empty image")
	}
	if !contains(result.Text, "warm reunion") {
		t.Fatalf("expected text to contain narration, got: %s", result.Text)
	}
}

func TestExtractInterleaved_TextOnly(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "Only narration, no image."},
					},
				},
			},
		},
	}

	result, err := extractInterleaved(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text == "" {
		t.Fatal("expected non-empty text")
	}
	if result.ImageBase64 != "" {
		t.Fatal("expected empty image for text-only response")
	}
}

func TestExtractInterleaved_NilResponse(t *testing.T) {
	_, err := extractInterleaved(nil)
	if err == nil {
		t.Fatal("expected error for nil response")
	}
}

func TestExtractInterleaved_NilContent(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: nil},
		},
	}

	_, err := extractInterleaved(resp)
	if err == nil {
		t.Fatal("expected error for nil content")
	}
}

func TestExtractInterleaved_Empty(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{},
				},
			},
		},
	}

	_, err := extractInterleaved(resp)
	if err == nil {
		t.Fatal("expected error for empty parts")
	}
}

func TestBuildStoryPagePrompt(t *testing.T) {
	gen := &Generator{anchor: NewCharacterAnchor()}

	prompt := gen.buildStoryPagePrompt("dinner table scene", "warm", "mother and child")

	for _, want := range []string{"story page", "dinner table", "warm", "mother and child", "narration", "watercolor"} {
		if !contains(prompt, want) {
			t.Fatalf("expected story page prompt to contain %q, got: %s", want, prompt)
		}
	}

	// Verify prompt injection hardening: user inputs wrapped in triple-quote delimiters.
	if !contains(prompt, `"""dinner table scene"""`) {
		t.Fatalf("expected triple-quote delimiters around scene input, got: %s", prompt)
	}
}

func TestBuildStoryPagePrompt_Minimal(t *testing.T) {
	gen := &Generator{anchor: NewCharacterAnchor()}

	prompt := gen.buildStoryPagePrompt("park scene", "", "")

	if !contains(prompt, "park scene") {
		t.Fatalf("expected prompt to contain scene, got: %s", prompt)
	}
	if !contains(prompt, "story page") {
		t.Fatalf("expected prompt to reference story page, got: %s", prompt)
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
