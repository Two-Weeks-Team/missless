package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/auth"
	"github.com/Two-Weeks-Team/missless/internal/config"
	"github.com/Two-Weeks-Team/missless/internal/live"
	"github.com/Two-Weeks-Team/missless/internal/retry"
	"github.com/Two-Weeks-Team/missless/internal/scene"
	"github.com/Two-Weeks-Team/missless/internal/session"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

// LiveModel is the Gemini model used for the Live API connection.
const LiveModel = "gemini-2.5-flash-native-audio-preview-12-2025"

// maxConcurrentWS is the application-level limit for simultaneous WebSocket sessions.
// Each session holds a genai client + Live API session + goroutines, so this must stay low.
const maxConcurrentWS = 20

// activeWSConns tracks the number of active WebSocket connections.
var activeWSConns atomic.Int64

// ActiveWSCount returns the current number of active WebSocket connections (for testing/monitoring).
func ActiveWSCount() int64 {
	return activeWSConns.Load()
}

// newUpgrader creates a WebSocket upgrader with origin checking based on environment.
func newUpgrader(cfg *config.Config) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if !cfg.IsProd() {
				return true
			}
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			// Allow configured domain (e.g. missless.co)
			if strings.HasSuffix(origin, cfg.Domain) {
				return true
			}
			// Allow same-origin (e.g. Cloud Run auto-generated URL)
			if strings.Contains(origin, r.Host) {
				return true
			}
			return false
		},
	}
}

// RegisterWebSocket registers the WebSocket endpoint for browser ↔ Go proxy.
func RegisterWebSocket(mux *http.ServeMux, cfg *config.Config, sessions *auth.SessionStore) {
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, cfg, sessions)
	})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, cfg *config.Config, sessions *auth.SessionStore) {
	// Protection: Origin check (in upgrader) + rate limiter + connection limit.
	// Session auth is NOT required here because the onboarding flow connects
	// the WebSocket before the user completes OAuth login.

	// Application-level connection limit to prevent resource exhaustion.
	if activeWSConns.Load() >= maxConcurrentWS {
		slog.Warn("ws_connection_limit", "active", activeWSConns.Load(), "max", maxConcurrentWS)
		http.Error(w, "Service Busy", http.StatusServiceUnavailable)
		return
	}

	up := newUpgrader(cfg)
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket_upgrade_failed", "error", err)
		return
	}
	defer conn.Close()

	activeWSConns.Add(1)
	defer activeWSConns.Add(-1)

	slog.Info("websocket_connected", "remote", r.RemoteAddr, "active_ws", activeWSConns.Load())

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
	toolHandler.SetGenaiClient(client)

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

	// Send initial greeting trigger so the model speaks first.
	err = liveSession.SendClientContent(genai.LiveClientContentInput{
		Turns: []*genai.Content{
			genai.NewContentFromText("(사용자가 방금 접속했습니다. 따뜻하게 인사해주세요.)", "user"),
		},
	})
	if err != nil {
		slog.Error("initial_greeting_failed", "error", err)
	}

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
