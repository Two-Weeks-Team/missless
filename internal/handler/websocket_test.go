package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Two-Weeks-Team/missless/internal/auth"
	"github.com/Two-Weeks-Team/missless/internal/config"
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

func TestWebSocket_AuthRequired_Prod(t *testing.T) {
	sessions := auth.NewSessionStore()
	cfg := &config.Config{
		GeminiAPIKey: "test-key",
		ProjectID:    "test-project",
		Environment:  "production",
		Domain:       "missless.co",
	}

	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()

	handleWebSocket(rec, req, cfg, sessions)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated prod request, got %d", rec.Code)
	}
}

func TestWebSocket_AuthSkipped_Dev(t *testing.T) {
	sessions := auth.NewSessionStore()
	cfg := &config.Config{
		GeminiAPIKey: "test-key",
		ProjectID:    "test-project",
		Environment:  "development",
		Domain:       "localhost",
	}

	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()

	// In dev mode, auth is not required. The handler will proceed past auth
	// and fail at WebSocket upgrade (since this isn't a real WS request),
	// but it should NOT return 401.
	handleWebSocket(rec, req, cfg, sessions)

	if rec.Code == http.StatusUnauthorized {
		t.Fatal("dev mode should not require authentication")
	}
}

func TestWebSocket_ConnectionLimit(t *testing.T) {
	// Verify that ActiveWSCount starts at 0.
	if ActiveWSCount() != 0 {
		t.Fatalf("expected 0 active connections, got %d", ActiveWSCount())
	}

	// Simulate reaching the limit by manually setting the counter.
	activeWSConns.Store(maxConcurrentWS)
	defer activeWSConns.Store(0)

	sessions := auth.NewSessionStore()
	cfg := &config.Config{
		GeminiAPIKey: "test-key",
		ProjectID:    "test-project",
		Environment:  "development",
		Domain:       "localhost",
	}

	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()

	handleWebSocket(rec, req, cfg, sessions)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 at connection limit, got %d", rec.Code)
	}
}
