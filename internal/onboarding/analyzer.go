package onboarding

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Two-Weeks-Team/missless/internal/retry"
)

// VideoAnalysis holds the results of YouTube video analysis.
type VideoAnalysis struct {
	SpeechPatterns       []string    `json:"speechPatterns"`
	Expressions          []string    `json:"expressions"`
	PersonalityTraits    []string    `json:"personalityTraits"`
	Highlights           []Highlight `json:"highlights"`
	VoiceCharacteristics string      `json:"voiceCharacteristics"`
}

// Highlight represents a memorable moment in a video.
type Highlight struct {
	Timestamp   string `json:"timestamp"`
	Description string `json:"description"`
	Expression  string `json:"expression"`
}

// Analyzer analyzes YouTube videos using Gemini 2.5 Pro.
type Analyzer struct {
	// TODO: T08 - Add genai.Client field
}

// AnalyzeYouTubeURL analyzes a YouTube URL directly without downloading.
// Model: gemini-2.5-pro (supports YouTube URL as FileData)
// Only public videos are supported (unlisted → gallery fallback).
func (a *Analyzer) AnalyzeYouTubeURL(ctx context.Context, videoURL, targetPerson string) (*VideoAnalysis, error) {
	slog.Info("youtube_analysis_start", "url", videoURL, "target", targetPerson)

	var result *VideoAnalysis
	err := retry.WithBackoff(ctx, 3, func() error {
		// TODO: T08 - Call gemini-2.5-pro with FileData{FileURI: videoURL}
		// Use ResponseMIMEType: "application/json" for structured output
		// Temperature: 0.2 for factual analysis
		return fmt.Errorf("not yet implemented")
	})

	if err != nil {
		return nil, fmt.Errorf("youtube analysis failed: %w", err)
	}

	return result, nil
}

func buildAnalysisPrompt(targetPerson string) string {
	return fmt.Sprintf(`Analyze "%s" in this video.

Extract:
1. speechPatterns: Speech habits, frequent expressions (minimum 5)
2. expressions: Facial expressions, reactions, gestures (minimum 3)
3. personalityTraits: Character traits (minimum 4)
4. highlights: Memorable moments with timestamps (minimum 3)
5. voiceCharacteristics: Tone, speed, pitch

Focus only on "%s" if multiple people present.
Output JSON only.`, targetPerson, targetPerson)
}
