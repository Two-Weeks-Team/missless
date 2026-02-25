package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestMemoryStore_SaveFromAnalysis(t *testing.T) {
	store := NewStore(100, nil)
	ctx := context.Background()

	highlights := []AnalysisHighlight{
		{Timestamp: "0:15", Description: "카페에서 함께 커피를 마시며 웃는 모습", Expression: "happy"},
		{Timestamp: "1:30", Description: "공원에서 산책하며 대화하는 장면", Expression: "calm"},
		{Timestamp: "3:00", Description: "생일 파티에서 케이크를 자르는 모습", Expression: "joyful"},
	}

	err := store.SaveFromAnalysis(ctx, "persona-1", highlights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.Count("persona-1") != 3 {
		t.Fatalf("expected 3 memories, got %d", store.Count("persona-1"))
	}
}

func TestMemoryStore_SaveFromAnalysis_Empty(t *testing.T) {
	store := NewStore(100, nil)
	ctx := context.Background()

	err := store.SaveFromAnalysis(ctx, "persona-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Count("persona-1") != 0 {
		t.Fatalf("expected 0 memories, got %d", store.Count("persona-1"))
	}
}

func TestMemoryStore_Search_Found(t *testing.T) {
	store := NewStore(100, nil)
	ctx := context.Background()

	highlights := []AnalysisHighlight{
		{Timestamp: "0:15", Description: "카페에서 함께 커피를 마시며 웃는 모습", Expression: "happy"},
		{Timestamp: "1:30", Description: "공원에서 산책하며 대화하는 장면", Expression: "calm"},
		{Timestamp: "3:00", Description: "카페에서 생일 파티를 하는 모습", Expression: "joyful"},
	}
	_ = store.SaveFromAnalysis(ctx, "persona-1", highlights)

	results, err := store.Search(ctx, "persona-1", "카페")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for '카페', got %d", len(results))
	}
}

func TestMemoryStore_Search_NotFound(t *testing.T) {
	store := NewStore(100, nil)
	ctx := context.Background()

	highlights := []AnalysisHighlight{
		{Timestamp: "0:15", Description: "카페에서 함께 커피를 마시며 웃는 모습", Expression: "happy"},
	}
	_ = store.SaveFromAnalysis(ctx, "persona-1", highlights)

	results, err := store.Search(ctx, "persona-1", "학교")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for '학교', got %d", len(results))
	}
}

func TestMemoryStore_Search_EmptyStore(t *testing.T) {
	store := NewStore(100, nil)
	ctx := context.Background()

	results, err := store.Search(ctx, "nonexistent", "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Fatalf("expected nil results, got %v", results)
	}
}

func TestContainsKeyword(t *testing.T) {
	tests := []struct {
		text    string
		keyword string
		want    bool
	}{
		{"카페에서 커피를 마시는 모습", "카페", true},
		{"공원에서 산책하는 모습", "카페", false},
		{"Hello World", "world", true}, // case insensitive
		{"Hello World", "WORLD", true}, // case insensitive
		{"", "test", false},
		{"some text", "", true},
	}

	for _, tc := range tests {
		got := ContainsKeyword(tc.text, tc.keyword)
		if got != tc.want {
			t.Errorf("ContainsKeyword(%q, %q) = %v, want %v", tc.text, tc.keyword, got, tc.want)
		}
	}
}

func TestMemoryStore_Save_Single(t *testing.T) {
	store := NewStore(100, nil)
	ctx := context.Background()

	err := store.Save(ctx, "persona-1", Memory{
		ID:          "mem-1",
		Topic:       "여행",
		Description: "제주도 여행에서 찍은 사진",
		Timestamp:   "2:00",
		Source:      "user_input",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.Count("persona-1") != 1 {
		t.Fatalf("expected 1 memory, got %d", store.Count("persona-1"))
	}

	results, _ := store.Search(ctx, "persona-1", "제주도")
	if len(results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(results))
	}
}

func TestMemoryStore_MaxLimit(t *testing.T) {
	store := NewStore(5, nil)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		_ = store.Save(ctx, "persona-1", Memory{
			ID:          "mem",
			Topic:       "topic",
			Description: "desc",
			Source:      "user_input",
		})
	}

	if store.Count("persona-1") != 5 {
		t.Fatalf("expected max 5 memories, got %d", store.Count("persona-1"))
	}
}

func TestMemoryStore_Search_MultiKeyword(t *testing.T) {
	store := NewStore(100, nil)
	ctx := context.Background()

	highlights := []AnalysisHighlight{
		{Timestamp: "0:15", Description: "카페에서 커피를 마시며 웃는 모습", Expression: "happy"},
		{Timestamp: "1:30", Description: "카페에서 커피와 케이크를 먹는 장면", Expression: "calm"},
		{Timestamp: "3:00", Description: "공원에서 케이크를 먹는 모습", Expression: "joyful"},
	}
	_ = store.SaveFromAnalysis(ctx, "persona-1", highlights)

	// "카페 케이크" should rank the second highlight highest (matches both).
	results, err := store.Search(ctx, "persona-1", "카페 케이크")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// First result should have score 2 (matches both keywords).
	if !ContainsKeyword(results[0].Description, "카페") || !ContainsKeyword(results[0].Description, "케이크") {
		t.Fatalf("expected first result to match both keywords, got: %s", results[0].Description)
	}
}

// --- Additional tests for improved coverage ---

func TestMatchScore(t *testing.T) {
	tests := []struct {
		name     string
		memory   Memory
		keywords []string
		want     int
	}{
		{
			"exact match in topic",
			Memory{Topic: "coffee shop", Description: "visited a place"},
			[]string{"coffee"},
			1,
		},
		{
			"match in description only",
			Memory{Topic: "travel", Description: "paris cafe"},
			[]string{"cafe"},
			1,
		},
		{
			"match in both topic and description",
			Memory{Topic: "cafe visit", Description: "best cafe ever"},
			[]string{"cafe"},
			1, // matchScore concatenates topic+description, single Contains check per keyword
		},
		{
			"no match",
			Memory{Topic: "beach", Description: "waves"},
			[]string{"mountain"},
			0,
		},
		{
			"empty keywords",
			Memory{Topic: "anything", Description: "here"},
			[]string{},
			0,
		},
		{
			"multiple keywords both match",
			Memory{Topic: "beach sunset", Description: "beautiful waves"},
			[]string{"beach", "waves"},
			2,
		},
		{
			"multiple keywords partial match",
			Memory{Topic: "beach sunset", Description: "beautiful sky"},
			[]string{"beach", "waves"},
			1,
		},
		{
			"case insensitive matching",
			Memory{Topic: "COFFEE Shop", Description: "Best LATTE"},
			[]string{"coffee", "latte"},
			2,
		},
		{
			"empty topic and description",
			Memory{Topic: "", Description: ""},
			[]string{"anything"},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchScore(tt.memory, tt.keywords)
			if got != tt.want {
				t.Errorf("matchScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSplitKeywords(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{"two words", "hello world", 2},
		{"single word", "single", 1},
		{"empty string", "", 0},
		{"extra spaces", "  spaces  between  ", 2},
		{"uppercase", "UPPER CASE", 2},
		{"mixed case", "Hello World", 2},
		{"tabs and spaces", "\thello\t world\t", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitKeywords(tt.query)
			if len(got) != tt.want {
				t.Errorf("splitKeywords(%q) returned %d keywords %v, want %d", tt.query, len(got), got, tt.want)
			}
			// Verify all keywords are lowercase
			for _, kw := range got {
				if kw != strings.ToLower(kw) {
					t.Errorf("keyword %q is not lowercase", kw)
				}
			}
		})
	}
}

func TestExtractTopics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"mixed case", "Hello World", "hello world"},
		{"empty string", "", ""},
		{"all uppercase", "UPPERCASE", "uppercase"},
		{"all lowercase", "already lowercase", "already lowercase"},
		{"with numbers", "Event 2024", "event 2024"},
		{"unicode korean", "카페에서 커피", "카페에서 커피"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTopics(tt.input)
			if got != tt.want {
				t.Errorf("extractTopics(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSearchInMemory(t *testing.T) {
	t.Run("basic search with ranking", func(t *testing.T) {
		mems := []Memory{
			{Topic: "coffee", Description: "morning coffee at cafe"},
			{Topic: "beach", Description: "sunset at the beach"},
			{Topic: "coffee beach", Description: "cafe by the sea"},
		}

		results := searchInMemory(mems, []string{"coffee"})
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("empty keywords returns nil", func(t *testing.T) {
		mems := []Memory{
			{Topic: "anything", Description: "here"},
		}
		results := searchInMemory(mems, []string{})
		// With empty keywords, matchScore returns 0, so no results pass the filter.
		// But searchInMemory checks len(mems)==0 first. Let's verify the actual behavior.
		if len(results) != 0 {
			t.Errorf("expected 0 results for empty keywords, got %d", len(results))
		}
	})

	t.Run("nil memories returns nil", func(t *testing.T) {
		results := searchInMemory(nil, []string{"test"})
		if results != nil {
			t.Errorf("expected nil for nil memories, got %v", results)
		}
	})

	t.Run("empty memories slice returns nil", func(t *testing.T) {
		results := searchInMemory([]Memory{}, []string{"test"})
		if results != nil {
			t.Errorf("expected nil for empty memories, got %v", results)
		}
	})

	t.Run("results capped at 10", func(t *testing.T) {
		mems := make([]Memory, 15)
		for i := range mems {
			mems[i] = Memory{
				Topic:       "common keyword",
				Description: fmt.Sprintf("description %d with keyword", i),
			}
		}
		results := searchInMemory(mems, []string{"keyword"})
		if len(results) != 10 {
			t.Errorf("expected max 10 results, got %d", len(results))
		}
	})

	t.Run("sorting by score descending", func(t *testing.T) {
		mems := []Memory{
			{Topic: "alpha", Description: "no match here"},       // score 0 for "beta gamma"
			{Topic: "beta item", Description: "gamma included"},  // score 2
			{Topic: "gamma only", Description: "nothing else"},   // score 1
			{Topic: "beta gamma", Description: "beta and gamma"}, // score 2
		}
		results := searchInMemory(mems, []string{"beta", "gamma"})
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		// First two results should each have score 2
		score0 := matchScore(results[0], []string{"beta", "gamma"})
		score1 := matchScore(results[1], []string{"beta", "gamma"})
		score2 := matchScore(results[2], []string{"beta", "gamma"})
		if score0 < score1 {
			t.Errorf("results not sorted: first score %d < second score %d", score0, score1)
		}
		if score1 < score2 {
			t.Errorf("results not sorted: second score %d < third score %d", score1, score2)
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		mems := []Memory{
			{Topic: "alpha", Description: "bravo"},
			{Topic: "charlie", Description: "delta"},
		}
		results := searchInMemory(mems, []string{"zzzz"})
		if len(results) != 0 {
			t.Errorf("expected 0 results for non-matching keyword, got %d", len(results))
		}
	})
}

func TestMemoryStore_SaveFromAnalysis_MaxLimit(t *testing.T) {
	s := NewStore(3, nil) // max 3 per key

	highlights := make([]AnalysisHighlight, 5)
	for i := range highlights {
		highlights[i] = AnalysisHighlight{
			Timestamp:   fmt.Sprintf("0:%02d", i),
			Description: fmt.Sprintf("highlight %d", i),
			Expression:  "neutral",
		}
	}

	err := s.SaveFromAnalysis(context.Background(), "persona1", highlights)
	if err != nil {
		t.Fatalf("SaveFromAnalysis failed: %v", err)
	}

	// Should be capped at maxPerKey=3
	if got := s.Count("persona1"); got != 3 {
		t.Errorf("expected count 3 (capped), got %d", got)
	}
}

func TestMemoryStore_Save_MaxLimit(t *testing.T) {
	s := NewStore(3, nil)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		err := s.Save(ctx, "persona1", Memory{
			ID:          fmt.Sprintf("m-%d", i),
			Topic:       fmt.Sprintf("topic %d", i),
			Description: fmt.Sprintf("desc %d", i),
		})
		if err != nil {
			t.Fatalf("Save failed at iteration %d: %v", i, err)
		}
	}

	// Should be capped at 3
	if got := s.Count("persona1"); got != 3 {
		t.Errorf("expected 3 (capped), got %d", got)
	}

	// The most recent 3 should remain (items 2, 3, 4)
	results, err := s.Search(ctx, "persona1", "topic")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	for _, r := range results {
		if r.ID == "m-0" || r.ID == "m-1" {
			t.Errorf("old memory %s should have been evicted", r.ID)
		}
	}
}

func TestMemoryStore_Count_EmptyKey(t *testing.T) {
	s := NewStore(10, nil)
	if got := s.Count("nonexistent"); got != 0 {
		t.Errorf("expected 0 for nonexistent key, got %d", got)
	}
}

func TestMemoryStore_Save_SetsCreatedAt(t *testing.T) {
	s := NewStore(10, nil)
	ctx := context.Background()

	mem := Memory{
		ID:          "test-1",
		Topic:       "test topic",
		Description: "test desc",
	}

	// CreatedAt is zero value
	if !mem.CreatedAt.IsZero() {
		t.Fatal("expected zero CreatedAt before Save")
	}

	err := s.Save(ctx, "persona1", mem)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// After save, the stored memory should have a non-zero CreatedAt
	results, _ := s.Search(ctx, "persona1", "test")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt after Save")
	}
}

func TestMemoryStore_Search_EmptyQuery(t *testing.T) {
	s := NewStore(10, nil)
	ctx := context.Background()

	_ = s.Save(ctx, "p1", Memory{ID: "1", Topic: "hello", Description: "world"})

	results, err := s.Search(ctx, "p1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty query, got %v", results)
	}
}

func TestMemoryStore_Search_WhitespaceOnlyQuery(t *testing.T) {
	s := NewStore(10, nil)
	ctx := context.Background()

	_ = s.Save(ctx, "p1", Memory{ID: "1", Topic: "hello", Description: "world"})

	results, err := s.Search(ctx, "p1", "   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for whitespace-only query, got %v", results)
	}
}

func TestMemoryStore_MultiplePersonas(t *testing.T) {
	s := NewStore(10, nil)
	ctx := context.Background()

	_ = s.Save(ctx, "persona-a", Memory{ID: "a1", Topic: "alpha", Description: "first"})
	_ = s.Save(ctx, "persona-b", Memory{ID: "b1", Topic: "beta", Description: "second"})

	resultsA, _ := s.Search(ctx, "persona-a", "alpha")
	resultsB, _ := s.Search(ctx, "persona-b", "alpha")

	if len(resultsA) != 1 {
		t.Errorf("expected 1 result for persona-a, got %d", len(resultsA))
	}
	if len(resultsB) != 0 {
		t.Errorf("expected 0 results for persona-b searching 'alpha', got %d", len(resultsB))
	}

	if s.Count("persona-a") != 1 || s.Count("persona-b") != 1 {
		t.Errorf("expected 1 memory each, got a=%d b=%d", s.Count("persona-a"), s.Count("persona-b"))
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	s := NewStore(100, nil)
	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent saves
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = s.Save(ctx, "shared", Memory{
				ID:          fmt.Sprintf("m-%d", idx),
				Topic:       "topic",
				Description: fmt.Sprintf("desc %d", idx),
			})
		}(i)
	}

	// Concurrent searches while saves are happening
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.Search(ctx, "shared", "topic")
		}()
	}

	wg.Wait()

	if got := s.Count("shared"); got != 20 {
		t.Errorf("expected 20 memories, got %d", got)
	}
}

func TestMemoryStore_ConcurrentSaveFromAnalysis(t *testing.T) {
	s := NewStore(200, nil)
	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent SaveFromAnalysis calls for different personas
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			highlights := []AnalysisHighlight{
				{Timestamp: "0:01", Description: fmt.Sprintf("highlight A for persona %d", idx), Expression: "happy"},
				{Timestamp: "0:02", Description: fmt.Sprintf("highlight B for persona %d", idx), Expression: "calm"},
			}
			_ = s.SaveFromAnalysis(ctx, fmt.Sprintf("persona-%d", idx), highlights)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		if got := s.Count(fmt.Sprintf("persona-%d", i)); got != 2 {
			t.Errorf("persona-%d: expected 2 memories, got %d", i, got)
		}
	}
}

func TestMemoryStore_SaveFromAnalysis_IDFormat(t *testing.T) {
	s := NewStore(100, nil)
	ctx := context.Background()

	highlights := []AnalysisHighlight{
		{Timestamp: "0:10", Description: "first highlight", Expression: "happy"},
		{Timestamp: "0:20", Description: "second highlight", Expression: "calm"},
	}

	_ = s.SaveFromAnalysis(ctx, "test-persona", highlights)

	results, _ := s.Search(ctx, "test-persona", "highlight")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify IDs follow the expected pattern
	for _, r := range results {
		if !strings.HasPrefix(r.ID, "test-persona-analysis-") {
			t.Errorf("unexpected ID format: %s, expected prefix 'test-persona-analysis-'", r.ID)
		}
	}

	// Verify source is set correctly
	for _, r := range results {
		if r.Source != "video_analysis" {
			t.Errorf("expected source 'video_analysis', got %q", r.Source)
		}
	}
}

func TestMemoryStore_SaveFromAnalysis_FieldMapping(t *testing.T) {
	s := NewStore(100, nil)
	ctx := context.Background()

	highlights := []AnalysisHighlight{
		{Timestamp: "1:30", Description: "Beautiful sunset view", Expression: "amazed"},
	}

	_ = s.SaveFromAnalysis(ctx, "persona-x", highlights)

	results, _ := s.Search(ctx, "persona-x", "sunset")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Timestamp != "1:30" {
		t.Errorf("expected timestamp '1:30', got %q", r.Timestamp)
	}
	if r.Description != "Beautiful sunset view" {
		t.Errorf("expected description 'Beautiful sunset view', got %q", r.Description)
	}
	if r.Expression != "amazed" {
		t.Errorf("expected expression 'amazed', got %q", r.Expression)
	}
	// Topic should be lowercased description
	if r.Topic != "beautiful sunset view" {
		t.Errorf("expected topic 'beautiful sunset view', got %q", r.Topic)
	}
}
