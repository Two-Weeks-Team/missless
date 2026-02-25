package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
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

func TestFirestoreStore_SaveSession_Overwrite(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	ctx := context.Background()

	data1 := &SessionData{PersonaName: "Original", MatchedVoice: "voice1"}
	_ = store.SaveSession(ctx, "s1", data1)

	data2 := &SessionData{PersonaName: "Updated", MatchedVoice: "voice2"}
	_ = store.SaveSession(ctx, "s1", data2)

	got, err := store.GetSession(ctx, "s1")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if got.PersonaName != "Updated" {
		t.Errorf("expected Updated, got %s", got.PersonaName)
	}
	if got.MatchedVoice != "voice2" {
		t.Errorf("expected voice2, got %s", got.MatchedVoice)
	}
}

func TestFirestoreStore_SaveSession_SetsTimestamps(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	ctx := context.Background()

	data := &SessionData{PersonaName: "Test"}
	_ = store.SaveSession(ctx, "s1", data)

	got, _ := store.GetSession(ctx, "s1")
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
	if got.SessionID != "s1" {
		t.Errorf("SessionID should be set to s1, got %s", got.SessionID)
	}
}

func TestFirestoreStore_SaveSession_PreservesCreatedAt(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	ctx := context.Background()

	// First save sets CreatedAt.
	data := &SessionData{PersonaName: "Test"}
	_ = store.SaveSession(ctx, "s1", data)
	got1, _ := store.GetSession(ctx, "s1")
	createdAt := got1.CreatedAt

	// Small delay to ensure timestamps differ.
	time.Sleep(time.Millisecond)

	// Second save should preserve the original CreatedAt.
	data2 := &SessionData{PersonaName: "Test2", CreatedAt: createdAt}
	_ = store.SaveSession(ctx, "s1", data2)
	got2, _ := store.GetSession(ctx, "s1")

	if !got2.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt should be preserved, got %v, want %v", got2.CreatedAt, createdAt)
	}
	if !got2.UpdatedAt.After(createdAt) {
		t.Error("UpdatedAt should be after CreatedAt on second save")
	}
}

func TestFirestoreStore_SaveTranscripts_Multiple(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	ctx := context.Background()

	_ = store.SaveSession(ctx, "s1", &SessionData{PersonaName: "Test"})

	batch1 := []Transcript{{Role: "user", Text: "hello"}}
	_ = store.SaveTranscripts(ctx, "s1", batch1, false)

	batch2 := []Transcript{{Role: "model", Text: "hi"}, {Role: "user", Text: "bye"}}
	_ = store.SaveTranscripts(ctx, "s1", batch2, true)

	got, _ := store.GetSession(ctx, "s1")
	if len(got.Transcripts) != 3 {
		t.Errorf("expected 3 transcripts, got %d", len(got.Transcripts))
	}
	if got.ReunionCount != 1 {
		t.Errorf("expected reunion count 1, got %d", got.ReunionCount)
	}
}

func TestFirestoreStore_SaveTranscripts_UpdatesTimestamp(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	ctx := context.Background()

	_ = store.SaveSession(ctx, "s1", &SessionData{PersonaName: "Test"})
	got1, _ := store.GetSession(ctx, "s1")
	firstUpdate := got1.UpdatedAt

	time.Sleep(time.Millisecond)

	_ = store.SaveTranscripts(ctx, "s1", []Transcript{{Role: "user", Text: "hi"}}, false)
	got2, _ := store.GetSession(ctx, "s1")

	if !got2.UpdatedAt.After(firstUpdate) {
		t.Error("UpdatedAt should advance after SaveTranscripts")
	}
}

func TestFirestoreStore_Close_NilClient(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	err := store.Close()
	if err != nil {
		t.Errorf("Close with nil client should not error, got %v", err)
	}
}

func TestFirestoreStore_Flush(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	// Flush is a no-op but should not panic.
	store.Flush(context.Background())
}

func TestFirestoreStore_ConcurrentAccess(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent saves to different sessions.
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sid := fmt.Sprintf("session-%d", idx)
			_ = store.SaveSession(ctx, sid, &SessionData{
				PersonaName: fmt.Sprintf("persona-%d", idx),
			})
		}(i)
	}

	wg.Wait()

	// Concurrent reads.
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sid := fmt.Sprintf("session-%d", idx)
			_, _ = store.GetSession(ctx, sid)
		}(i)
	}

	wg.Wait()
}

func TestFirestoreStore_ConcurrentMixed(t *testing.T) {
	store := NewFirestoreStore("test", nil)
	ctx := context.Background()

	// Pre-create sessions.
	for i := 0; i < 5; i++ {
		_ = store.SaveSession(ctx, fmt.Sprintf("s%d", i), &SessionData{
			PersonaName: fmt.Sprintf("p%d", i),
		})
	}

	var wg sync.WaitGroup

	// Concurrent save transcripts + get sessions.
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sid := fmt.Sprintf("s%d", idx%5)
			if idx%2 == 0 {
				_ = store.SaveTranscripts(ctx, sid, []Transcript{
					{Role: "user", Text: fmt.Sprintf("msg-%d", idx)},
				}, idx%3 == 0)
			} else {
				_, _ = store.GetSession(ctx, sid)
			}
		}(i)
	}

	wg.Wait()
}

func TestToInterfaceSlice(t *testing.T) {
	transcripts := []Transcript{
		{Role: "user", Text: "hello"},
		{Role: "model", Text: "hi"},
	}
	result := toInterfaceSlice(transcripts)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
	// Verify types.
	for _, r := range result {
		if _, ok := r.(Transcript); !ok {
			t.Errorf("expected Transcript type, got %T", r)
		}
	}
}

func TestToInterfaceSlice_Empty(t *testing.T) {
	result := toInterfaceSlice(nil)
	if len(result) != 0 {
		t.Errorf("expected 0 for nil input, got %d", len(result))
	}
}
