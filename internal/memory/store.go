package memory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Memory represents a stored memory for recall_memory tool.
type Memory struct {
	ID          string    `json:"id" firestore:"id"`
	Topic       string    `json:"topic" firestore:"topic"`
	Description string    `json:"description" firestore:"description"`
	Timestamp   string    `json:"timestamp" firestore:"timestamp"`
	Expression  string    `json:"expression" firestore:"expression"`
	Source      string    `json:"source" firestore:"source"` // "video_analysis" | "user_input"
	CreatedAt   time.Time `json:"createdAt" firestore:"createdAt"`
}

// AnalysisHighlight is the input format from video analysis.
type AnalysisHighlight struct {
	Timestamp   string `json:"timestamp"`
	Description string `json:"description"`
	Expression  string `json:"expression"`
}

// Store manages persona memories.
// When a Firestore client is provided, memories persist to Cloud Firestore.
// Falls back to in-memory storage when client is nil.
// Lock ordering: Store.mu is Level 6 (lowest).
type Store struct {
	mu        sync.RWMutex
	client    *firestore.Client
	memories  map[string][]Memory // in-memory fallback: personaID → memories
	maxPerKey int
}

// NewStore creates a new memory store.
// If client is nil, falls back to in-memory storage.
func NewStore(maxPerKey int, client *firestore.Client) *Store {
	if client != nil {
		slog.Info("memory_store_initialized", "mode", "firestore")
	} else {
		slog.Info("memory_store_initialized", "mode", "in-memory")
	}
	return &Store{
		client:    client,
		memories:  make(map[string][]Memory),
		maxPerKey: maxPerKey,
	}
}

// memoriesCol returns the memories subcollection for a persona.
func (s *Store) memoriesCol(personaID string) *firestore.CollectionRef {
	return s.client.Collection("personas").Doc(personaID).Collection("memories")
}

// SaveFromAnalysis batch-saves highlights from video analysis as memories.
func (s *Store) SaveFromAnalysis(ctx context.Context, personaID string, highlights []AnalysisHighlight) error {
	if len(highlights) == 0 {
		return nil
	}

	now := time.Now()
	mems := make([]Memory, 0, len(highlights))
	for i, h := range highlights {
		mems = append(mems, Memory{
			ID:          fmt.Sprintf("%s-analysis-%d", personaID, i),
			Topic:       extractTopics(h.Description),
			Description: h.Description,
			Timestamp:   h.Timestamp,
			Expression:  h.Expression,
			Source:      "video_analysis",
			CreatedAt:   now,
		})
	}

	if s.client != nil {
		bw := s.client.BulkWriter(ctx)
		col := s.memoriesCol(personaID)
		for _, m := range mems {
			if _, err := bw.Set(col.Doc(m.ID), m); err != nil {
				slog.Error("memory_bulkwriter_set_failed", "persona", personaID, "id", m.ID, "error", err)
			}
		}
		bw.End()
		slog.Info("memory_batch_saved", "persona", personaID, "count", len(mems), "mode", "firestore")
		return nil
	}

	// In-memory fallback.
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memories[personaID] = append(s.memories[personaID], mems...)
	if len(s.memories[personaID]) > s.maxPerKey {
		s.memories[personaID] = s.memories[personaID][len(s.memories[personaID])-s.maxPerKey:]
	}
	slog.Info("memory_batch_saved", "persona", personaID, "count", len(mems), "mode", "in-memory")
	return nil
}

// Save stores a single memory.
func (s *Store) Save(ctx context.Context, personaID string, memory Memory) error {
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = time.Now()
	}

	if s.client != nil {
		docID := memory.ID
		if docID == "" {
			docID = fmt.Sprintf("%s-%d", personaID, time.Now().UnixNano())
		}
		_, err := s.memoriesCol(personaID).Doc(docID).Set(ctx, memory)
		if err != nil {
			slog.Error("memory_save_failed", "persona", personaID, "error", err)
			return err
		}
		slog.Info("memory_saved", "persona", personaID, "topic", memory.Topic, "mode", "firestore")
		return nil
	}

	// In-memory fallback.
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memories[personaID] = append(s.memories[personaID], memory)
	if len(s.memories[personaID]) > s.maxPerKey {
		s.memories[personaID] = s.memories[personaID][len(s.memories[personaID])-s.maxPerKey:]
	}
	slog.Info("memory_saved", "persona", personaID, "topic", memory.Topic, "mode", "in-memory")
	return nil
}

// Search finds memories matching the query keywords.
// Returns at most 10 results sorted by relevance (match count).
func (s *Store) Search(ctx context.Context, personaID, query string) ([]Memory, error) {
	keywords := splitKeywords(query)
	if len(keywords) == 0 {
		return nil, nil
	}

	if s.client != nil {
		return s.searchFirestore(ctx, personaID, keywords)
	}

	// In-memory fallback.
	s.mu.RLock()
	defer s.mu.RUnlock()
	return searchInMemory(s.memories[personaID], keywords), nil
}

// searchFirestore fetches all memories for a persona and filters by keywords.
func (s *Store) searchFirestore(ctx context.Context, personaID string, keywords []string) ([]Memory, error) {
	iter := s.memoriesCol(personaID).Documents(ctx)
	defer iter.Stop()

	var allMems []Memory
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var m Memory
		if err := doc.DataTo(&m); err != nil {
			continue
		}
		allMems = append(allMems, m)
	}

	return searchInMemory(allMems, keywords), nil
}

// searchInMemory filters and ranks memories by keyword match score.
func searchInMemory(mems []Memory, keywords []string) []Memory {
	if len(mems) == 0 {
		return nil
	}

	type scored struct {
		mem   Memory
		score int
	}

	var results []scored
	for _, m := range mems {
		score := matchScore(m, keywords)
		if score > 0 {
			results = append(results, scored{mem: m, score: score})
		}
	}

	// Sort by score descending (insertion sort for small N).
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].score > results[j-1].score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	limit := 10
	if len(results) < limit {
		limit = len(results)
	}

	out := make([]Memory, limit)
	for i := 0; i < limit; i++ {
		out[i] = results[i].mem
	}
	return out
}

// Count returns the number of memories for a persona.
func (s *Store) Count(personaID string) int {
	if s.client != nil {
		// For Firestore, we'd need to count documents.
		// For performance, return from in-memory cache if available.
		s.mu.RLock()
		defer s.mu.RUnlock()
		return len(s.memories[personaID])
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.memories[personaID])
}

// matchScore returns how many keywords match in a memory's topic and description.
func matchScore(m Memory, keywords []string) int {
	score := 0
	lower := strings.ToLower(m.Topic + " " + m.Description)
	for _, kw := range keywords {
		if ContainsKeyword(lower, kw) {
			score++
		}
	}
	return score
}

// ContainsKeyword checks if text contains the keyword (case-insensitive).
func ContainsKeyword(text, keyword string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(keyword))
}

// splitKeywords splits a query into individual keywords.
func splitKeywords(query string) []string {
	parts := strings.Fields(strings.ToLower(query))
	var keywords []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) > 0 {
			keywords = append(keywords, p)
		}
	}
	return keywords
}

// extractTopics creates a topic string from a description by taking key words.
func extractTopics(description string) string {
	return strings.ToLower(description)
}
