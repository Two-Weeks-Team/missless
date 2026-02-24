'use client';

export default function Home() {
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
