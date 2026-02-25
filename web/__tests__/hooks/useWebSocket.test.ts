import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWebSocket } from '../../hooks/useWebSocket';
import type { ServerMessage } from '../../hooks/useWebSocket';

// ---------------------------------------------------------------------------
// Mock WebSocket
// ---------------------------------------------------------------------------

class MockWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;

  // Instance constants mirroring the static ones (spec-compliant shape)
  readonly CONNECTING = 0;
  readonly OPEN = 1;
  readonly CLOSING = 2;
  readonly CLOSED = 3;

  url: string;
  binaryType: string = 'blob';
  readyState: number = MockWebSocket.CONNECTING;
  onopen: ((ev: Event) => void) | null = null;
  onclose: ((ev: CloseEvent) => void) | null = null;
  onerror: ((ev: Event) => void) | null = null;
  onmessage: ((ev: MessageEvent) => void) | null = null;

  send = vi.fn();
  close = vi.fn();

  constructor(url: string) {
    this.url = url;
    instances.push(this);
  }

  // Helpers to drive the mock from tests
  simulateOpen() {
    this.readyState = MockWebSocket.OPEN;
    this.onopen?.(new Event('open'));
  }
  simulateClose() {
    this.readyState = MockWebSocket.CLOSED;
    this.onclose?.(new CloseEvent('close'));
  }
  simulateError() {
    this.onerror?.(new Event('error'));
  }
  simulateMessage(data: ArrayBuffer | string) {
    this.onmessage?.(new MessageEvent('message', { data }));
  }
}

let instances: MockWebSocket[] = [];

beforeEach(() => {
  instances = [];
  vi.stubGlobal('WebSocket', MockWebSocket);
});

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useWebSocket', () => {
  const TEST_URL = 'ws://localhost:18080/ws';

  // 1. Initial state ----------------------------------------------------------
  it('initial state is disconnected', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    expect(result.current.state).toBe('disconnected');
  });

  // 2. connect() creates WebSocket with correct URL and binaryType ------------
  it('connect() creates WebSocket with correct URL and sets binaryType', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });

    expect(instances).toHaveLength(1);
    expect(instances[0].url).toBe(TEST_URL);
    expect(instances[0].binaryType).toBe('arraybuffer');
  });

  // 3a. State: connecting -> connected (onopen) --------------------------------
  it('transitions to connected on WebSocket open', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });
    expect(result.current.state).toBe('connecting');

    act(() => {
      instances[0].simulateOpen();
    });
    expect(result.current.state).toBe('connected');
  });

  // 3b. State: connecting -> error (onerror) -----------------------------------
  it('transitions to error on WebSocket error', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });

    act(() => {
      instances[0].simulateError();
    });
    expect(result.current.state).toBe('error');
  });

  // 3c. State: connected -> disconnected (onclose) -----------------------------
  it('transitions to disconnected on WebSocket close', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });
    act(() => {
      instances[0].simulateOpen();
    });
    expect(result.current.state).toBe('connected');

    act(() => {
      instances[0].simulateClose();
    });
    expect(result.current.state).toBe('disconnected');
  });

  // 4. Binary message dispatched as { type: 'audio', data } --------------------
  it('dispatches binary ArrayBuffer messages as audio type', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });
    act(() => {
      instances[0].simulateOpen();
    });

    const pcm = new ArrayBuffer(1024);
    act(() => {
      instances[0].simulateMessage(pcm);
    });

    expect(onMessage).toHaveBeenCalledTimes(1);
    expect(onMessage).toHaveBeenCalledWith({ type: 'audio', data: pcm });
  });

  // 5. JSON message is parsed and dispatched -----------------------------------
  it('parses and dispatches JSON text messages', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });
    act(() => {
      instances[0].simulateOpen();
    });

    const payload: ServerMessage = { type: 'session_ready' };
    act(() => {
      instances[0].simulateMessage(JSON.stringify(payload));
    });

    expect(onMessage).toHaveBeenCalledTimes(1);
    expect(onMessage).toHaveBeenCalledWith(payload);
  });

  // 6. send() with ArrayBuffer sends binary directly ---------------------------
  it('send() transmits ArrayBuffer data directly as binary', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });
    act(() => {
      instances[0].simulateOpen();
    });

    const audioData = new ArrayBuffer(512);
    act(() => {
      result.current.send({ type: 'audio', data: audioData });
    });

    expect(instances[0].send).toHaveBeenCalledTimes(1);
    expect(instances[0].send).toHaveBeenCalledWith(audioData);
  });

  // 7. send() with JSON message sends JSON.stringify ---------------------------
  it('send() transmits JSON messages as stringified text', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });
    act(() => {
      instances[0].simulateOpen();
    });

    act(() => {
      result.current.send({ type: 'select_video', videoId: 'abc123' });
    });

    expect(instances[0].send).toHaveBeenCalledTimes(1);
    expect(instances[0].send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'select_video', videoId: 'abc123' })
    );
  });

  // 8. send() does nothing when not connected ----------------------------------
  it('send() is a no-op when WebSocket is not open', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    // No connection at all
    act(() => {
      result.current.send({ type: 'audio_stream_end' });
    });
    // No WebSocket instance was created, so nothing to assert on send.
    // Now connect but keep in CONNECTING state (not OPEN)
    act(() => {
      result.current.connect();
    });
    // readyState is CONNECTING (0), not OPEN (1)
    act(() => {
      result.current.send({ type: 'audio_stream_end' });
    });

    expect(instances[0].send).not.toHaveBeenCalled();
  });

  // 9. disconnect() calls close and nullifies ref ------------------------------
  it('disconnect() closes the WebSocket', () => {
    const onMessage = vi.fn();
    const { result } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });
    act(() => {
      instances[0].simulateOpen();
    });

    act(() => {
      result.current.disconnect();
    });

    expect(instances[0].close).toHaveBeenCalledTimes(1);

    // After disconnect, send should be a no-op (ref is null)
    act(() => {
      result.current.send({ type: 'audio_stream_end' });
    });
    expect(instances[0].send).not.toHaveBeenCalled();
  });

  // 10. Cleanup on unmount calls close -----------------------------------------
  it('closes WebSocket on unmount', () => {
    const onMessage = vi.fn();
    const { result, unmount } = renderHook(() => useWebSocket(TEST_URL, onMessage));

    act(() => {
      result.current.connect();
    });
    act(() => {
      instances[0].simulateOpen();
    });

    unmount();

    expect(instances[0].close).toHaveBeenCalled();
  });
});
