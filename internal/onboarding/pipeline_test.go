package onboarding

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"google.golang.org/genai"
)

// mockAnalysisJSON is a valid analysis response for tests.
const mockAnalysisJSON = `{
	"speechPatterns": ["says 'you know'", "rising intonation"],
	"expressions": ["warm smile", "head nod"],
	"personalityTraits": ["caring", "patient", "humorous", "wise"],
	"highlights": [{"timestamp":"0:30","description":"laughing","expression":"joy"}],
	"voiceCharacteristics": "warm female alto, moderate pace"
}`

// mockPersonaJSON is a valid persona response for tests.
const mockPersonaJSON = `{
	"name": "Mom",
	"personality": "Warm and caring",
	"speechStyle": "Gentle with rising intonation",
	"matchedVoice": "Sulafat",
	"frequentPhrases": ["you know", "sweetie"],
	"emotionalPatterns": "nurturing and supportive",
	"languageCode": "ko"
}`

func mockGenerateResponse(jsonText string) generateFunc {
	return func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		return &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{Content: &genai.Content{
					Parts: []*genai.Part{{Text: jsonText}},
				}},
			},
		}, nil
	}
}

func mockGenerateError(errMsg string) generateFunc {
	return func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		return nil, fmt.Errorf("%s", errMsg)
	}
}

func TestPipeline_Run_Success(t *testing.T) {
	analyzer := &Analyzer{generate: mockGenerateResponse(mockAnalysisJSON)}
	matcher := &VoiceMatcher{generate: mockGenerateResponse(mockPersonaJSON)}
	pipeline := NewPipeline(analyzer, matcher)

	persona, err := pipeline.Run(
		context.Background(),
		[]string{"https://youtube.com/watch?v=test1"},
		"Mom",
		func(string, int) {},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if persona == nil {
		t.Fatal("expected non-nil persona")
	}
	if persona.Name != "Mom" {
		t.Fatalf("expected name 'Mom', got %q", persona.Name)
	}
	if persona.MatchedVoice != "Sulafat" {
		t.Fatalf("expected voice 'Sulafat', got %q", persona.MatchedVoice)
	}
	if persona.LanguageCode != "ko" {
		t.Fatalf("expected languageCode 'ko', got %q", persona.LanguageCode)
	}
}

func TestPipeline_Run_NoVideos(t *testing.T) {
	analyzer := &Analyzer{generate: mockGenerateResponse(mockAnalysisJSON)}
	matcher := &VoiceMatcher{generate: mockGenerateResponse(mockPersonaJSON)}
	pipeline := NewPipeline(analyzer, matcher)

	_, err := pipeline.Run(
		context.Background(),
		[]string{}, // Empty video list
		"Mom",
		func(string, int) {},
	)

	if err == nil {
		t.Fatal("expected error for empty video list")
	}
	if !strings.Contains(err.Error(), "no successful video analyses") {
		t.Fatalf("expected 'no successful video analyses' error, got: %v", err)
	}
}

func TestPipeline_Run_AllFail(t *testing.T) {
	analyzer := &Analyzer{generate: mockGenerateError("api unavailable")}
	matcher := &VoiceMatcher{generate: mockGenerateResponse(mockPersonaJSON)}
	pipeline := NewPipeline(analyzer, matcher)

	_, err := pipeline.Run(
		context.Background(),
		[]string{"https://youtube.com/watch?v=test1"},
		"Mom",
		func(string, int) {},
	)

	if err == nil {
		t.Fatal("expected error when all analyses fail")
	}
}

func TestVoiceMatcher_GenderFilter(t *testing.T) {
	female := FilterVoicesByGender("female")
	male := FilterVoicesByGender("male")

	if len(female) != 14 {
		t.Fatalf("expected 14 female voices, got %d", len(female))
	}
	if len(male) != 16 {
		t.Fatalf("expected 16 male voices, got %d", len(male))
	}

	// Verify no gender mixing.
	for _, v := range female {
		if v.Gender != "female" {
			t.Fatalf("expected female, got %s for %s", v.Gender, v.Name)
		}
	}
	for _, v := range male {
		if v.Gender != "male" {
			t.Fatalf("expected male, got %s for %s", v.Gender, v.Name)
		}
	}
}

func TestVoiceMatcher_30Presets(t *testing.T) {
	if len(PresetVoices) != 30 {
		t.Fatalf("expected 30 preset voices, got %d", len(PresetVoices))
	}

	// Check uniqueness.
	seen := make(map[string]bool)
	for _, v := range PresetVoices {
		if seen[v.Name] {
			t.Fatalf("duplicate voice name: %s", v.Name)
		}
		seen[v.Name] = true

		if v.Gender == "" || v.Tone == "" || v.AgeHint == "" {
			t.Fatalf("voice %s has empty fields", v.Name)
		}
	}

	// Verify isValidVoice works.
	if !isValidVoice("Aoede") {
		t.Fatal("expected Aoede to be valid")
	}
	if isValidVoice("NonexistentVoice") {
		t.Fatal("expected NonexistentVoice to be invalid")
	}
}

func TestPipeline_ProgressFn(t *testing.T) {
	analyzer := &Analyzer{generate: mockGenerateResponse(mockAnalysisJSON)}
	matcher := &VoiceMatcher{generate: mockGenerateResponse(mockPersonaJSON)}
	pipeline := NewPipeline(analyzer, matcher)

	var mu sync.Mutex
	var percents []int

	_, err := pipeline.Run(
		context.Background(),
		[]string{"https://youtube.com/watch?v=test1"},
		"Mom",
		func(step string, pct int) {
			mu.Lock()
			percents = append(percents, pct)
			mu.Unlock()
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(percents) < 3 {
		t.Fatalf("expected at least 3 progress updates, got %d", len(percents))
	}

	// First should be 0%, last should be 100%.
	if percents[0] != 0 {
		t.Fatalf("expected first progress 0%%, got %d%%", percents[0])
	}
	if percents[len(percents)-1] != 100 {
		t.Fatalf("expected last progress 100%%, got %d%%", percents[len(percents)-1])
	}
}
