package memory

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Memory represents a stored memory for recall_memory tool.
type Memory struct {
	ID          string    `json:"id" firestore:"id"`
	Topic       string    `json:"topic" firestore:"topic"`
	Description string    `json:"description" firestore:"description"`
	Timestamp   string    `json:"timestamp" firestore:"timestamp"`
	Source      string    `json:"source" firestore:"source"` // "video_analysis" | "user_input"
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
}

// Store manages persona memories in Firestore.
// Lock ordering: MemoryStore.mu is Level 6 (lowest).
type Store struct {
	mu        sync.RWMutex
	cache     map[string][]Memory // LRU cache with size limit
	maxCache  int
	projectID string
	// TODO: T16 - Add firestore.Client field
}

// NewStore creates a new memory store.
func NewStore(projectID string, maxCache int) *Store {
	return &Store{
		cache:     make(map[string][]Memory),
		maxCache:  maxCache,
		projectID: projectID,
	}
}

// Search finds memories by topic keywords.
func (s *Store) Search(ctx context.Context, personaID, topic string) ([]Memory, error) {
	slog.Info("memory_search", "persona", personaID, "topic", topic)

	// Check cache first
	s.mu.RLock()
	if cached, ok := s.cache[personaID+":"+topic]; ok {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	// TODO: T16 - Firestore query: personas/{personaId}/memories
	// WHERE keywords array-contains-any [topic keywords]
	// LIMIT 10, ORDER BY relevance DESC

	return nil, fmt.Errorf("not yet implemented")
}

// Save stores a new memory.
func (s *Store) Save(ctx context.Context, personaID string, memory Memory) error {
	slog.Info("memory_save", "persona", personaID, "topic", memory.Topic)

	// TODO: T16 - Write to personas/{personaId}/memories/{memoryId}

	return fmt.Errorf("not yet implemented")
}
