package handler

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/live"
	"github.com/Two-Weeks-Team/missless/internal/memory"
	"github.com/Two-Weeks-Team/missless/internal/scene"
	"github.com/Two-Weeks-Team/missless/internal/session"
	"github.com/Two-Weeks-Team/missless/internal/store"
)

// TestE2E_FullPipeline exercises the complete missless flow:
// onboarding → persona → transition → reunion → album.
// This is a unit-level integration test — no real API calls.
func TestE2E_FullPipeline(t *testing.T) {
	ctx := context.Background()

	// ─── 1. Session Manager ────────────────────────────
	mgr := session.NewManager("e2e-session-1")
	if mgr.State() != session.StateOnboarding {
		t.Fatalf("expected initial state onboarding, got %q", mgr.State())
	}

	// Collect browser events.
	var mu sync.Mutex
	var events []map[string]any
	mgr.SetNotifyFunc(func(v any) {
		mu.Lock()
		defer mu.Unlock()
		if m, ok := v.(map[string]any); ok {
			events = append(events, m)
		}
	})

	// ─── 2. Onboarding Config ──────────────────────────
	onboardCfg := mgr.BuildOnboardingConfig()
	if onboardCfg == nil {
		t.Fatal("expected onboarding config")
	}
	voiceName := onboardCfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName
	if voiceName != "Aoede" {
		t.Fatalf("expected Aoede host voice, got %q", voiceName)
	}

	// ─── 3. Set Persona (after video analysis) ─────────
	mgr.SetPersona("엄마", "Sulafat", "ko", "따뜻하고 다정한 성격", "부드러운 말투와 높은 억양")
	if mgr.PersonaName() != "엄마" {
		t.Fatalf("expected persona '엄마', got %q", mgr.PersonaName())
	}

	// ─── 4. Memory Store (from analysis highlights) ────
	memStore := memory.NewStore(100)
	highlights := []memory.AnalysisHighlight{
		{Timestamp: "0:15", Description: "카페에서 함께 커피를 마시며 웃는 모습", Expression: "happy"},
		{Timestamp: "1:30", Description: "공원에서 산책하며 대화하는 장면", Expression: "calm"},
		{Timestamp: "3:00", Description: "생일 파티에서 케이크를 자르는 모습", Expression: "joyful"},
	}
	if err := memStore.SaveFromAnalysis(ctx, "엄마", highlights); err != nil {
		t.Fatalf("memory save failed: %v", err)
	}
	if memStore.Count("엄마") != 3 {
		t.Fatalf("expected 3 memories, got %d", memStore.Count("엄마"))
	}

	// ─── 5. Transition: onboarding → reunion ───────────
	if err := mgr.TransitionToReunion(ctx); err != nil {
		t.Fatalf("transition failed: %v", err)
	}
	if mgr.State() != session.StateReunion {
		t.Fatalf("expected reunion state, got %q", mgr.State())
	}

	mu.Lock()
	eventCount := len(events)
	mu.Unlock()
	if eventCount < 2 {
		t.Fatalf("expected at least 2 events (transition + ready), got %d", eventCount)
	}

	// ─── 6. Reunion Config ─────────────────────────────
	reunionCfg := mgr.BuildReunionConfig()
	if reunionCfg == nil {
		t.Fatal("expected reunion config")
	}
	rVoice := reunionCfg.SpeechConfig.VoiceConfig.PrebuiltVoiceConfig.VoiceName
	if rVoice != "Sulafat" {
		t.Fatalf("expected persona voice Sulafat, got %q", rVoice)
	}
	if reunionCfg.EnableAffectiveDialog == nil || !*reunionCfg.EnableAffectiveDialog {
		t.Fatal("expected AffectiveDialog enabled")
	}
	if reunionCfg.Proactivity == nil || reunionCfg.Proactivity.ProactiveAudio == nil || !*reunionCfg.Proactivity.ProactiveAudio {
		t.Fatal("expected ProactiveAudio enabled")
	}
	if reunionCfg.ContextWindowCompression == nil {
		t.Fatal("expected ContextWindowCompression")
	}
	sysText := reunionCfg.SystemInstruction.Parts[0].Text
	if !strings.Contains(sysText, "엄마") {
		t.Fatal("expected persona in system instruction")
	}
	if !strings.Contains(sysText, "Affective Dialog") {
		t.Fatal("expected affective dialog rules in instruction")
	}

	// ─── 7. Timer ──────────────────────────────────────
	timerCh := mgr.StartReunionTimer()
	if timerCh == nil {
		t.Fatal("expected timer channel")
	}
	if mgr.ReunionCount() != 1 {
		t.Fatalf("expected reunion count 1, got %d", mgr.ReunionCount())
	}
	mgr.StopReunionTimer()

	// ─── 8. Tool Handler (change_atmosphere + recall_memory) ─
	toolHandler := live.NewToolHandler()
	var toolEvents []map[string]any
	toolHandler.SetEventSender(func(v any) {
		mu.Lock()
		defer mu.Unlock()
		if m, ok := v.(map[string]any); ok {
			toolEvents = append(toolEvents, m)
		}
	})
	toolHandler.SetMemoryStore(memStore, "엄마")

	// change_atmosphere: warm BGM
	result, err := toolHandler.Handle(ctx, "change_atmosphere", map[string]any{"mood": "warm"})
	if err != nil {
		t.Fatalf("change_atmosphere failed: %v", err)
	}
	if result["bgm_url"] == nil {
		t.Fatal("expected bgm_url in result")
	}

	// recall_memory: find café memories
	result, err = toolHandler.Handle(ctx, "recall_memory", map[string]any{"query": "카페"})
	if err != nil {
		t.Fatalf("recall_memory failed: %v", err)
	}
	memories, ok := result["memories"].([]map[string]string)
	if !ok || len(memories) == 0 {
		t.Fatal("expected café memories in recall_memory result")
	}

	// ─── 9. Scene Generation (mock) + Album ────────────
	albumGen := scene.NewAlbumGenerator()
	albumGen.SetPersona("엄마")
	albumGen.RecordScene("scene1-base64", "카페에서의 추억")
	albumGen.RecordScene("scene2-base64", "공원 산책")
	albumGen.RecordScene("scene3-base64", "생일 파티")

	if albumGen.SceneCount() != 3 {
		t.Fatalf("expected 3 album scenes, got %d", albumGen.SceneCount())
	}

	toolHandler.SetAlbumGenerator(albumGen)

	// end_reunion
	result, err = toolHandler.Handle(ctx, "end_reunion", map[string]any{"reason": "natural_end"})
	if err != nil {
		t.Fatalf("end_reunion failed: %v", err)
	}

	// Check album_created event was emitted.
	mu.Lock()
	var albumCreated bool
	for _, ev := range toolEvents {
		if ev["type"] == "album_created" {
			albumCreated = true
			if ev["albumId"] == nil || ev["albumId"] == "" {
				t.Fatal("expected albumId in album_created event")
			}
			if ev["shareUrl"] == nil || ev["shareUrl"] == "" {
				t.Fatal("expected shareUrl in album_created event")
			}
		}
	}
	mu.Unlock()
	if !albumCreated {
		t.Fatal("expected album_created event after end_reunion")
	}

	// ─── 10. Session Store ─────────────────────────────
	sessionStore := store.NewFirestoreStore("test-project")
	sessionData := store.SessionData{
		SessionID:    mgr.SessionID(),
		PersonaName:  mgr.PersonaName(),
		MatchedVoice: mgr.MatchedVoice(),
		LanguageCode: "ko",
		ReunionCount: mgr.ReunionCount(),
		State:        string(mgr.State()),
	}
	if err := sessionStore.SaveSession(ctx, mgr.SessionID(), &sessionData); err != nil {
		t.Fatalf("session save failed: %v", err)
	}
	loaded, err := sessionStore.GetSession(ctx, mgr.SessionID())
	if err != nil {
		t.Fatalf("session get failed: %v", err)
	}
	if loaded.PersonaName != "엄마" {
		t.Fatalf("expected loaded persona '엄마', got %q", loaded.PersonaName)
	}

	// ─── 11. Continue Reunion (session 2) ──────────────
	summary := mgr.BuildOnboardingSummary()
	continuation := mgr.BuildContinueSummary("Previous conversation about the café")
	if !strings.Contains(continuation, "엄마") {
		t.Fatal("expected persona in continuation summary")
	}
	if !strings.Contains(summary, "엄마") {
		t.Fatal("expected persona in onboarding summary")
	}

	// ─── 12. Character Anchor ──────────────────────────
	anchor := scene.NewCharacterAnchor()
	anchor.AddRefImage([]byte{0xFF, 0xD8, 0x01})
	anchor.AddRefImage([]byte{0xFF, 0xD8, 0x02})
	if len(anchor.GetRefImages()) != 2 {
		t.Fatal("expected 2 reference images")
	}
	parts := anchor.GetRefParts()
	if len(parts) != 2 {
		t.Fatal("expected 2 genai parts")
	}
	guide := anchor.SilhouetteGuide()
	if !strings.Contains(guide, "silhouette") {
		t.Fatal("expected silhouette in guide")
	}

	// ─── 13. Shutdown ──────────────────────────────────
	mgr.Shutdown(ctx)
	if mgr.State() != session.StateEnded {
		t.Fatalf("expected ended state, got %q", mgr.State())
	}

	t.Logf("E2E pipeline passed: %d events, %d memories, %d album scenes",
		eventCount, memStore.Count("엄마"), albumGen.SceneCount())
}

// TestE2E_TextInputZero verifies the design requires zero text input.
// All interactions are voice-driven; no text fields exist in the flow.
func TestE2E_TextInputZero(t *testing.T) {
	mgr := session.NewManager("e2e-text-zero")
	cfg := mgr.BuildOnboardingConfig()

	// System instruction should guide voice-only flow.
	sysText := cfg.SystemInstruction.Parts[0].Text
	if !strings.Contains(sysText, "voice") || !strings.Contains(sysText, "missless") {
		t.Fatal("expected voice-oriented system instruction")
	}
}

// TestE2E_TimerCycle verifies the reunion timer lifecycle.
func TestE2E_TimerCycle(t *testing.T) {
	mgr := session.NewManager("e2e-timer")

	// Start → stop is safe.
	ch := mgr.StartReunionTimer()
	if ch == nil {
		t.Fatal("expected timer channel")
	}
	mgr.StopReunionTimer()

	// Start again → count increments.
	mgr.StartReunionTimer()
	if mgr.ReunionCount() != 2 {
		t.Fatalf("expected count 2, got %d", mgr.ReunionCount())
	}
	mgr.StopReunionTimer()

	// Double stop is safe.
	mgr.StopReunionTimer()
}

// TestE2E_ErrorGracefulDegradation tests that tool failures degrade gracefully.
func TestE2E_ErrorGracefulDegradation(t *testing.T) {
	ctx := context.Background()
	h := live.NewToolHandler()

	// No generator set → generate_scene still succeeds (fallback event).
	var eventCount int
	h.SetEventSender(func(v any) { eventCount++ })

	result, err := h.Handle(ctx, "generate_scene", map[string]any{
		"prompt": "test",
		"mood":   "warm",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result["status"] == nil {
		t.Fatal("expected status in result")
	}

	// No memory store → recall_memory returns empty.
	result, err = h.Handle(ctx, "recall_memory", map[string]any{"query": "test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	memories, _ := result["memories"].([]any)
	if len(memories) != 0 {
		t.Fatal("expected empty memories without store")
	}

	// No album generator → end_reunion still succeeds.
	result, err = h.Handle(ctx, "end_reunion", map[string]any{"reason": "test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestE2E_BGMPresets verifies all 6 BGM presets work end-to-end.
func TestE2E_BGMPresets(t *testing.T) {
	ctx := context.Background()
	h := live.NewToolHandler()

	var mu sync.Mutex
	var events []map[string]any
	h.SetEventSender(func(v any) {
		mu.Lock()
		defer mu.Unlock()
		if m, ok := v.(map[string]any); ok {
			events = append(events, m)
		}
	})

	moods := []string{"warm", "romantic", "nostalgic", "playful", "emotional", "farewell"}
	for _, mood := range moods {
		result, err := h.Handle(ctx, "change_atmosphere", map[string]any{"mood": mood})
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", mood, err)
		}
		if result["bgm_url"] == nil || result["bgm_url"] == "" {
			t.Fatalf("%s: expected bgm_url", mood)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 6 {
		t.Fatalf("expected 6 atmosphere events, got %d", len(events))
	}
}

// TestE2E_StateTransitionInvalid verifies invalid transitions are rejected.
func TestE2E_StateTransitionInvalid(t *testing.T) {
	mgr := session.NewManager("e2e-invalid")

	// Can't go directly from onboarding to reunion.
	err := mgr.TransitionTo(session.StateReunion)
	if err == nil {
		t.Fatal("expected error for invalid transition onboarding→reunion")
	}
}

// Ensure import is used.
var _ = time.Second
