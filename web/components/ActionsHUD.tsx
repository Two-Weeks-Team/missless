'use client';

type ActionsHUDProps = {
  sessionState: string;
};

type ActionItem = {
  label: string;
  hint: string;
};

const ONBOARDING_ACTIONS: ActionItem[] = [
  { label: 'Talk', hint: 'Tell missless who you miss' },
  { label: 'Share Video', hint: 'Paste a YouTube link' },
];

const REUNION_ACTIONS: ActionItem[] = [
  { label: 'Talk', hint: 'Have a conversation' },
  { label: 'Scene', hint: '"Paint me a picture of..."' },
  { label: 'Music', hint: '"Play something peaceful"' },
  { label: 'Memory', hint: '"Remember when we..."' },
  { label: 'Album', hint: '"Save this moment"' },
];

export default function ActionsHUD({ sessionState }: ActionsHUDProps) {
  const actions = sessionState === 'reunion' ? REUNION_ACTIONS : ONBOARDING_ACTIONS;

  return (
    <div
      style={{
        position: 'absolute',
        top: '1rem',
        right: '1rem',
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
        minWidth: '140px',
      }}
    >
      <div style={{ fontWeight: 600, fontSize: '0.8125rem', marginBottom: '0.125rem' }}>
        You can...
      </div>
      {actions.map((a) => (
        <div key={a.label} style={{ display: 'flex', flexDirection: 'column', gap: '0.0625rem' }}>
          <span style={{ fontWeight: 500 }}>{a.label}</span>
          <span style={{ color: 'var(--color-muted)', fontSize: '0.6875rem' }}>{a.hint}</span>
        </div>
      ))}
    </div>
  );
}
