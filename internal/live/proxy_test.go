package live

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func setupTestWebSocket(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	connCh := make(chan *websocket.Conn, 1)
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		connCh <- conn
	}))
	t.Cleanup(server.Close)

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { clientConn.Close() })

	serverConn := <-connCh
	t.Cleanup(func() { serverConn.Close() })

	return clientConn, serverConn
}

func TestProxy_BrowserConnect(t *testing.T) {
	clientConn, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	if proxy.browserConn == nil {
		t.Fatal("expected browserConn to be set")
	}
	if proxy.toolHandler == nil {
		t.Fatal("expected toolHandler to be set")
	}

	_ = clientConn
}

func TestProxy_SessionNilDuringSwap(t *testing.T) {
	_, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	session := proxy.getSession()
	if session != nil {
		t.Fatal("expected nil session")
	}

	proxy.SwapSession(nil)
	session = proxy.getSession()
	if session != nil {
		t.Fatal("expected nil session after swap")
	}
}

func TestProxy_GracefulClose(t *testing.T) {
	_, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	ctx, cancel := context.WithCancel(context.Background())
	proxy.Run(ctx)

	time.Sleep(50 * time.Millisecond)

	cancel()
	proxy.Close()
}

func TestProxy_TextEvent(t *testing.T) {
	clientConn, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	proxy.Run(ctx)

	msg := map[string]string{"type": "test", "data": "hello"}
	if err := clientConn.WriteJSON(msg); err != nil {
		t.Fatalf("write json failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	proxy.Close()
}

func TestProxy_SendJSON(t *testing.T) {
	clientConn, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		proxy.Close()
	}()
	proxy.Run(ctx)

	proxy.sendJSON(map[string]string{"type": "transcript", "text": "hello"})

	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	var received map[string]string
	if err := json.Unmarshal(data, &received); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if received["type"] != "transcript" {
		t.Fatalf("expected type=transcript, got %s", received["type"])
	}
	if received["text"] != "hello" {
		t.Fatalf("expected text=hello, got %s", received["text"])
	}
}

func TestProxy_SendBinary(t *testing.T) {
	clientConn, serverConn := setupTestWebSocket(t)

	toolHandler := NewToolHandler()
	proxy := NewProxy(serverConn, nil, toolHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		proxy.Close()
	}()
	proxy.Run(ctx)

	pcmData := []byte{0x01, 0x02, 0x03, 0x04}
	proxy.sendBinary(pcmData)

	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msgType, data, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if msgType != websocket.BinaryMessage {
		t.Fatalf("expected binary message, got %d", msgType)
	}
	if len(data) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(data))
	}
}
