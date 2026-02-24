package handler

import (
	"log/slog"
	"net/http"

	"github.com/Two-Weeks-Team/missless/internal/config"
)

// RegisterOAuth registers OAuth endpoints.
func RegisterOAuth(mux *http.ServeMux, cfg *config.Config) {
	mux.HandleFunc("GET /auth/login", func(w http.ResponseWriter, r *http.Request) {
		handleOAuthLogin(w, r, cfg)
	})
	mux.HandleFunc("GET /auth/callback", func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(w, r, cfg)
	})
}

func handleOAuthLogin(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// TODO: T07 - Implement Google OAuth 2.0 login flow
	// 1. Generate state token
	// 2. Build OAuth URL with youtube.readonly scope
	// 3. Redirect to Google
	slog.Info("oauth_login_attempt")
	http.Error(w, "OAuth not yet implemented", http.StatusNotImplemented)
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// TODO: T07 - Handle OAuth callback
	// 1. Validate state token
	// 2. Exchange code for tokens
	// 3. Store in session
	// 4. Redirect to app
	slog.Info("oauth_callback")
	http.Error(w, "OAuth callback not yet implemented", http.StatusNotImplemented)
}
