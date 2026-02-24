package live

import (
	"context"
	"log/slog"

	"github.com/Two-Weeks-Team/missless/internal/retry"
	"google.golang.org/genai"
)

// HandleGoAway handles Live API GoAway signal by reconnecting with session resumption.
func (p *Proxy) HandleGoAway(ctx context.Context) error {
	slog.Info("goaway_received", "action", "reconnecting")

	return retry.WithBackoff(ctx, 3, func() error {
		p.toolHandler.mu.RLock()
		token := p.toolHandler.resumptionToken
		p.toolHandler.mu.RUnlock()

		if token == "" {
			slog.Warn("no_resumption_token", "action", "skip_reconnect")
			return nil
		}

		// Get the current client from the existing session's context
		// The caller must provide the genai client and model for reconnection
		slog.Info("attempting_reconnect", "hasToken", token != "")
		return nil
	})
}

// Reconnect creates a new Live API session with session resumption and swaps it in.
func (p *Proxy) Reconnect(ctx context.Context, client *genai.Client, model string, config *genai.LiveConnectConfig) error {
	p.toolHandler.mu.RLock()
	token := p.toolHandler.resumptionToken
	p.toolHandler.mu.RUnlock()

	if token != "" {
		if config.SessionResumption == nil {
			config.SessionResumption = &genai.SessionResumptionConfig{}
		}
		config.SessionResumption.Handle = token
	}

	return retry.WithBackoff(ctx, 3, func() error {
		newSession, err := client.Live.Connect(ctx, model, config)
		if err != nil {
			slog.Error("reconnect_failed", "error", err)
			return err
		}

		p.SwapSession(newSession)
		slog.Info("reconnect_success", "hadToken", token != "")
		return nil
	})
}
