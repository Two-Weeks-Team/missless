package store

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
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
// Uses in-memory storage for hackathon; designed for Firestore migration.
// Lock ordering: FirestoreStore.mu is Level 6 (lowest).
type FirestoreStore struct {
	mu        sync.RWMutex
	projectID string
	sessions  map[string]*SessionData
}

// NewFirestoreStore creates a new Firestore store.
func NewFirestoreStore(projectID string) *FirestoreStore {
	return &FirestoreStore{
		projectID: projectID,
		sessions:  make(map[string]*SessionData),
	}
}

// SaveSession persists session state.
func (fs *FirestoreStore) SaveSession(ctx context.Context, sessionID string, data *SessionData) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	now := time.Now()
	data.SessionID = sessionID
	data.UpdatedAt = now
	if data.CreatedAt.IsZero() {
		data.CreatedAt = now
	}

	fs.sessions[sessionID] = data
	slog.Info("firestore_save_session", "session", sessionID, "persona", data.PersonaName)
	return nil
}

// GetSession retrieves session state. Returns ErrSessionNotFound if missing.
func (fs *FirestoreStore) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
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

	slog.Info("firestore_save_transcripts",
		"session", sessionID,
		"newTranscripts", len(transcripts),
		"totalTranscripts", len(data.Transcripts),
		"reunionCount", data.ReunionCount,
	)
	return nil
}

// Flush commits any pending writes (used during graceful shutdown).
func (fs *FirestoreStore) Flush(ctx context.Context) {
	slog.Info("firestore_flush")
}
