'use client';

import { useEffect, useRef } from 'react';

export type ChatMessage = {
  id: string;
  role: 'model' | 'user';
  text: string;
  finished: boolean;
};

export default function ChatPanel({ messages }: { messages: ChatMessage[] }) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  if (messages.length === 0) return null;

  return (
    <div
      style={{
        position: 'absolute',
        bottom: '5rem',
        left: '50%',
        transform: 'translateX(-50%)',
        width: 'min(90%, 600px)',
        maxHeight: '40vh',
        overflowY: 'auto',
        display: 'flex',
        flexDirection: 'column',
        gap: '0.5rem',
        padding: '0.75rem',
        zIndex: 10,
        scrollbarWidth: 'thin',
        scrollbarColor: 'rgba(255,255,255,0.2) transparent',
      }}
    >
      {messages.map((msg) => (
        <div
          key={msg.id}
          style={{
            display: 'flex',
            justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
          }}
        >
          <div
            style={{
              maxWidth: '80%',
              padding: '0.5rem 0.875rem',
              borderRadius:
                msg.role === 'user'
                  ? '1rem 1rem 0.25rem 1rem'
                  : '1rem 1rem 1rem 0.25rem',
              background:
                msg.role === 'user'
                  ? 'rgba(99,102,241,0.7)'
                  : 'rgba(255,255,255,0.1)',
              color: 'var(--color-text)',
              fontSize: '0.875rem',
              lineHeight: 1.5,
              opacity: msg.finished ? 1 : 0.7,
              backdropFilter: 'blur(8px)',
            }}
          >
            {msg.text}
          </div>
        </div>
      ))}
      <div ref={bottomRef} />
    </div>
  );
}
