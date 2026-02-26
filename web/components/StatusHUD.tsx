'use client';

type StatusHUDProps = {
  connection: string;
  isRecording: boolean;
  isPlaying: boolean;
  sessionState: string;
};

const STATE_LABELS: Record<string, string> = {
  welcome: 'Ready',
  onboarding: 'Onboarding',
  youtube_grid: 'Select Video',
  person_select: 'Select Person',
  analyzing: 'Analyzing...',
  transition: 'Transitioning...',
  reunion: 'Reunion',
};

export default function StatusHUD({
  connection,
  isRecording,
  isPlaying,
  sessionState,
}: StatusHUDProps) {
  const connColor =
    connection === 'connected'
      ? '#4ade80'
      : connection === 'connecting'
        ? '#fbbf24'
        : '#ef4444';

  return (
    <div
      style={{
        position: 'absolute',
        top: '1rem',
        left: '1rem',
        background: 'rgba(0,0,0,0.5)',
        backdropFilter: 'blur(12px)',
        borderRadius: '0.75rem',
        padding: '0.625rem 0.875rem',
        display: 'flex',
        flexDirection: 'column',
        gap: '0.375rem',
        fontSize: '0.75rem',
        color: 'var(--color-text)',
        zIndex: 20,
        minWidth: '120px',
      }}
    >
      <div style={{ fontWeight: 600, fontSize: '0.8125rem', marginBottom: '0.125rem' }}>
        missless
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.375rem' }}>
        <div
          style={{
            width: 6,
            height: 6,
            borderRadius: '50%',
            background: connColor,
          }}
        />
        <span style={{ color: 'var(--color-muted)' }}>{connection}</span>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.375rem' }}>
        <span style={{ color: 'var(--color-muted)' }}>State:</span>
        <span>{STATE_LABELS[sessionState] ?? sessionState}</span>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        {isRecording && (
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
            <div
              style={{
                width: 6,
                height: 6,
                borderRadius: '50%',
                background: '#ef4444',
                animation: 'pulse 1.5s infinite',
              }}
            />
            Mic
          </span>
        )}
        {isPlaying && (
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
            <div
              style={{
                width: 6,
                height: 6,
                borderRadius: '50%',
                background: '#4ade80',
                animation: 'pulse 1.5s infinite',
              }}
            />
            Speaking
          </span>
        )}
      </div>
    </div>
  );
}
