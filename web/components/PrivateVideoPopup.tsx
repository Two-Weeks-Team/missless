'use client';

interface PrivateVideoPopupProps {
  onClose: () => void;
}

export default function PrivateVideoPopup({ onClose }: PrivateVideoPopupProps) {
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.7)',
        zIndex: 40,
      }}
      onClick={onClose}
    >
      <div
        style={{
          background: 'var(--color-surface)',
          borderRadius: '1rem',
          padding: '2rem',
          maxWidth: '320px',
          textAlign: 'center',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <p style={{ fontSize: '1rem', color: 'var(--color-text)', marginBottom: '0.5rem' }}>
          This video is private or unlisted
        </p>
        <p style={{ fontSize: '0.875rem', color: 'var(--color-muted)', marginBottom: '1.5rem' }}>
          Private or unlisted videos cannot be analyzed directly.
          Please select a public video or upload the video file.
        </p>
        <button
          onClick={onClose}
          style={{
            padding: '0.5rem 2rem',
            background: 'var(--color-primary)',
            color: 'white',
            border: 'none',
            borderRadius: '2rem',
            cursor: 'pointer',
            fontSize: '0.875rem',
          }}
        >
          OK
        </button>
      </div>
    </div>
  );
}
