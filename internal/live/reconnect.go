package live

import (
	"context"
	"log/slog"

	"github.com/Two-Weeks-Team/missless/internal/retry"
)

// HandleGoAway handles Live API GoAway signal by reconnecting with session resumption.
func (p *Proxy) HandleGoAway(ctx context.Context) error {
	slog.Info("goaway_received", "action", "reconnecting")

	return retry.WithBackoff(ctx, 3, func() error {
		// TODO: T02 - Reconnect with SessionResumptionConfig
		// 1. Close old session
		// 2. Create new LiveConnectConfig with resumption handle
		// 3. Reconnect to Live API
		// 4. Swap session pointer (under Proxy.mu lock)
		return nil
	})
}
