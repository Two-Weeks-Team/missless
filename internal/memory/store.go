package memory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
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

// Store manages persona memories (in-memory for hackathon; Firestore-ready interface).
// Lock ordering: Store.mu is Level 6 (lowest).
type Store struct {
	mu        sync.RWMutex
	memories  map[string][]Memory // personaID → memories
	maxPerKey int
}

// NewStore creates a new memory store.
func NewStore(maxPerKey int) *Store {
	return &Store{
		memories:  make(map[string][]Memory),
		maxPerKey: maxPerKey,
	}
}

// SaveFromAnalysis batch-saves highlights from video analysis as memories.
func (s *Store) SaveFromAnalysis(ctx context.Context, personaID string, highlights []AnalysisHighlight) error {
	if len(highlights) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for i, h := range highlights {
		mem := Memory{
			ID:          fmt.Sprintf("%s-analysis-%d", personaID, i),
			Topic:       extractTopics(h.Description),
			Description: h.Description,
			Timestamp:   h.Timestamp,
			Expression:  h.Expression,
			Source:      "video_analysis",
			CreatedAt:   now,
		}
		s.memories[personaID] = append(s.memories[personaID], mem)
	}

	// Trim to max.
	if len(s.memories[personaID]) > s.maxPerKey {
		s.memories[personaID] = s.memories[personaID][len(s.memories[personaID])-s.maxPerKey:]
	}

	slog.Info("memory_batch_saved", "persona", personaID, "count", len(highlights))
	return nil
}

// Save stores a single memory.
func (s *Store) Save(ctx context.Context, personaID string, memory Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = time.Now()
	}

	s.memories[personaID] = append(s.memories[personaID], memory)

	if len(s.memories[personaID]) > s.maxPerKey {
		s.memories[personaID] = s.memories[personaID][len(s.memories[personaID])-s.maxPerKey:]
	}

	slog.Info("memory_saved", "persona", personaID, "topic", memory.Topic)
	return nil
}

// Search finds memories matching the query keywords.
// Returns at most 10 results sorted by relevance (match count).
func (s *Store) Search(ctx context.Context, personaID, query string) ([]Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mems, ok := s.memories[personaID]
	if !ok || len(mems) == 0 {
		return nil, nil
	}

	keywords := splitKeywords(query)
	if len(keywords) == 0 {
		return nil, nil
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

	// Sort by score descending (simple insertion sort for small N).
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

	slog.Info("memory_search", "persona", personaID, "query", query, "found", len(out))
	return out, nil
}

// Count returns the number of memories for a persona.
func (s *Store) Count(personaID string) int {
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
	// Simple: use the description itself as the topic for keyword matching.
	return strings.ToLower(description)
}
