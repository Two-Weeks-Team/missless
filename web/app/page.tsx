'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useWebSocket, ServerMessage } from '../hooks/useWebSocket';
import { useAudio } from '../hooks/useAudio';
import SceneDisplay from '../components/SceneDisplay';
import SessionTransition from '../components/SessionTransition';

type TransitionPhase = 'idle' | 'transitioning' | 'ready';

const CONNECTION_COLORS: Record<string, string> = {
  connected: '#4ade80',
  connecting: '#fbbf24',
  disconnected: '#ef4444',
  error: '#ef4444',
};

export default function Home() {
  const [started, setStarted] = useState(false);
  const [previewSrc, setPreviewSrc] = useState<string | null>(null);
  const [finalSrc, setFinalSrc] = useState<string | null>(null);
  const [transition, setTransition] = useState<TransitionPhase>('idle');
  const [transcript, setTranscript] = useState<string>('');
  const readyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const { initAudioContext, playPCM, cleanup: cleanupAudio } = useAudio();

  const handleMessage = useCallback((msg: ServerMessage) => {
    switch (msg.type) {
      case 'audio':
        playPCM(msg.data);
        break;
      case 'scene_preview':
        setPreviewSrc(msg.image);
        break;
      case 'scene_final':
        setFinalSrc(msg.image);
        break;
      case 'session_transition':
        setTransition('transitioning');
        break;
      case 'session_ready':
        setTransition('ready');
        break;
      case 'transcript':
        setTranscript(msg.text);
        break;
    }
  }, [playPCM]);

  // Clean up transition timer on unmount or phase change
  useEffect(() => {
    if (transition === 'ready') {
      readyTimerRef.current = setTimeout(() => setTransition('idle'), 1600);
    }
    return () => {
      if (readyTimerRef.current) {
        clearTimeout(readyTimerRef.current);
        readyTimerRef.current = null;
      }
    };
  }, [transition]);

  const wsUrl = typeof window !== 'undefined'
    ? `ws://${window.location.hostname}:${window.location.port || '18080'}/ws`
    : '';

  const { state, connect, disconnect } = useWebSocket(wsUrl, handleMessage);

  const handleStart = () => {
    initAudioContext();
    connect();
    setStarted(true);
  };

  const handleStop = () => {
    disconnect();
    cleanupAudio();
    setStarted(false);
    setPreviewSrc(null);
    setFinalSrc(null);
    setTransition('idle');
    setTranscript('');
  };

  if (!started) {
    return (
      <main
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100dvh',
          textAlign: 'center',
          padding: '2rem',
        }}
      >
        <h1 style={{ fontSize: '3rem', marginBottom: '1rem' }}>missless</h1>
        <p style={{ fontSize: '1.25rem', color: 'var(--color-muted)', maxWidth: '400px' }}>
          그리운 사람과의 가상 재회
        </p>
        <button
          onClick={handleStart}
          style={{
            marginTop: '2rem',
            padding: '1rem 2.5rem',
            fontSize: '1.125rem',
            background: 'var(--color-primary)',
            color: 'white',
            border: 'none',
            borderRadius: '2rem',
            cursor: 'pointer',
          }}
        >
          시작하기
        </button>
      </main>
    );
  }

  return (
    <main style={{ position: 'relative', height: '100dvh', width: '100dvw' }}>
      <SceneDisplay previewSrc={previewSrc} finalSrc={finalSrc} />
      <SessionTransition phase={transition} />

      {/* Connection indicator */}
      <div
        style={{
          position: 'absolute',
          top: '1rem',
          right: '1rem',
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          zIndex: 10,
        }}
      >
        <div
          style={{
            width: 8,
            height: 8,
            borderRadius: '50%',
            background: CONNECTION_COLORS[state] ?? '#ef4444',
          }}
        />
        <span style={{ fontSize: '0.75rem', color: 'var(--color-muted)' }}>
          {state}
        </span>
      </div>

      {/* Transcript overlay */}
      {transcript && (
        <div
          style={{
            position: 'absolute',
            bottom: '6rem',
            left: '50%',
            transform: 'translateX(-50%)',
            maxWidth: '80%',
            padding: '0.75rem 1.5rem',
            background: 'rgba(0,0,0,0.6)',
            borderRadius: '1rem',
            color: 'var(--color-text)',
            fontSize: '1rem',
            textAlign: 'center',
            zIndex: 10,
          }}
        >
          {transcript}
        </div>
      )}

      {/* Stop button */}
      <button
        onClick={handleStop}
        style={{
          position: 'absolute',
          bottom: '2rem',
          left: '50%',
          transform: 'translateX(-50%)',
          padding: '0.75rem 2rem',
          fontSize: '1rem',
          background: 'rgba(255,255,255,0.1)',
          color: 'var(--color-text)',
          border: '1px solid rgba(255,255,255,0.2)',
          borderRadius: '2rem',
          cursor: 'pointer',
          zIndex: 10,
        }}
      >
        종료
      </button>
    </main>
  );
}
