'use client';

import { useEffect, useState } from 'react';

type TransitionPhase = 'idle' | 'transitioning' | 'ready';

interface SessionTransitionProps {
  phase: TransitionPhase;
}

export default function SessionTransition({ phase }: SessionTransitionProps) {
  const [opacity, setOpacity] = useState(0);

  useEffect(() => {
    if (phase === 'transitioning') {
      setOpacity(1);
    } else if (phase === 'ready') {
      const timer = setTimeout(() => setOpacity(0), 800);
      return () => clearTimeout(timer);
    }
  }, [phase]);

  if (phase === 'idle') return null;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--color-bg)',
        opacity,
        transition: 'opacity 0.8s ease-in-out',
        zIndex: 50,
        pointerEvents: phase === 'transitioning' ? 'auto' : 'none',
      }}
    >
      <div style={{ textAlign: 'center' }}>
        {phase === 'transitioning' && (
          <>
            <div
              style={{
                width: 40,
                height: 40,
                border: '3px solid var(--color-muted)',
                borderTopColor: 'var(--color-primary)',
                borderRadius: '50%',
                animation: 'spin 1s linear infinite',
                margin: '0 auto 1.5rem',
              }}
            />
            <p style={{ color: 'var(--color-muted)', fontSize: '1rem' }}>
              Please wait...
            </p>
          </>
        )}
        {phase === 'ready' && (
          <p style={{ fontSize: '1.25rem', color: 'var(--color-text)' }}>
            Ready
          </p>
        )}
      </div>
    </div>
  );
}
