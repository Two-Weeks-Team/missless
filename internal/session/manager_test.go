package session

import (
	"context"
	"strings"
	"sync"
	"testing"

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

	// Must have Audio + Text modalities.
	if len(cfg.ResponseModalities) != 2 {
		t.Fatalf("expected 2 modalities, got %d", len(cfg.ResponseModalities))
	}
	hasAudio, hasText := false, false
	for _, m := range cfg.ResponseModalities {
		if m == genai.ModalityAudio {
			hasAudio = true
		}
		if m == genai.ModalityText {
			hasText = true
		}
	}
	if !hasAudio || !hasText {
		t.Fatal("expected both AUDIO and TEXT modalities")
	}

	// System instruction must mention Korean greeting.
	if cfg.SystemInstruction == nil || len(cfg.SystemInstruction.Parts) == 0 {
		t.Fatal("expected system instruction")
	}
	sysText := cfg.SystemInstruction.Parts[0].Text
	if !strings.Contains(sysText, "missless") {
		t.Fatalf("expected system instruction to mention 'missless', got: %s", sysText)
	}
	if !strings.Contains(sysText, "환영") {
		t.Fatalf("expected Korean greeting in system instruction")
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

	// Session resumption must be transparent.
	if cfg.SessionResumption == nil || !cfg.SessionResumption.Transparent {
		t.Fatal("expected transparent session resumption")
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

	// First: session_transition with "눈 감아보세요".
	if events[0]["type"] != "session_transition" {
		t.Fatalf("expected type 'session_transition', got %q", events[0]["type"])
	}
	msg, _ := events[0]["message"].(string)
	if !strings.Contains(msg, "눈을 감아보세요") {
		t.Fatalf("expected '눈을 감아보세요' in message, got %q", msg)
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
