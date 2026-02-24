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

	p.mu.Lock()
	client := p.client
	model := p.model
	config := p.liveConfig
	p.mu.Unlock()

	if client == nil || config == nil {
		slog.Warn("reconnect_params_missing", "action", "skip_reconnect")
		return nil
	}

	return p.Reconnect(ctx, client, model, config)
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
