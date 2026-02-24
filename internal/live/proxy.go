package live

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Proxy manages the bidirectional WebSocket proxy between browser and Gemini Live API.
// Lock ordering: Proxy.mu is Level 2.
type Proxy struct {
	browserConn *websocket.Conn
	toolHandler *ToolHandler
	mu          sync.Mutex
	wg          sync.WaitGroup
	done        chan struct{}
}

// NewProxy creates a new proxy instance.
func NewProxy(browserConn *websocket.Conn, toolHandler *ToolHandler) *Proxy {
	return &Proxy{
		browserConn: browserConn,
		toolHandler: toolHandler,
		done:        make(chan struct{}),
	}
}

// Run starts the bidirectional proxy forwarding goroutines.
func (p *Proxy) Run(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.wg.Add(2)
	go func() {
		defer p.wg.Done()
		p.forwardBrowserToLive(ctx)
	}()
	go func() {
		defer p.wg.Done()
		p.forwardLiveToBrowser(ctx)
	}()
}

// forwardBrowserToLive reads from browser WebSocket and sends to Live API.
func (p *Proxy) forwardBrowserToLive(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("browser_to_live_panic", "recover", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.done:
			return
		default:
		}

		p.browserConn.SetReadDeadline(time.Now().Add(30 * time.Second))
		_, _, err := p.browserConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Info("browser_disconnected")
			} else {
				slog.Error("browser_read_error", "error", err)
			}
			return
		}

		// TODO: T02 - Forward audio/events to Live API session
	}
}

// forwardLiveToBrowser reads from Live API and sends to browser.
func (p *Proxy) forwardLiveToBrowser(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("live_to_browser_panic", "recover", r)
		}
	}()

	// TODO: T02 - Receive from Live API session and forward to browser
	// Handle: audio response, tool calls, GoAway, session resumption
	<-ctx.Done()
}

// Close terminates the proxy and waits for goroutines.
func (p *Proxy) Close() {
	close(p.done)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("proxy_goroutines_exited")
	case <-time.After(5 * time.Second):
		slog.Warn("proxy_shutdown_timeout")
	}
}
