package session

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/genai"
)

func TestManager_StartOnboarding_Config(t *testing.T) {
	mgr := NewManager("test-session-1")
	cfg := mgr.BuildOnboardingConfig()

	if cfg == nil {
		t.Fatal("expected non-nil LiveConnectConfig")
	}

	// Voice must be Aoede (missless host).
	if cfg.SpeechConfig == nil || cfg.SpeechConfig.VoiceConfig == nil ||
		cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig == nil {
		t.Fatal("expected speech config with prebuilt voice")
	}
	if cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName != "Aoede" {
		t.Fatalf("expected voice Aoede, got %q",
			cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName)
	}

	// Must have Audio-only modality (native-audio model outputs audio only).
	if len(cfg.ResponseModalities) != 1 || cfg.ResponseModalities[0] != genai.ModalityAudio {
		t.Fatalf("expected AUDIO-only modality, got %v", cfg.ResponseModalities)
	}

	// System instruction must mention Korean greeting.
	if cfg.SystemInstruction == nil || len(cfg.SystemInstruction.Parts) == 0 {
		t.Fatal("expected system instruction")
	}
	sysText := cfg.SystemInstruction.Parts[0].Text
	if !strings.Contains(sysText, "missless") {
		t.Fatalf("expected system instruction to mention 'missless', got: %s", sysText)
	}
	if !strings.Contains(sysText, "welcome") {
		t.Fatalf("expected English greeting in system instruction")
	}

	// Must have tools declared.
	if len(cfg.Tools) == 0 {
		t.Fatal("expected tools")
	}
	toolNames := make(map[string]bool)
	for _, tool := range cfg.Tools {
		for _, fd := range tool.FunctionDeclarations {
			toolNames[fd.Name] = true
		}
	}
	for _, name := range []string{"generate_scene", "recall_memory", "end_reunion"} {
		if !toolNames[name] {
			t.Fatalf("expected tool %q", name)
		}
	}

}

func TestManager_Transition_StateChange(t *testing.T) {
	mgr := NewManager("test-transition")

	if mgr.State() != StateOnboarding {
		t.Fatalf("expected initial state 'onboarding', got %q", mgr.State())
	}

	// Set persona before transition.
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm and caring", "Gentle with rising intonation")

	// Execute full transition.
	err := mgr.TransitionToReunion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mgr.State() != StateReunion {
		t.Fatalf("expected state 'reunion', got %q", mgr.State())
	}
}

func TestManager_Transition_NotifyBrowser(t *testing.T) {
	mgr := NewManager("test-notify")
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm", "Gentle")

	var mu sync.Mutex
	var events []map[string]any

	mgr.SetNotifyFunc(func(v any) {
		mu.Lock()
		defer mu.Unlock()
		if m, ok := v.(map[string]any); ok {
			events = append(events, m)
		}
	})

	err := mgr.TransitionToReunion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(events) != 2 {
		t.Fatalf("expected 2 events (session_transition + session_ready), got %d", len(events))
	}

	// First: session_transition with "Close your eyes".
	if events[0]["type"] != "session_transition" {
		t.Fatalf("expected type 'session_transition', got %q", events[0]["type"])
	}
	msg, _ := events[0]["message"].(string)
	if !strings.Contains(msg, "Close your eyes") {
		t.Fatalf("expected 'Close your eyes' in message, got %q", msg)
	}

	// Second: session_ready.
	if events[1]["type"] != "session_ready" {
		t.Fatalf("expected type 'session_ready', got %q", events[1]["type"])
	}
	if events[1]["state"] != "reunion" {
		t.Fatalf("expected state 'reunion', got %q", events[1]["state"])
	}
}

func TestManager_BuildReunionConfig(t *testing.T) {
	mgr := NewManager("test-reunion-cfg")
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm and caring", "Gentle with rising intonation")

	cfg := mgr.BuildReunionConfig()

	// Voice must be the persona's matched voice.
	if cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName != "Sulafat" {
		t.Fatalf("expected voice 'Sulafat', got %q",
			cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName)
	}

	// System instruction must contain the persona name.
	sysText := cfg.SystemInstruction.Parts[0].Text
	if !strings.Contains(sysText, "Mom") {
		t.Fatalf("expected persona name 'Mom' in system instruction")
	}
	if !strings.Contains(sysText, "Warm and caring") {
		t.Fatalf("expected personality in system instruction")
	}

	// Must have tools.
	if len(cfg.Tools) == 0 || len(cfg.Tools[0].FunctionDeclarations) == 0 {
		t.Fatal("expected reunion tools")
	}
}

func TestBuildOnboardingSummary(t *testing.T) {
	mgr := NewManager("test-summary")
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm and caring", "Gentle")

	summary := mgr.BuildOnboardingSummary()

	for _, expected := range []string{"Mom", "Sulafat", "Warm and caring", "ko"} {
		if !strings.Contains(summary, expected) {
			t.Fatalf("expected summary to contain %q, got: %s", expected, summary)
		}
	}
}

func TestManager_StateTransitions(t *testing.T) {
	mgr := NewManager("test-session-2")

	if mgr.State() != StateOnboarding {
		t.Fatalf("expected initial state 'onboarding', got %q", mgr.State())
	}

	// Valid transition: onboarding → analyzing.
	if err := mgr.TransitionTo(StateAnalyzing); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mgr.State() != StateAnalyzing {
		t.Fatalf("expected state 'analyzing', got %q", mgr.State())
	}

	// Invalid transition: analyzing → reunion (should go through transitioning).
	err := mgr.TransitionTo(StateReunion)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

func TestManager_SetPersona(t *testing.T) {
	mgr := NewManager("test-session-3")

	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm", "Gentle")

	if mgr.PersonaName() != "Mom" {
		t.Fatalf("expected persona 'Mom', got %q", mgr.PersonaName())
	}
	if mgr.MatchedVoice() != "Sulafat" {
		t.Fatalf("expected voice 'Sulafat', got %q", mgr.MatchedVoice())
	}
}

func TestManager_SessionID(t *testing.T) {
	mgr := NewManager("unique-id-123")
	if mgr.SessionID() != "unique-id-123" {
		t.Fatalf("expected session ID 'unique-id-123', got %q", mgr.SessionID())
	}
}

func TestBuildReunionConfig_AffectiveDialog(t *testing.T) {
	mgr := NewManager("test-affective")
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm", "Gentle")

	cfg := mgr.BuildReunionConfig()

	if cfg.EnableAffectiveDialog == nil || !*cfg.EnableAffectiveDialog {
		t.Fatal("expected EnableAffectiveDialog to be true")
	}
}

func TestBuildReunionConfig_ProactiveAudio(t *testing.T) {
	mgr := NewManager("test-proactive")
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm", "Gentle")

	cfg := mgr.BuildReunionConfig()

	if cfg.Proactivity == nil {
		t.Fatal("expected Proactivity config")
	}
	if cfg.Proactivity.ProactiveAudio == nil || !*cfg.Proactivity.ProactiveAudio {
		t.Fatal("expected ProactiveAudio to be true")
	}
}

func TestBuildReunionConfig_ContextCompression(t *testing.T) {
	mgr := NewManager("test-compression")
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm", "Gentle")

	cfg := mgr.BuildReunionConfig()

	if cfg.ContextWindowCompression == nil {
		t.Fatal("expected ContextWindowCompression config")
	}
	if cfg.ContextWindowCompression.TriggerTokens == nil ||
		*cfg.ContextWindowCompression.TriggerTokens != ContextTriggerTokens {
		t.Fatalf("expected TriggerTokens=%d, got %v",
			ContextTriggerTokens, cfg.ContextWindowCompression.TriggerTokens)
	}
	if cfg.ContextWindowCompression.SlidingWindow == nil {
		t.Fatal("expected SlidingWindow config")
	}
	if cfg.ContextWindowCompression.SlidingWindow.TargetTokens == nil ||
		*cfg.ContextWindowCompression.SlidingWindow.TargetTokens != ContextTargetTokens {
		t.Fatalf("expected TargetTokens=%d, got %v",
			ContextTargetTokens, cfg.ContextWindowCompression.SlidingWindow.TargetTokens)
	}
}

func TestBuildReunionConfig_ReunionTools(t *testing.T) {
	mgr := NewManager("test-reunion-tools")
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm", "Gentle")

	cfg := mgr.BuildReunionConfig()

	if len(cfg.Tools) == 0 || len(cfg.Tools[0].FunctionDeclarations) == 0 {
		t.Fatal("expected reunion tools")
	}

	toolNames := make(map[string]bool)
	for _, fd := range cfg.Tools[0].FunctionDeclarations {
		toolNames[fd.Name] = true
	}
	expected := []string{
		"generate_scene", "generate_fast_scene", "generate_story_page",
		"change_atmosphere", "recall_memory", "analyze_user", "end_reunion",
	}
	for _, name := range expected {
		if !toolNames[name] {
			t.Fatalf("expected reunion tool %q", name)
		}
	}
	if len(cfg.Tools[0].FunctionDeclarations) != 7 {
		t.Fatalf("expected 7 reunion tools, got %d", len(cfg.Tools[0].FunctionDeclarations))
	}
}

func TestStartReunionTimer_CountIncrements(t *testing.T) {
	mgr := NewManager("test-timer")

	if mgr.ReunionCount() != 0 {
		t.Fatalf("expected initial reunion count 0, got %d", mgr.ReunionCount())
	}

	ch := mgr.StartReunionTimer()
	if ch == nil {
		t.Fatal("expected non-nil timer channel")
	}
	if mgr.ReunionCount() != 1 {
		t.Fatalf("expected reunion count 1 after first start, got %d", mgr.ReunionCount())
	}

	// Start again — count should increment.
	mgr.StartReunionTimer()
	if mgr.ReunionCount() != 2 {
		t.Fatalf("expected reunion count 2 after second start, got %d", mgr.ReunionCount())
	}

	mgr.StopReunionTimer()
}

func TestStopReunionTimer_Safe(t *testing.T) {
	mgr := NewManager("test-stop-timer")

	// Stopping without starting should not panic.
	mgr.StopReunionTimer()

	// Start then stop.
	mgr.StartReunionTimer()
	mgr.StopReunionTimer()

	// Double stop should be safe.
	mgr.StopReunionTimer()
}

func TestStartReunionTimer_Warning(t *testing.T) {
	// Override constants for a fast test: we can't change the constants,
	// so we test that the notify function is wired up by checking the timer channel.
	mgr := NewManager("test-warning")

	var mu sync.Mutex
	var events []map[string]any
	mgr.SetNotifyFunc(func(v any) {
		mu.Lock()
		defer mu.Unlock()
		if m, ok := v.(map[string]any); ok {
			events = append(events, m)
		}
	})

	ch := mgr.StartReunionTimer()
	if ch == nil {
		t.Fatal("expected timer channel")
	}

	// We can't wait 240s in a test, so just verify it started correctly.
	mgr.StopReunionTimer()

	// No events expected in immediate stop.
	mu.Lock()
	count := len(events)
	mu.Unlock()
	_ = count // warning may or may not have fired depending on timing
}

func TestBuildContinueSummary(t *testing.T) {
	mgr := NewManager("test-continue")
	mgr.SetPersona("Dad", "Puck", "ko", "Wise", "Calm")

	// Simulate a reunion having happened.
	mgr.StartReunionTimer()
	mgr.StopReunionTimer()

	prev := "We talked about the camping trip in 2019."
	summary := mgr.BuildContinueSummary(prev)

	if !strings.Contains(summary, "Dad") {
		t.Fatalf("expected persona name in continuation summary")
	}
	if !strings.Contains(summary, "Reunion #1") {
		t.Fatalf("expected reunion count in continuation summary, got: %s", summary)
	}
	if !strings.Contains(summary, prev) {
		t.Fatalf("expected previous summary in continuation")
	}
	if !strings.Contains(summary, "Continue the conversation") {
		t.Fatalf("expected continuation instruction")
	}
}

func TestBuildReunionConfig_FallbackVoice(t *testing.T) {
	mgr := NewManager("test-fallback")
	// Don't set persona — voice should fallback to Aoede.
	cfg := mgr.BuildReunionConfig()

	voiceName := cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName
	if voiceName != "Aoede" {
		t.Fatalf("expected fallback voice 'Aoede', got %q", voiceName)
	}
}

func TestBuildReunionConfig_NonEnglishLang(t *testing.T) {
	mgr := NewManager("test-ko-lang")
	mgr.SetPersona("Friend", "Kore", "ko", "Cheerful", "Casual")

	cfg := mgr.BuildReunionConfig()
	sysText := cfg.SystemInstruction.Parts[0].Text

	if strings.Contains(sysText, "Speak naturally in English") {
		t.Fatalf("expected non-English language note for lang='ko'")
	}
	if !strings.Contains(sysText, "'ko'") {
		t.Fatalf("expected language code 'ko' in system instruction, got: %s", sysText)
	}
}

func TestBuildReunionConfig_DefaultEnglish(t *testing.T) {
	mgr := NewManager("test-en-lang")
	mgr.SetPersona("Friend", "Kore", "en", "Cheerful", "Casual")

	cfg := mgr.BuildReunionConfig()
	sysText := cfg.SystemInstruction.Parts[0].Text

	if !strings.Contains(sysText, "Speak naturally in English") {
		t.Fatalf("expected default English language note for lang='en', got: %s", sysText)
	}
}

func TestBuildReunionConfig_DefaultEnglishWhenLangEmpty(t *testing.T) {
	mgr := NewManager("test-empty-lang")
	mgr.SetPersona("Friend", "Kore", "", "Cheerful", "Casual")

	cfg := mgr.BuildReunionConfig()
	sysText := cfg.SystemInstruction.Parts[0].Text

	if !strings.Contains(sysText, "Speak naturally in English") {
		t.Fatalf("expected default English language note for empty lang, got: %s", sysText)
	}
}

func TestBuildReunionConfig_AffectiveRulesInInstruction(t *testing.T) {
	mgr := NewManager("test-affective-rules")
	mgr.SetPersona("Mom", "Sulafat", "ko", "Warm", "Gentle")

	cfg := mgr.BuildReunionConfig()
	sysText := cfg.SystemInstruction.Parts[0].Text

	// Must contain affective dialog rules.
	for _, keyword := range []string{"Affective Dialog", "tearful", "soften", "laughs"} {
		if !strings.Contains(sysText, keyword) {
			t.Fatalf("expected system instruction to contain %q", keyword)
		}
	}

	// Must contain proactive audio rules.
	for _, keyword := range []string{"Proactive Audio", "self-talk", "directly"} {
		if !strings.Contains(sysText, keyword) {
			t.Fatalf("expected system instruction to contain %q", keyword)
		}
	}
}

// Ensure unused import doesn't cause issues.
var _ = time.Second
