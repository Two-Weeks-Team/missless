package onboarding

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Two-Weeks-Team/missless/internal/retry"
)

// PresetVoice represents one of the 30 HD preset voices (V7 verified).
type PresetVoice struct {
	Name    string
	Gender  string
	Tone    string
	AgeHint string
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
	// TODO: T09 - Add genai.Client field
}

// MatchFromAnalyses creates a Persona with matched voice.
// Model: gemini-2.5-pro (structured JSON output)
func (vm *VoiceMatcher) MatchFromAnalyses(ctx context.Context, analyses []*VideoAnalysis, targetPerson string) (*Persona, error) {
	slog.Info("voice_matching_start", "target", targetPerson, "analyses", len(analyses))

	var persona *Persona
	err := retry.WithBackoff(ctx, 3, func() error {
		// TODO: T09 - Call gemini-2.5-pro to match voice
		// Include all 30 preset voices with metadata
		// Temperature: 0.3 for balanced creativity
		return fmt.Errorf("not yet implemented")
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
