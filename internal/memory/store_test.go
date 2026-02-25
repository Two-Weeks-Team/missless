package memory

import (
	"context"
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
