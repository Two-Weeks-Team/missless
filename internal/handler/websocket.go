package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/config"
	"github.com/Two-Weeks-Team/missless/internal/live"
	"github.com/Two-Weeks-Team/missless/internal/retry"
	"github.com/Two-Weeks-Team/missless/internal/scene"
	"github.com/Two-Weeks-Team/missless/internal/session"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

// LiveModel is the Gemini model used for the Live API connection.
const LiveModel = "gemini-2.5-flash-native-audio"

// newUpgrader creates a WebSocket upgrader with origin checking based on environment.
func newUpgrader(cfg *config.Config) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if !cfg.IsProd() {
				return true
			}
			origin := r.Header.Get("Origin")
			return origin == "" || strings.HasSuffix(origin, cfg.Domain)
		},
	}
}

// RegisterWebSocket registers the WebSocket endpoint for browser ↔ Go proxy.
func RegisterWebSocket(mux *http.ServeMux, cfg *config.Config) {
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, cfg)
	})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	up := newUpgrader(cfg)
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket_upgrade_failed", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("websocket_connected", "remote", r.RemoteAddr)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Create genai client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			APIVersion: "v1alpha",
		},
	})
	if err != nil {
		slog.Error("genai_client_failed", "error", err)
		conn.WriteJSON(map[string]string{"type": "tool_error", "message": "Failed to initialize AI client"})
		return
	}

	// Create session manager
	mgr := session.NewManager(r.RemoteAddr)

	// Create image generator and tool handler
	sceneGen := scene.NewGenerator(client)
	toolHandler := live.NewToolHandler()
	toolHandler.SetGenerator(sceneGen)

	// Connect to Live API with onboarding config and retry
	liveConfig := mgr.BuildOnboardingConfig()
	var liveSession *genai.Session
	err = retry.WithBackoff(ctx, 3, func() error {
		var connectErr error
		liveSession, connectErr = client.Live.Connect(ctx, LiveModel, liveConfig)
		return connectErr
	})
	if err != nil {
		slog.Error("live_connect_failed", "error", err)
		conn.WriteJSON(map[string]string{"type": "tool_error", "message": "Failed to connect to Live API"})
		return
	}

	// Create and run proxy
	proxy := live.NewProxy(conn, liveSession, toolHandler)
	proxy.SetReconnectParams(client, LiveModel, liveConfig)
	proxy.Run(ctx)

	slog.Info("session_started",
		"remote", r.RemoteAddr,
		"state", string(mgr.State()),
	)

	// Block until proxy goroutines exit (browser disconnect, error, or server shutdown)
	proxy.Wait()

	cancel()
	proxy.Close()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	mgr.Shutdown(shutdownCtx)
	slog.Info("session_ended", "remote", r.RemoteAddr)
}
