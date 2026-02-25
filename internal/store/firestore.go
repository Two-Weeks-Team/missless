package store

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
)

// ErrSessionNotFound is returned when a session does not exist.
var ErrSessionNotFound = errors.New("session not found")

// Transcript represents a single conversation turn.
type Transcript struct {
	Role      string    `json:"role" firestore:"role"` // "user" | "model"
	Text      string    `json:"text" firestore:"text"`
	Timestamp time.Time `json:"timestamp" firestore:"timestamp"`
}

// SessionData holds the full session state persisted to Firestore.
type SessionData struct {
	SessionID    string            `json:"sessionId" firestore:"sessionId"`
	PersonaName  string            `json:"personaName" firestore:"personaName"`
	MatchedVoice string            `json:"matchedVoice" firestore:"matchedVoice"`
	LanguageCode string            `json:"languageCode" firestore:"languageCode"`
	OAuthToken   map[string]string `json:"oauthToken" firestore:"oauthToken"`
	Transcripts  []Transcript      `json:"transcripts" firestore:"transcripts"`
	ReunionCount int               `json:"reunionCount" firestore:"reunionCount"`
	State        string            `json:"state" firestore:"state"`
	CreatedAt    time.Time         `json:"createdAt" firestore:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt" firestore:"updatedAt"`
}

// FirestoreStore manages session and persona data.
// When a Firestore client is provided, data is persisted to Cloud Firestore.
// Falls back to in-memory storage when client is nil.
// Lock ordering: FirestoreStore.mu is Level 6 (lowest).
type FirestoreStore struct {
	mu        sync.RWMutex
	projectID string
	client    *firestore.Client
	sessions  map[string]*SessionData // in-memory fallback
}

// NewFirestoreStore creates a new Firestore store.
// If client is nil, falls back to in-memory storage.
func NewFirestoreStore(projectID string, client *firestore.Client) *FirestoreStore {
	if client != nil {
		slog.Info("firestore_store_initialized", "mode", "firestore", "project", projectID)
	} else {
		slog.Info("firestore_store_initialized", "mode", "in-memory", "project", projectID)
	}
	return &FirestoreStore{
		projectID: projectID,
		client:    client,
		sessions:  make(map[string]*SessionData),
	}
}

// sessionsCol returns the sessions collection reference.
func (fs *FirestoreStore) sessionsCol() *firestore.CollectionRef {
	return fs.client.Collection("sessions")
}

// SaveSession persists session state.
func (fs *FirestoreStore) SaveSession(ctx context.Context, sessionID string, data *SessionData) error {
	now := time.Now()
	data.SessionID = sessionID
	data.UpdatedAt = now
	if data.CreatedAt.IsZero() {
		data.CreatedAt = now
	}

	if fs.client != nil {
		_, err := fs.sessionsCol().Doc(sessionID).Set(ctx, data)
		if err != nil {
			slog.Error("firestore_save_session_failed", "session", sessionID, "error", err)
			return err
		}
		slog.Info("firestore_save_session", "session", sessionID, "persona", data.PersonaName, "mode", "firestore")
		return nil
	}

	// In-memory fallback.
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.sessions[sessionID] = data
	slog.Info("firestore_save_session", "session", sessionID, "persona", data.PersonaName, "mode", "in-memory")
	return nil
}

// GetSession retrieves session state. Returns ErrSessionNotFound if missing.
func (fs *FirestoreStore) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	if fs.client != nil {
		doc, err := fs.sessionsCol().Doc(sessionID).Get(ctx)
		if err != nil {
			slog.Warn("firestore_get_session_not_found", "session", sessionID)
			return nil, ErrSessionNotFound
		}
		var data SessionData
		if err := doc.DataTo(&data); err != nil {
			return nil, err
		}
		return &data, nil
	}

	// In-memory fallback.
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	data, ok := fs.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return data, nil
}

// SaveTranscripts appends conversation transcripts and increments reunion count.
func (fs *FirestoreStore) SaveTranscripts(ctx context.Context, sessionID string, transcripts []Transcript, incrementReunion bool) error {
	if fs.client != nil {
		updates := []firestore.Update{
			{Path: "transcripts", Value: firestore.ArrayUnion(toInterfaceSlice(transcripts)...)},
			{Path: "updatedAt", Value: time.Now()},
		}
		if incrementReunion {
			updates = append(updates, firestore.Update{Path: "reunionCount", Value: firestore.Increment(1)})
		}
		_, err := fs.sessionsCol().Doc(sessionID).Update(ctx, updates)
		if err != nil {
			slog.Error("firestore_save_transcripts_failed", "session", sessionID, "error", err)
			return err
		}
		slog.Info("firestore_save_transcripts", "session", sessionID, "new", len(transcripts), "mode", "firestore")
		return nil
	}

	// In-memory fallback.
	fs.mu.Lock()
	defer fs.mu.Unlock()
	data, ok := fs.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}
	data.Transcripts = append(data.Transcripts, transcripts...)
	if incrementReunion {
		data.ReunionCount++
	}
	data.UpdatedAt = time.Now()
	slog.Info("firestore_save_transcripts", "session", sessionID, "new", len(transcripts), "total", len(data.Transcripts), "mode", "in-memory")
	return nil
}

// Flush commits any pending writes (used during graceful shutdown).
func (fs *FirestoreStore) Flush(ctx context.Context) {
	slog.Info("firestore_flush")
}

// Close closes the Firestore client.
func (fs *FirestoreStore) Close() error {
	if fs.client != nil {
		return fs.client.Close()
	}
	return nil
}

// toInterfaceSlice converts a slice of Transcripts to []any for ArrayUnion.
func toInterfaceSlice(transcripts []Transcript) []any {
	result := make([]any, len(transcripts))
	for i, t := range transcripts {
		result[i] = t
	}
	return result
}
