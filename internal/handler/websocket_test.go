package handler

import (
	"testing"

	"google.golang.org/genai"
)

func TestBuildOnboardingConfig(t *testing.T) {
	cfg := buildOnboardingConfig()

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// Check response modalities
	if len(cfg.ResponseModalities) != 2 {
		t.Fatalf("expected 2 modalities, got %d", len(cfg.ResponseModalities))
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

	expected := []string{"generate_scene", "change_atmosphere", "end_reunion"}
	for _, name := range expected {
		if !toolNames[name] {
			t.Fatalf("expected tool %s in declarations", name)
		}
	}

	// Check session resumption
	if cfg.SessionResumption == nil {
		t.Fatal("expected SessionResumption config")
	}
	if !cfg.SessionResumption.Transparent {
		t.Fatal("expected transparent session resumption")
	}
}

func TestBuildOnboardingConfig_Modalities(t *testing.T) {
	cfg := buildOnboardingConfig()

	hasAudio := false
	hasText := false
	for _, m := range cfg.ResponseModalities {
		if m == genai.ModalityAudio {
			hasAudio = true
		}
		if m == genai.ModalityText {
			hasText = true
		}
	}

	if !hasAudio {
		t.Fatal("expected AUDIO modality")
	}
	if !hasText {
		t.Fatal("expected TEXT modality")
	}
}
