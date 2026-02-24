package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Two-Weeks-Team/missless/internal/auth"
	"github.com/Two-Weeks-Team/missless/internal/config"
	"github.com/Two-Weeks-Team/missless/internal/media"
	"golang.org/x/oauth2"
)

// OAuthDeps holds dependencies for OAuth handlers.
type OAuthDeps struct {
	Config   *oauth2.Config
	Sessions *auth.SessionStore
	YouTube  *media.YouTubeClient
	IsProd   bool
	// ExchangeFunc overrides token exchange for testing. If nil, uses Config.Exchange.
	ExchangeFunc func(ctx context.Context, code string) (*oauth2.Token, error)
}

// RegisterOAuth registers OAuth and video listing endpoints.
func RegisterOAuth(mux *http.ServeMux, cfg *config.Config, sessions *auth.SessionStore) {
	oauthCfg := auth.NewOAuthConfig(cfg.YouTubeClientID, cfg.YouTubeSecret, cfg.OAuthRedirect)
	deps := &OAuthDeps{
		Config:   oauthCfg,
		Sessions: sessions,
		YouTube:  media.NewYouTubeClient(),
		IsProd:   cfg.IsProd(),
	}

	mux.HandleFunc("GET /auth/login", func(w http.ResponseWriter, r *http.Request) {
		handleOAuthLogin(w, r, deps)
	})
	mux.HandleFunc("GET /auth/callback", func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(w, r, deps)
	})
	mux.HandleFunc("GET /api/videos", func(w http.ResponseWriter, r *http.Request) {
		handleListVideos(w, r, deps)
	})
}

func handleOAuthLogin(w http.ResponseWriter, r *http.Request, deps *OAuthDeps) {
	state, err := auth.GenerateState()
	if err != nil {
		slog.Error("oauth_state_error", "error", err)
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	deps.Sessions.StoreState(state)
	auth.SetStateCookie(w, state, deps.IsProd)

	url := deps.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	slog.Info("oauth_login_redirect", "url_len", len(url))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request, deps *OAuthDeps) {
	// Validate state: query param must match cookie and exist in store.
	queryState := r.URL.Query().Get("state")
	cookieState := auth.GetStateFromCookie(r)

	if queryState == "" || queryState != cookieState || !deps.Sessions.ValidateState(queryState) {
		slog.Warn("oauth_invalid_state", "query", queryState)
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for token.
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	var token *oauth2.Token
	var err error
	if deps.ExchangeFunc != nil {
		token, err = deps.ExchangeFunc(r.Context(), code)
	} else {
		token, err = deps.Config.Exchange(r.Context(), code)
	}
	if err != nil {
		slog.Error("oauth_exchange_failed", "error", err)
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	// Store session and set cookie.
	sessionID, err := deps.Sessions.CreateSession(token)
	if err != nil {
		slog.Error("session_create_failed", "error", err)
		http.Error(w, "Session creation failed", http.StatusInternalServerError)
		return
	}

	auth.SetSessionCookie(w, sessionID, deps.IsProd)
	slog.Info("oauth_success")

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func handleListVideos(w http.ResponseWriter, r *http.Request, deps *OAuthDeps) {
	session := deps.Sessions.GetSessionFromRequest(r)
	if session == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	token := session.GetToken()
	videos, err := deps.YouTube.ListUserVideos(r.Context(), token)
	if err != nil {
		slog.Error("youtube_list_failed", "error", err)
		http.Error(w, "Failed to list videos", http.StatusInternalServerError)
		return
	}

	analyzable, needsUpload := media.ClassifyVideos(videos)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"analyzable":  analyzable,
		"needsUpload": needsUpload,
	})
}
