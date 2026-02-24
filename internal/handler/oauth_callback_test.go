package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Two-Weeks-Team/missless/internal/auth"
	"golang.org/x/oauth2"
)

func TestOAuth_HandleLogin_RedirectURL(t *testing.T) {
	sessions := auth.NewSessionStore()
	deps := &OAuthDeps{
		Config:   auth.NewOAuthConfig("test-client-id", "test-secret", "http://localhost:18080/auth/callback"),
		Sessions: sessions,
		IsProd:   false,
	}

	req := httptest.NewRequest("GET", "/auth/login", nil)
	rec := httptest.NewRecorder()

	handleOAuthLogin(rec, req, deps)

	resp := rec.Result()
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		t.Fatal("expected Location header")
	}
	if !strings.Contains(location, "accounts.google.com") {
		t.Fatalf("expected Google OAuth URL, got %s", location)
	}
	if !strings.Contains(location, "youtube.readonly") {
		t.Fatalf("expected youtube.readonly scope in URL, got %s", location)
	}
	if !strings.Contains(location, "state=") {
		t.Fatalf("expected state parameter in URL, got %s", location)
	}
}

func TestOAuth_HandleCallback_InvalidState(t *testing.T) {
	sessions := auth.NewSessionStore()
	deps := &OAuthDeps{
		Config:   auth.NewOAuthConfig("test-client-id", "test-secret", "http://localhost:18080/auth/callback"),
		Sessions: sessions,
		IsProd:   false,
	}

	req := httptest.NewRequest("GET", "/auth/callback?state=invalid&code=test", nil)
	rec := httptest.NewRecorder()

	handleOAuthCallback(rec, req, deps)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestOAuth_HandleCallback_ExchangeFail(t *testing.T) {
	sessions := auth.NewSessionStore()
	state := "test-state-abc123"
	sessions.StoreState(state)

	deps := &OAuthDeps{
		Config:   auth.NewOAuthConfig("test-client-id", "test-secret", "http://localhost:18080/auth/callback"),
		Sessions: sessions,
		IsProd:   false,
		ExchangeFunc: func(ctx context.Context, code string) (*oauth2.Token, error) {
			return nil, fmt.Errorf("exchange failed: invalid grant")
		},
	}

	req := httptest.NewRequest("GET", "/auth/callback?state="+state+"&code=test-code", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: state,
	})
	rec := httptest.NewRecorder()

	handleOAuthCallback(rec, req, deps)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
