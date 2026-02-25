package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFirestoreStore_SaveSession(t *testing.T) {
	fs := NewFirestoreStore("test-project", nil)
	ctx := context.Background()

	data := &SessionData{
		PersonaName:  "Mom",
		MatchedVoice: "Sulafat",
		LanguageCode: "ko",
		State:        "onboarding",
		OAuthToken:   map[string]string{"access_token": "test-token"},
	}

	err := fs.SaveSession(ctx, "session-1", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify data was stored.
	got, err := fs.GetSession(ctx, "session-1")
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}
	if got.PersonaName != "Mom" {
		t.Fatalf("expected persona 'Mom', got %q", got.PersonaName)
	}
	if got.MatchedVoice != "Sulafat" {
		t.Fatalf("expected voice 'Sulafat', got %q", got.MatchedVoice)
	}
	if got.SessionID != "session-1" {
		t.Fatalf("expected sessionID 'session-1', got %q", got.SessionID)
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if got.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
	}
}

func TestFirestoreStore_GetSession_NotFound(t *testing.T) {
	fs := NewFirestoreStore("test-project", nil)
	ctx := context.Background()

	_, err := fs.GetSession(ctx, "nonexistent-session")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got: %v", err)
	}
}

func TestFirestoreStore_SaveTranscripts(t *testing.T) {
	fs := NewFirestoreStore("test-project", nil)
	ctx := context.Background()

	// Create session first.
	data := &SessionData{
		PersonaName:  "Dad",
		MatchedVoice: "Charon",
		State:        "reunion",
	}
	if err := fs.SaveSession(ctx, "session-2", data); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	// Save transcripts with reunion increment.
	transcripts := []Transcript{
		{Role: "user", Text: "아빠 보고 싶어요", Timestamp: time.Now()},
		{Role: "model", Text: "나도 보고 싶었어", Timestamp: time.Now()},
	}

	err := fs.SaveTranscripts(ctx, "session-2", transcripts, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify transcripts and reunion count.
	got, err := fs.GetSession(ctx, "session-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Transcripts) != 2 {
		t.Fatalf("expected 2 transcripts, got %d", len(got.Transcripts))
	}
	if got.ReunionCount != 1 {
		t.Fatalf("expected reunionCount 1, got %d", got.ReunionCount)
	}

	// Append more transcripts without reunion increment.
	more := []Transcript{
		{Role: "user", Text: "요즘 어떻게 지내세요?", Timestamp: time.Now()},
	}
	err = fs.SaveTranscripts(ctx, "session-2", more, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err = fs.GetSession(ctx, "session-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Transcripts) != 3 {
		t.Fatalf("expected 3 transcripts, got %d", len(got.Transcripts))
	}
	if got.ReunionCount != 1 {
		t.Fatalf("expected reunionCount still 1, got %d", got.ReunionCount)
	}
}

func TestFirestoreStore_SaveTranscripts_NotFound(t *testing.T) {
	fs := NewFirestoreStore("test-project", nil)
	ctx := context.Background()

	err := fs.SaveTranscripts(ctx, "nonexistent", []Transcript{
		{Role: "user", Text: "hello"},
	}, false)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got: %v", err)
	}
}
