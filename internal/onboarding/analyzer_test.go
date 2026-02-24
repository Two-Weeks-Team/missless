package onboarding

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"google.golang.org/genai"
)

func TestAnalyzer_BuildAnalysisPrompt(t *testing.T) {
	prompt := buildAnalysisPrompt("Mom")

	if !strings.Contains(prompt, "Mom") {
		t.Fatalf("expected prompt to contain targetPerson 'Mom', got: %s", prompt)
	}
	for _, keyword := range []string{"speechPatterns", "expressions", "personalityTraits", "highlights", "voiceCharacteristics"} {
		if !strings.Contains(prompt, keyword) {
			t.Fatalf("expected prompt to contain %q, got: %s", keyword, prompt)
		}
	}
}

func TestAnalyzer_ParseAnalysis_ValidJSON(t *testing.T) {
	validJSON := `{
		"speechPatterns": ["always says 'you know'", "ends with rising tone"],
		"expressions": ["warm smile", "head tilt when listening"],
		"personalityTraits": ["caring", "patient", "humorous", "wise"],
		"highlights": [
			{"timestamp": "0:42", "description": "laughing at joke", "expression": "joy"},
			{"timestamp": "1:15", "description": "telling a story", "expression": "nostalgia"},
			{"timestamp": "2:30", "description": "giving advice", "expression": "warmth"}
		],
		"voiceCharacteristics": "warm alto, moderate pace, gentle tone"
	}`

	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: validJSON},
					},
				},
			},
		},
	}

	analysis, err := parseAnalysis(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(analysis.SpeechPatterns) != 2 {
		t.Fatalf("expected 2 speech patterns, got %d", len(analysis.SpeechPatterns))
	}
	if len(analysis.PersonalityTraits) != 4 {
		t.Fatalf("expected 4 personality traits, got %d", len(analysis.PersonalityTraits))
	}
	if len(analysis.Highlights) != 3 {
		t.Fatalf("expected 3 highlights, got %d", len(analysis.Highlights))
	}
	if analysis.VoiceCharacteristics == "" {
		t.Fatal("expected non-empty voice characteristics")
	}
}

func TestAnalyzer_ParseAnalysis_InvalidJSON(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "this is not valid JSON {{{"},
					},
				},
			},
		},
	}

	_, err := parseAnalysis(resp)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse analysis JSON") {
		t.Fatalf("expected JSON parse error, got: %v", err)
	}
}

func TestAnalyzer_RetryOnFailure(t *testing.T) {
	callCount := 0

	validJSON := `{"speechPatterns":["test"],"expressions":["smile"],"personalityTraits":["kind","warm","funny","smart"],"highlights":[{"timestamp":"0:00","description":"intro","expression":"happy"}],"voiceCharacteristics":"warm"}`

	analyzer := &Analyzer{
		generate: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			callCount++
			if callCount < 3 {
				return nil, fmt.Errorf("transient error attempt %d", callCount)
			}
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: validJSON},
							},
						},
					},
				},
			}, nil
		},
	}

	result, err := analyzer.AnalyzeYouTubeURL(
		context.Background(),
		"https://www.youtube.com/watch?v=test",
		"TestPerson",
		func(string, int) {},
	)

	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if callCount < 3 {
		t.Fatalf("expected at least 3 calls (2 failures + 1 success), got %d", callCount)
	}
}

func TestAnalyzer_ProgressCallback(t *testing.T) {
	validJSON := `{"speechPatterns":["test"],"expressions":["smile"],"personalityTraits":["kind","warm","funny","smart"],"highlights":[{"timestamp":"0:00","description":"test","expression":"happy"}],"voiceCharacteristics":"warm"}`

	analyzer := &Analyzer{
		generate: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: validJSON},
							},
						},
					},
				},
			}, nil
		},
	}

	var mu sync.Mutex
	var steps []string
	var percents []int

	_, err := analyzer.AnalyzeYouTubeURL(
		context.Background(),
		"https://www.youtube.com/watch?v=test",
		"TestPerson",
		func(step string, pct int) {
			mu.Lock()
			steps = append(steps, step)
			percents = append(percents, pct)
			mu.Unlock()
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have at least 4 progress updates: analyzing, sending, parsing, complete.
	if len(steps) < 4 {
		t.Fatalf("expected at least 4 progress callbacks, got %d: %v", len(steps), steps)
	}

	// Verify progress is non-decreasing.
	for i := 1; i < len(percents); i++ {
		if percents[i] < percents[i-1] {
			t.Fatalf("progress decreased: %d → %d at step %d", percents[i-1], percents[i], i)
		}
	}
}
