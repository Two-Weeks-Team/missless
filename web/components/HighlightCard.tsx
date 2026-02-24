'use client';

export interface Highlight {
  timestamp: string;
  description: string;
  expression: string;
}

interface HighlightCardProps {
  highlights: Highlight[];
}

export default function HighlightCard({ highlights }: HighlightCardProps) {
  if (highlights.length === 0) return null;

  return (
    <div
      style={{
        position: 'absolute',
        top: '4rem',
        right: '1rem',
        maxHeight: '60vh',
        overflowY: 'auto',
        display: 'flex',
        flexDirection: 'column',
        gap: '0.5rem',
        zIndex: 20,
        width: '280px',
      }}
    >
      {highlights.map((h, i) => (
        <div
          key={`${h.timestamp}-${i}`}
          style={{
            padding: '0.75rem 1rem',
            background: 'rgba(0,0,0,0.6)',
            backdropFilter: 'blur(8px)',
            borderRadius: '0.75rem',
            borderLeft: '3px solid var(--color-primary)',
            animation: 'slideIn 0.3s ease-out',
          }}
        >
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              marginBottom: '0.25rem',
            }}
          >
            <span style={{ fontSize: '0.75rem', color: 'var(--color-primary)' }}>
              {h.timestamp}
            </span>
            <span style={{ fontSize: '0.75rem', color: 'var(--color-muted)' }}>
              {h.expression}
            </span>
          </div>
          <p style={{ fontSize: '0.875rem', color: 'var(--color-text)', margin: 0 }}>
            {h.description}
          </p>
        </div>
      ))}
    </div>
  );
}
