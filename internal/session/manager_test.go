package session

import (
	"strings"
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

	mgr.SetPersona("Mom", "Sulafat", "ko")

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
