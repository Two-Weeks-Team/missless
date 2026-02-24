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
          비공개 영상입니다
        </p>
        <p style={{ fontSize: '0.875rem', color: 'var(--color-muted)', marginBottom: '1.5rem' }}>
          비공개 또는 미등록 영상은 직접 분석할 수 없어요.
          공개 영상을 선택하거나, 영상을 직접 업로드해주세요.
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
          확인
        </button>
      </div>
    </div>
  );
}
