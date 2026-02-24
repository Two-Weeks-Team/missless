package store

import (
	"context"
	"fmt"
	"log/slog"
)

// FirestoreStore manages session and persona data in Cloud Firestore.
// Lock ordering: MemoryStore.mu is Level 6 (lowest).
type FirestoreStore struct {
	projectID string
	// TODO: T01 - Add firestore.Client field
}

// NewFirestoreStore creates a new Firestore store.
func NewFirestoreStore(projectID string) *FirestoreStore {
	return &FirestoreStore{projectID: projectID}
}

// SaveSession persists session state to Firestore.
func (fs *FirestoreStore) SaveSession(ctx context.Context, sessionID string, data map[string]any) error {
	slog.Info("firestore_save_session", "session", sessionID)
	// TODO: T01 - Write to sessions/{sessionId}
	return fmt.Errorf("not yet implemented")
}

// LoadSession retrieves session state from Firestore.
func (fs *FirestoreStore) LoadSession(ctx context.Context, sessionID string) (map[string]any, error) {
	slog.Info("firestore_load_session", "session", sessionID)
	// TODO: T01 - Read from sessions/{sessionId}
	return nil, fmt.Errorf("not yet implemented")
}

// Flush commits any pending writes (used during graceful shutdown).
func (fs *FirestoreStore) Flush(ctx context.Context) {
	slog.Info("firestore_flush")
	// TODO: Commit any pending batch writes
}
