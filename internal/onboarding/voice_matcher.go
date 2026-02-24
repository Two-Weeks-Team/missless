package onboarding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Two-Weeks-Team/missless/internal/retry"
	"google.golang.org/genai"
)

// PresetVoice represents one of the 30 HD preset voices (V7 verified).
type PresetVoice struct {
	Name    string `json:"name"`
	Gender  string `json:"gender"`
	Tone    string `json:"tone"`
	AgeHint string `json:"ageHint"`
}

// All 30 HD preset voices available in Gemini Live API.
var PresetVoices = []PresetVoice{
	// Female (14)
	{Name: "Achernar", Gender: "female", Tone: "soft", AgeHint: "mature"},
	{Name: "Aoede", Gender: "female", Tone: "breezy", AgeHint: "young"},
	{Name: "Autonoe", Gender: "female", Tone: "bright", AgeHint: "young"},
	{Name: "Callirrhoe", Gender: "female", Tone: "easy-going", AgeHint: "young"},
	{Name: "Despina", Gender: "female", Tone: "smooth", AgeHint: "mature"},
	{Name: "Erinome", Gender: "female", Tone: "clear", AgeHint: "young"},
	{Name: "Gacrux", Gender: "female", Tone: "mature", AgeHint: "mature"},
	{Name: "Kore", Gender: "female", Tone: "firm", AgeHint: "young"},
	{Name: "Laomedeia", Gender: "female", Tone: "upbeat", AgeHint: "young"},
	{Name: "Leda", Gender: "female", Tone: "youthful", AgeHint: "young"},
	{Name: "Pulcherrima", Gender: "female", Tone: "forward", AgeHint: "young"},
	{Name: "Sulafat", Gender: "female", Tone: "warm", AgeHint: "mature"},
	{Name: "Vindemiatrix", Gender: "female", Tone: "gentle", AgeHint: "mature"},
	{Name: "Zephyr", Gender: "female", Tone: "bright", AgeHint: "young"},
	// Male (16)
	{Name: "Achird", Gender: "male", Tone: "friendly", AgeHint: "young"},
	{Name: "Algenib", Gender: "male", Tone: "gravelly", AgeHint: "mature"},
	{Name: "Algieba", Gender: "male", Tone: "smooth", AgeHint: "mature"},
	{Name: "Alnilam", Gender: "male", Tone: "firm", AgeHint: "mature"},
	{Name: "Charon", Gender: "male", Tone: "informative", AgeHint: "mature"},
	{Name: "Enceladus", Gender: "male", Tone: "breathy", AgeHint: "young"},
	{Name: "Fenrir", Gender: "male", Tone: "excitable", AgeHint: "young"},
	{Name: "Iapetus", Gender: "male", Tone: "clear", AgeHint: "young"},
	{Name: "Orus", Gender: "male", Tone: "firm", AgeHint: "mature"},
	{Name: "Puck", Gender: "male", Tone: "upbeat", AgeHint: "young"},
	{Name: "Rasalgethi", Gender: "male", Tone: "informative", AgeHint: "mature"},
	{Name: "Sadachbia", Gender: "male", Tone: "lively", AgeHint: "young"},
	{Name: "Sadaltager", Gender: "male", Tone: "knowledgeable", AgeHint: "mature"},
	{Name: "Schedar", Gender: "male", Tone: "even", AgeHint: "mature"},
	{Name: "Umbriel", Gender: "male", Tone: "easy-going", AgeHint: "young"},
	{Name: "Zubenelgenubi", Gender: "male", Tone: "casual", AgeHint: "young"},
}

// VoiceMatcher maps video analysis results to optimal HD preset voice.
type VoiceMatcher struct {
	generate generateFunc
}

// NewVoiceMatcher creates a new voice matcher with a genai client.
func NewVoiceMatcher(client *genai.Client) *VoiceMatcher {
	return &VoiceMatcher{
		generate: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return client.Models.GenerateContent(ctx, model, contents, config)
		},
	}
}

// MatchFromAnalyses creates a Persona with matched voice from video analyses.
// Model: gemini-2.5-pro (structured JSON output)
func (vm *VoiceMatcher) MatchFromAnalyses(ctx context.Context, analyses []*VideoAnalysis, targetPerson string) (*Persona, error) {
	slog.Info("voice_matching_start", "target", targetPerson, "analyses", len(analyses))

	prompt := buildVoiceMatchPrompt(analyses, targetPerson)

	var persona *Persona
	err := retry.WithBackoff(ctx, 3, func() error {
		temp := float32(0.3)
		resp, err := vm.generate(ctx, analysisModel, []*genai.Content{
			{Parts: []*genai.Part{genai.NewPartFromText(prompt)}},
		}, &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			Temperature:      &temp,
		})
		if err != nil {
			return fmt.Errorf("gemini generate: %w", err)
		}

		persona, err = parsePersona(resp)
		if err != nil {
			return err
		}

		// Validate the matched voice exists in presets.
		if !isValidVoice(persona.MatchedVoice) {
			slog.Warn("invalid_voice_fallback",
				"matched", persona.MatchedVoice,
				"fallback", "Aoede",
			)
			persona.MatchedVoice = "Aoede"
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("voice matching failed: %w", err)
	}

	slog.Info("voice_matched",
		"persona", persona.Name,
		"voice", persona.MatchedVoice,
	)
	return persona, nil
}

// FilterVoicesByGender returns voices matching the given gender.
func FilterVoicesByGender(gender string) []PresetVoice {
	var filtered []PresetVoice
	for _, v := range PresetVoices {
		if v.Gender == gender {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func buildVoiceMatchPrompt(analyses []*VideoAnalysis, targetPerson string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Based on the following analysis of \"%s\":\n\n", targetPerson))

	for i, a := range analyses {
		sb.WriteString(fmt.Sprintf("--- Analysis %d ---\n", i+1))
		sb.WriteString(fmt.Sprintf("Voice: %s\n", a.VoiceCharacteristics))
		sb.WriteString(fmt.Sprintf("Personality: %s\n", strings.Join(a.PersonalityTraits, ", ")))
		sb.WriteString(fmt.Sprintf("Speech patterns: %s\n", strings.Join(a.SpeechPatterns, ", ")))
		sb.WriteString(fmt.Sprintf("Expressions: %s\n\n", strings.Join(a.Expressions, ", ")))
	}

	sb.WriteString("Select the best matching HD preset voice from this list:\n")
	for _, v := range PresetVoices {
		sb.WriteString(fmt.Sprintf("- %s (gender: %s, tone: %s, age: %s)\n", v.Name, v.Gender, v.Tone, v.AgeHint))
	}

	sb.WriteString(`
Create a persona JSON with these fields:
- name: the person's name
- personality: brief personality description
- speechStyle: speaking style summary
- matchedVoice: selected preset voice name (MUST be from the list above)
- frequentPhrases: array of common phrases
- emotionalPatterns: emotional tendencies
- languageCode: primary language code (e.g., "ko" for Korean)

Match voice gender and tone to the analyzed person's characteristics.
Output JSON only.`)

	return sb.String()
}

func parsePersona(resp *genai.GenerateContentResponse) (*Persona, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts in response")
	}

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			var persona Persona
			if err := json.Unmarshal([]byte(part.Text), &persona); err != nil {
				return nil, fmt.Errorf("parse persona JSON: %w", err)
			}
			return &persona, nil
		}
	}

	return nil, fmt.Errorf("no text content in response")
}

func isValidVoice(name string) bool {
	for _, v := range PresetVoices {
		if v.Name == name {
			return true
		}
	}
	return false
}
