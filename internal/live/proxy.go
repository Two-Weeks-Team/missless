package live

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/util"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

const (
	// browserSendBufSize is the channel buffer size for outbound browser messages.
	browserSendBufSize = 64

	// browserReadTimeout is the read deadline for browser WebSocket messages.
	browserReadTimeout = 60 * time.Second

	// browserWriteTimeout is the write deadline for browser WebSocket messages.
	browserWriteTimeout = 10 * time.Second

	// sessionPollInterval is the polling interval when waiting for a live session.
	sessionPollInterval = 100 * time.Millisecond

	// shutdownTimeout is the max wait time for goroutines to exit during Close.
	shutdownTimeout = 5 * time.Second
)

// Proxy manages the bidirectional WebSocket proxy between browser and Gemini Live API.
// Lock ordering: Proxy.mu is Level 2.
type Proxy struct {
	browserConn   *websocket.Conn
	liveSession   *genai.Session
	toolHandler   *ToolHandler
	client        *genai.Client
	model         string
	liveConfig    *genai.LiveConnectConfig
	mu            sync.Mutex
	wg            sync.WaitGroup
	done          chan struct{}
	shutdownOnce  sync.Once // ensures initiateShutdown runs exactly once
	sendToBrowser chan []byte
	started       bool // guards against duplicate Run calls
	closed        bool // guards against duplicate Close calls

	// Transcription buffers (only accessed from readLive goroutine — no lock needed).
	inputTransBuf  strings.Builder
	outputTransBuf strings.Builder
}

// NewProxy creates a new proxy instance.
func NewProxy(browserConn *websocket.Conn, liveSession *genai.Session, toolHandler *ToolHandler) *Proxy {
	p := &Proxy{
		browserConn:   browserConn,
		liveSession:   liveSession,
		toolHandler:   toolHandler,
		done:          make(chan struct{}),
		sendToBrowser: make(chan []byte, browserSendBufSize),
	}
	// Wire tool handler events to browser via the proxy send channel.
	toolHandler.SetEventSender(p.sendJSON)
	return p
}

// SetReconnectParams stores the client, model, and config needed for GoAway reconnection.
func (p *Proxy) SetReconnectParams(client *genai.Client, model string, config *genai.LiveConnectConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.client = client
	p.model = model
	p.liveConfig = config
}

// initiateShutdown closes the done channel and live session exactly once,
// signaling all goroutines to exit. It does NOT wait for goroutines (safe to
// call from within a goroutine without deadlock).
func (p *Proxy) initiateShutdown() {
	p.shutdownOnce.Do(func() {
		p.mu.Lock()
		sess := p.liveSession
		p.liveSession = nil
		p.mu.Unlock()

		close(p.done)

		if sess != nil {
			sess.Close()
		}
	})
}

// Run starts the bidirectional proxy forwarding goroutines.
func (p *Proxy) Run(ctx context.Context) {
	p.mu.Lock()
	if p.started || p.closed {
		started, closed := p.started, p.closed
		p.mu.Unlock()
		slog.Warn("proxy_run_rejected", "started", started, "closed", closed)
		return
	}
	p.started = true
	p.mu.Unlock()

	p.wg.Add(3)
	util.SafeGo(func() {
		defer p.wg.Done()
		p.forwardBrowserToLive(ctx)
		p.initiateShutdown() // browser disconnect → unblock siblings
	})
	util.SafeGo(func() {
		defer p.wg.Done()
		p.forwardLiveToBrowser(ctx)
		p.initiateShutdown() // live API error → unblock siblings
	})
	util.SafeGo(func() {
		defer p.wg.Done()
		p.browserWriter(ctx)
	})
}

// SwapSession replaces the live session (used for onboarding→reunion transition).
func (p *Proxy) SwapSession(newSession *genai.Session) {
	p.mu.Lock()
	old := p.liveSession
	p.liveSession = newSession
	p.mu.Unlock()

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

		p.browserConn.SetReadDeadline(time.Now().Add(browserReadTimeout))
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
			// Binary = PCM audio from browser; SDK handles base64 encoding internally.
			err = session.SendRealtimeInput(genai.LiveSendRealtimeInputParameters{
				Audio: &genai.Blob{
					MIMEType: "audio/pcm;rate=16000",
					Data:     data,
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
	consecutiveErrors := 0
	const maxConsecutiveErrors = 3
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
			// Session may be nil during SwapSession — wait for new one.
			consecutiveErrors = 0
			time.Sleep(sessionPollInterval)
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
			consecutiveErrors++
			if consecutiveErrors > maxConsecutiveErrors {
				slog.Error("live_receive_fatal", "error", err, "consecutive", consecutiveErrors)
				return
			}
			// May be a transient error from SwapSession closing old session — retry.
			slog.Warn("live_receive_error", "error", err, "consecutive", consecutiveErrors)
			time.Sleep(sessionPollInterval)
			continue
		}

		consecutiveErrors = 0
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
			if part.Text != "" && !part.Thought {
				// Capture non-thinking transcript for tool context (analyze_user).
				// Browser display uses OutputTranscription to avoid duplicates.
				p.toolHandler.AddTranscript("model", part.Text)
			}
		}
	}

	// Forward input transcription (what the user said).
	if content.InputTranscription != nil && content.InputTranscription.Text != "" {
		appendTranscriptionChunk(&p.inputTransBuf, content.InputTranscription.Text)
		accumulated := p.inputTransBuf.String()

		if content.InputTranscription.Finished {
			p.toolHandler.AddTranscript("user", accumulated)
			p.inputTransBuf.Reset()
		}
		p.sendJSON(map[string]any{
			"type":     "transcript",
			"role":     "user",
			"text":     accumulated,
			"finished": content.InputTranscription.Finished,
		})
	}

	// Forward output transcription (what the model said, as text).
	if content.OutputTranscription != nil && content.OutputTranscription.Text != "" {
		appendTranscriptionChunk(&p.outputTransBuf, content.OutputTranscription.Text)
		accumulated := p.outputTransBuf.String()

		if content.OutputTranscription.Finished {
			p.outputTransBuf.Reset()
		}
		p.sendJSON(map[string]any{
			"type":     "transcript",
			"role":     "model",
			"text":     accumulated,
			"finished": content.OutputTranscription.Finished,
		})
	}
}

// appendTranscriptionChunk appends a transcription delta to a buffer with smart spacing.
// Gemini Live API sends word-level deltas that may lack leading/trailing spaces.
func appendTranscriptionChunk(buf *strings.Builder, chunk string) {
	if buf.Len() == 0 || chunk == "" {
		buf.WriteString(chunk)
		return
	}
	s := buf.String()
	lastChar := s[len(s)-1]
	firstChar := chunk[0]
	if lastChar != ' ' && lastChar != '\n' && firstChar != ' ' && firstChar != '\n' {
		buf.WriteByte(' ')
	}
	buf.WriteString(chunk)
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
			continue
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
			p.browserConn.SetWriteDeadline(time.Now().Add(browserWriteTimeout))
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

// Wait blocks until all proxy goroutines have exited.
// Use this in the handler to keep the session alive for its full lifetime.
func (p *Proxy) Wait() {
	p.wg.Wait()
	slog.Info("proxy_goroutines_exited")
}

// Close terminates the proxy, closes the live session, and waits for goroutines
// with a bounded timeout to prevent hanging during shutdown.
func (p *Proxy) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	p.initiateShutdown()

	wgDone := make(chan struct{}, 1)
	util.SafeGo(func() {
		p.wg.Wait()
		close(wgDone)
	})

	select {
	case <-wgDone:
		slog.Info("proxy_goroutines_exited")
	case <-time.After(shutdownTimeout):
		slog.Warn("proxy_shutdown_timeout")
	}
}
