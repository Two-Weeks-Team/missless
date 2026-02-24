package onboarding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Two-Weeks-Team/missless/internal/retry"
	"google.golang.org/genai"
)

const analysisModel = "gemini-2.5-pro-preview-05-06"

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

// generateFunc abstracts Gemini content generation for testability.
type generateFunc func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)

// Analyzer analyzes YouTube videos using Gemini 2.5 Pro.
type Analyzer struct {
	generate generateFunc
}

// NewAnalyzer creates a new video analyzer with a genai client.
func NewAnalyzer(client *genai.Client) *Analyzer {
	return &Analyzer{
		generate: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return client.Models.GenerateContent(ctx, model, contents, config)
		},
	}
}

// AnalyzeYouTubeURL analyzes a YouTube URL directly without downloading.
// Model: gemini-2.5-pro (supports YouTube URL as FileData)
// Only public videos are supported (unlisted → gallery fallback).
func (a *Analyzer) AnalyzeYouTubeURL(ctx context.Context, videoURL, targetPerson string, progressFn func(string, int)) (*VideoAnalysis, error) {
	slog.Info("youtube_analysis_start", "url", videoURL, "target", targetPerson)
	progressFn("Analyzing video", 10)

	prompt := buildAnalysisPrompt(targetPerson)

	var result *VideoAnalysis
	err := retry.WithBackoff(ctx, 3, func() error {
		progressFn("Sending to Gemini", 20)

		temp := float32(0.2)
		resp, err := a.generate(ctx, analysisModel, []*genai.Content{
			{Parts: []*genai.Part{
				{FileData: &genai.FileData{FileURI: videoURL}},
				genai.NewPartFromText(prompt),
			}},
		}, &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			Temperature:      &temp,
		})
		if err != nil {
			return fmt.Errorf("gemini generate: %w", err)
		}

		progressFn("Parsing results", 70)

		result, err = parseAnalysis(resp)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("youtube analysis failed: %w", err)
	}

	progressFn("Analysis complete", 90)
	slog.Info("youtube_analysis_done", "url", videoURL)
	return result, nil
}

// AnalyzeUploadedFile analyzes a file referenced by Gemini File API URI.
// Used for gallery fallback when videos are unlisted/private.
func (a *Analyzer) AnalyzeUploadedFile(ctx context.Context, fileURI, mimeType, targetPerson string, progressFn func(string, int)) (*VideoAnalysis, error) {
	slog.Info("file_analysis_start", "uri", fileURI, "target", targetPerson)
	progressFn("Analyzing uploaded file", 10)

	prompt := buildAnalysisPrompt(targetPerson)

	var result *VideoAnalysis
	err := retry.WithBackoff(ctx, 3, func() error {
		progressFn("Sending to Gemini", 20)

		temp := float32(0.2)
		resp, err := a.generate(ctx, analysisModel, []*genai.Content{
			{Parts: []*genai.Part{
				{FileData: &genai.FileData{FileURI: fileURI, MIMEType: mimeType}},
				genai.NewPartFromText(prompt),
			}},
		}, &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			Temperature:      &temp,
		})
		if err != nil {
			return fmt.Errorf("gemini generate: %w", err)
		}

		progressFn("Parsing results", 70)

		result, err = parseAnalysis(resp)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("file analysis failed: %w", err)
	}

	progressFn("Analysis complete", 90)
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

// parseAnalysis extracts VideoAnalysis from a Gemini response.
func parseAnalysis(resp *genai.GenerateContentResponse) (*VideoAnalysis, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts in response")
	}

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			var analysis VideoAnalysis
			if err := json.Unmarshal([]byte(part.Text), &analysis); err != nil {
				return nil, fmt.Errorf("parse analysis JSON: %w", err)
			}
			return &analysis, nil
		}
	}

	return nil, fmt.Errorf("no text content in response")
}
