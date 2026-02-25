package handler

import (
	"testing"

	"github.com/Two-Weeks-Team/missless/internal/session"
	"google.golang.org/genai"
)

func TestBuildOnboardingConfig(t *testing.T) {
	mgr := session.NewManager("test-session")
	cfg := mgr.BuildOnboardingConfig()

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// Check response modalities — native-audio model is audio-only.
	if len(cfg.ResponseModalities) != 1 {
		t.Fatalf("expected 1 modality, got %d", len(cfg.ResponseModalities))
	}

	// Check voice config
	if cfg.SpeechConfig == nil {
		t.Fatal("expected SpeechConfig")
	}
	if cfg.SpeechConfig.VoiceConfig == nil {
		t.Fatal("expected VoiceConfig")
	}
	if cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig == nil {
		t.Fatal("expected PrebuiltVoiceConfig")
	}
	if cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName != "Aoede" {
		t.Fatalf("expected voice Aoede, got %s",
			cfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName)
	}

	// Check system instruction
	if cfg.SystemInstruction == nil {
		t.Fatal("expected SystemInstruction")
	}
	if len(cfg.SystemInstruction.Parts) == 0 {
		t.Fatal("expected system instruction parts")
	}

	// Check tools
	if len(cfg.Tools) == 0 {
		t.Fatal("expected tools")
	}

	toolNames := make(map[string]bool)
	for _, tool := range cfg.Tools {
		for _, fd := range tool.FunctionDeclarations {
			toolNames[fd.Name] = true
		}
	}

	expected := []string{"generate_scene", "generate_fast_scene", "change_atmosphere", "recall_memory", "analyze_user", "end_reunion"}
	for _, name := range expected {
		if !toolNames[name] {
			t.Fatalf("expected tool %s in declarations", name)
		}
	}
}

func TestBuildOnboardingConfig_Modalities(t *testing.T) {
	mgr := session.NewManager("test-session")
	cfg := mgr.BuildOnboardingConfig()

	// Native-audio model requires audio-only output.
	if len(cfg.ResponseModalities) != 1 || cfg.ResponseModalities[0] != genai.ModalityAudio {
		t.Fatalf("expected AUDIO-only modality, got %v", cfg.ResponseModalities)
	}
}
