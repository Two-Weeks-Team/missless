package live

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/util"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

// Proxy manages the bidirectional WebSocket proxy between browser and Gemini Live API.
// Lock ordering: Proxy.mu is Level 2.
type Proxy struct {
	browserConn   *websocket.Conn
	liveSession   *genai.Session
	toolHandler   *ToolHandler
	mu            sync.Mutex
	wg            sync.WaitGroup
	done          chan struct{}
	sendToBrowser chan []byte
}

// NewProxy creates a new proxy instance.
func NewProxy(browserConn *websocket.Conn, liveSession *genai.Session, toolHandler *ToolHandler) *Proxy {
	return &Proxy{
		browserConn:   browserConn,
		liveSession:   liveSession,
		toolHandler:   toolHandler,
		done:          make(chan struct{}),
		sendToBrowser: make(chan []byte, 64),
	}
}

// Run starts the bidirectional proxy forwarding goroutines.
func (p *Proxy) Run(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.wg.Add(3)
	util.SafeGo(func() {
		defer p.wg.Done()
		p.forwardBrowserToLive(ctx)
	})
	util.SafeGo(func() {
		defer p.wg.Done()
		p.forwardLiveToBrowser(ctx)
	})
	util.SafeGo(func() {
		defer p.wg.Done()
		p.browserWriter(ctx)
	})
}

// SwapSession replaces the live session (used for onboarding→reunion transition).
func (p *Proxy) SwapSession(newSession *genai.Session) {
	p.mu.Lock()
	defer p.mu.Unlock()
	old := p.liveSession
	p.liveSession = newSession
	if old != nil {
		old.Close()
	}
	slog.Info("session_swapped")
}

// getSession returns the current live session under lock.
func (p *Proxy) getSession() *genai.Session {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.liveSession
}

// forwardBrowserToLive reads from browser WebSocket and sends to Live API.
func (p *Proxy) forwardBrowserToLive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.done:
			return
		default:
		}

		p.browserConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		msgType, data, err := p.browserConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Info("browser_disconnected")
			} else {
				slog.Error("browser_read_error", "error", err)
			}
			return
		}

		session := p.getSession()
		if session == nil {
			slog.Warn("live_session_nil", "action", "skip_forward")
			continue
		}

		if msgType == websocket.BinaryMessage {
			// Binary = PCM audio from browser
			encoded := base64.StdEncoding.EncodeToString(data)
			err = session.SendRealtimeInput(genai.LiveSendRealtimeInputParameters{
				Audio: &genai.Blob{
					MIMEType: "audio/pcm;rate=16000",
					Data:     []byte(encoded),
				},
			})
			if err != nil {
				slog.Error("send_audio_to_live_failed", "error", err)
			}
		} else {
			// Text = JSON control messages from browser
			var msg map[string]any
			if err := json.Unmarshal(data, &msg); err != nil {
				slog.Warn("invalid_browser_json", "error", err)
				continue
			}
			slog.Debug("browser_text_message", "msg", msg)
		}
	}
}

// forwardLiveToBrowser reads from Live API and sends to browser.
func (p *Proxy) forwardLiveToBrowser(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.done:
			return
		default:
		}

		session := p.getSession()
		if session == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		msg, err := session.Receive()
		if err != nil {
			select {
			case <-p.done:
				return
			case <-ctx.Done():
				return
			default:
			}
			slog.Error("live_receive_error", "error", err)
			return
		}

		p.handleLiveMessage(ctx, msg)
	}
}

// handleLiveMessage dispatches a Live API server message.
func (p *Proxy) handleLiveMessage(ctx context.Context, msg *genai.LiveServerMessage) {
	if msg.SetupComplete != nil {
		slog.Info("live_setup_complete")
	}

	if msg.ServerContent != nil {
		p.handleServerContent(msg.ServerContent)
	}

	if msg.ToolCall != nil {
		p.handleToolCall(ctx, msg.ToolCall)
	}

	if msg.GoAway != nil {
		slog.Warn("live_goaway_received", "timeLeft", msg.GoAway.TimeLeft)
		util.SafeGo(func() {
			if err := p.HandleGoAway(ctx); err != nil {
				slog.Error("goaway_reconnect_failed", "error", err)
			}
		})
	}

	if msg.SessionResumptionUpdate != nil {
		if msg.SessionResumptionUpdate.NewHandle != "" {
			p.toolHandler.UpdateResumptionToken(msg.SessionResumptionUpdate.NewHandle)
			slog.Debug("resumption_token_updated")
		}
	}
}

// handleServerContent processes audio and text from the model.
func (p *Proxy) handleServerContent(content *genai.LiveServerContent) {
	if content.ModelTurn != nil {
		for _, part := range content.ModelTurn.Parts {
			if part.InlineData != nil && part.InlineData.MIMEType == "audio/pcm;rate=24000" {
				// Forward PCM audio as binary to browser
				p.sendBinary(part.InlineData.Data)
			}
			if part.Text != "" {
				// Forward transcript as JSON
				p.sendJSON(map[string]any{
					"type": "transcript",
					"role": "model",
					"text": part.Text,
				})
			}
		}
	}
}

// handleToolCall executes a tool and sends the response back to Live API.
func (p *Proxy) handleToolCall(ctx context.Context, tc *genai.LiveServerToolCall) {
	for _, fc := range tc.FunctionCalls {
		slog.Info("tool_call_received", "name", fc.Name, "id", fc.ID)

		result, err := p.toolHandler.Handle(ctx, fc.Name, fc.Args)
		if err != nil {
			slog.Error("tool_call_error", "name", fc.Name, "error", err)
			result = map[string]any{"error": err.Error()}
		}

		session := p.getSession()
		if session == nil {
			slog.Warn("session_nil_during_tool_response")
			return
		}

		err = session.SendToolResponse(genai.LiveSendToolResponseParameters{
			FunctionResponses: []*genai.FunctionResponse{
				{
					Name:     fc.Name,
					ID:       fc.ID,
					Response: result,
				},
			},
		})
		if err != nil {
			slog.Error("tool_response_send_failed", "name", fc.Name, "error", err)
		}
	}
}

// browserWriter writes messages to browser from the send channel.
func (p *Proxy) browserWriter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.done:
			return
		case data := <-p.sendToBrowser:
			p.browserConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			var msgType int
			if json.Valid(data) {
				msgType = websocket.TextMessage
			} else {
				msgType = websocket.BinaryMessage
			}
			if err := p.browserConn.WriteMessage(msgType, data); err != nil {
				slog.Error("browser_write_error", "error", err)
				return
			}
		}
	}
}

// sendJSON sends a JSON message to the browser.
func (p *Proxy) sendJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		slog.Error("json_marshal_error", "error", err)
		return
	}
	select {
	case p.sendToBrowser <- data:
	default:
		slog.Warn("browser_send_buffer_full")
	}
}

// sendBinary sends binary data to the browser.
func (p *Proxy) sendBinary(data []byte) {
	select {
	case p.sendToBrowser <- data:
	default:
		slog.Warn("browser_send_buffer_full")
	}
}

// Close terminates the proxy and waits for goroutines.
func (p *Proxy) Close() {
	close(p.done)

	p.mu.Lock()
	if p.liveSession != nil {
		p.liveSession.Close()
		p.liveSession = nil
	}
	p.mu.Unlock()

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
