'use client';

interface ProgressBarProps {
  step: string;
  percent: number;
}

export default function ProgressBar({ step, percent }: ProgressBarProps) {
  return (
    <div
      style={{
        position: 'absolute',
        bottom: '8rem',
        left: '50%',
        transform: 'translateX(-50%)',
        width: '80%',
        maxWidth: '400px',
        zIndex: 20,
      }}
    >
      <p
        style={{
          fontSize: '0.875rem',
          color: 'var(--color-muted)',
          marginBottom: '0.5rem',
          textAlign: 'center',
        }}
      >
        {step}
      </p>
      <div
        style={{
          width: '100%',
          height: 6,
          background: 'rgba(255,255,255,0.1)',
          borderRadius: 3,
          overflow: 'hidden',
        }}
      >
        <div
          data-testid="progress-fill"
          style={{
            width: `${Math.min(100, Math.max(0, percent))}%`,
            height: '100%',
            background: 'var(--color-primary)',
            borderRadius: 3,
            transition: 'width 0.4s ease-out',
          }}
        />
      </div>
      <p
        style={{
          fontSize: '0.75rem',
          color: 'var(--color-muted)',
          marginTop: '0.25rem',
          textAlign: 'right',
        }}
      >
        {percent}%
      </p>
    </div>
  );
}
