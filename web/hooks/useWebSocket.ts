import { useCallback, useEffect, useRef, useState } from 'react';

export type ServerMessage =
  | { type: 'audio'; data: ArrayBuffer }
  | { type: 'scene_preview'; image: string }
  | { type: 'scene_final'; image: string }
  | { type: 'atmosphere_change'; mood: string; bgm_url: string }
  | { type: 'tool_error'; tool: string; message: string }
  | { type: 'session_transition' }
  | { type: 'session_ready' }
  | { type: 'analysis_progress'; step: string; percent: number; highlight?: string }
  | { type: 'person_detected'; crops: string[] }
  | { type: 'youtube_videos'; videos: unknown[] }
  | { type: 'transcript'; role: string; text: string; finished?: boolean };

export type ClientMessage =
  | { type: 'audio'; data: ArrayBuffer }
  | { type: 'audio_stream_end' }
  | { type: 'select_video'; videoId: string }
  | { type: 'select_person'; personIndex: number }
  | { type: 'upload_video'; data: ArrayBuffer };

type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'error';

export function useWebSocket(url: string, onMessage: (msg: ServerMessage) => void) {
  const wsRef = useRef<WebSocket | null>(null);
  const [state, setState] = useState<ConnectionState>('disconnected');

  const connect = useCallback(() => {
    setState('connecting');
    const ws = new WebSocket(url);
    ws.binaryType = 'arraybuffer';

    ws.onopen = () => setState('connected');
    ws.onclose = () => setState('disconnected');
    ws.onerror = () => setState('error');

    ws.onmessage = (event) => {
      if (event.data instanceof ArrayBuffer) {
        onMessage({ type: 'audio', data: event.data });
      } else {
        const msg = JSON.parse(event.data) as ServerMessage;
        onMessage(msg);
      }
    };

    wsRef.current = ws;
  }, [url, onMessage]);

  const send = useCallback((msg: ClientMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      if ('data' in msg && msg.data instanceof ArrayBuffer) {
        wsRef.current.send(msg.data);
      } else {
        wsRef.current.send(JSON.stringify(msg));
      }
    }
  }, []);

  const disconnect = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
  }, []);

  useEffect(() => {
    return () => {
      wsRef.current?.close();
    };
  }, []);

  return { state, connect, send, disconnect };
}
