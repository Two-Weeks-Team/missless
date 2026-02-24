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

	// Create tool handler
	toolHandler := live.NewToolHandler()

	// Connect to Live API with onboarding config and retry
	liveConfig := buildOnboardingConfig()
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

// buildOnboardingConfig creates the Live API config for the onboarding phase.
func buildOnboardingConfig() *genai.LiveConnectConfig {
	return &genai.LiveConnectConfig{
		ResponseModalities: []genai.Modality{genai.ModalityAudio, genai.ModalityText},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: "Aoede",
				},
			},
		},
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(`You are a warm, empathetic AI guide for missless - a virtual reunion experience.
You help users reconnect with people they miss through AI-powered conversations.

During onboarding:
1. Greet the user warmly in Korean
2. Ask who they'd like to reconnect with
3. Ask for a YouTube video of that person to analyze their characteristics
4. Guide the user through the setup process

Be gentle, understanding, and supportive. This is an emotional experience.
Speak naturally in Korean unless the user prefers another language.`),
			},
		},
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{
						Name:        "generate_scene",
						Description: "Generate a background scene image based on the conversation mood and context",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"prompt":     {Type: genai.TypeString, Description: "Scene description prompt"},
								"mood":       {Type: genai.TypeString, Description: "Emotional mood (warm, nostalgic, joyful, etc.)"},
								"characters": {Type: genai.TypeString, Description: "Character descriptions for the scene"},
							},
							Required: []string{"prompt", "mood"},
						},
					},
					{
						Name:        "change_atmosphere",
						Description: "Change the background music to match the conversation mood",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"mood": {Type: genai.TypeString, Description: "Target mood for BGM (warm, nostalgic, joyful, bittersweet)"},
							},
							Required: []string{"mood"},
						},
					},
					{
						Name:        "end_reunion",
						Description: "End the reunion session and generate a memory album",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"reason": {Type: genai.TypeString, Description: "Reason for ending (user_request, natural_end, timeout)"},
							},
							Required: []string{"reason"},
						},
					},
				},
			},
		},
		SessionResumption: &genai.SessionResumptionConfig{
			Transparent: true,
		},
	}
}
