package handler

import (
	"log/slog"
	"net/http"

	"github.com/Two-Weeks-Team/missless/internal/config"
)

// RegisterWebSocket registers the WebSocket endpoint for browser ↔ Go proxy.
func RegisterWebSocket(mux *http.ServeMux, cfg *config.Config) {
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, cfg)
	})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// TODO: T02 - Implement WebSocket upgrade + Live API proxy
	// 1. Upgrade HTTP → WebSocket (gorilla/websocket)
	// 2. Create SessionManager
	// 3. Start Live API proxy (Onboarding session with Aoede voice)
	// 4. Handle bidirectional audio + events
	slog.Info("websocket_connect", "remote", r.RemoteAddr)
	http.Error(w, "WebSocket not yet implemented", http.StatusNotImplemented)
}
