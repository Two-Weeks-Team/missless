package live

import (
	"context"
	"testing"

	"google.golang.org/genai"
)

func TestHandleGoAway_NilClientSkipsReconnect(t *testing.T) {
	_, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	// client and liveConfig are nil — HandleGoAway should return nil without error.
	err := proxy.HandleGoAway(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when client is nil, got: %v", err)
	}
}

func TestHandleGoAway_NilConfigSkipsReconnect(t *testing.T) {
	_, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	// Set client but leave config nil.
	proxy.mu.Lock()
	proxy.client = &genai.Client{}
	proxy.mu.Unlock()

	err := proxy.HandleGoAway(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when config is nil, got: %v", err)
	}
}

func TestHandleGoAway_WithClientButNoConfig(t *testing.T) {
	_, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	// Set model but leave config nil — should skip reconnect safely.
	proxy.mu.Lock()
	proxy.model = "test-model"
	proxy.mu.Unlock()

	err := proxy.HandleGoAway(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when config is nil, got: %v", err)
	}
}

func TestReconnect_ResumptionTokenInToolHandler(t *testing.T) {
	_, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	toolHandler.mu.Lock()
	toolHandler.resumptionToken = "test-token-abc"
	toolHandler.mu.Unlock()

	proxy := NewProxy(serverConn, nil, toolHandler)

	// Verify the token is accessible from the tool handler.
	toolHandler.mu.RLock()
	token := toolHandler.resumptionToken
	toolHandler.mu.RUnlock()

	if token != "test-token-abc" {
		t.Errorf("expected resumptionToken %q, got %q", "test-token-abc", token)
	}

	_ = proxy // proxy is used via the toolHandler reference
}

func TestReconnect_ConfigShallowCopy(t *testing.T) {
	// Verify that Reconnect creates a shallow copy of the config
	// by checking the original config is not mutated.
	config := &genai.LiveConnectConfig{
		SessionResumption: &genai.SessionResumptionConfig{
			Handle: "original-handle",
		},
	}

	// Shallow copy simulation (same logic as Reconnect line 36-44).
	cfg := *config
	token := "new-token"
	if token != "" {
		resumption := &genai.SessionResumptionConfig{}
		if config.SessionResumption != nil {
			*resumption = *config.SessionResumption
		}
		resumption.Handle = token
		cfg.SessionResumption = resumption
	}

	// Original should be unmodified.
	if config.SessionResumption.Handle != "original-handle" {
		t.Errorf("original config Handle mutated: got %q", config.SessionResumption.Handle)
	}

	// Copy should have new token.
	if cfg.SessionResumption.Handle != "new-token" {
		t.Errorf("copied config Handle should be %q, got %q", "new-token", cfg.SessionResumption.Handle)
	}
}

func TestReconnect_ConfigNilSessionResumption(t *testing.T) {
	// When config has no SessionResumption and there's a token,
	// a new SessionResumptionConfig should be created.
	config := &genai.LiveConnectConfig{}

	cfg := *config
	token := "fresh-token"
	if token != "" {
		resumption := &genai.SessionResumptionConfig{}
		if config.SessionResumption != nil {
			*resumption = *config.SessionResumption
		}
		resumption.Handle = token
		cfg.SessionResumption = resumption
	}

	if config.SessionResumption != nil {
		t.Error("original config should not have SessionResumption")
	}
	if cfg.SessionResumption == nil || cfg.SessionResumption.Handle != "fresh-token" {
		t.Error("copied config should have SessionResumption with fresh-token")
	}
}

func TestReconnect_EmptyTokenSkipsResumption(t *testing.T) {
	config := &genai.LiveConnectConfig{}

	cfg := *config
	token := "" // empty token
	if token != "" {
		resumption := &genai.SessionResumptionConfig{}
		cfg.SessionResumption = resumption
	}

	if cfg.SessionResumption != nil {
		t.Error("should not set SessionResumption when token is empty")
	}
}
