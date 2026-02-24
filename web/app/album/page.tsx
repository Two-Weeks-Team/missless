'use client';

import { useSearchParams } from 'next/navigation';
import { Suspense } from 'react';

interface AlbumScene {
  imageUrl: string;
  caption: string;
  timestamp: string;
}

function AlbumContent() {
  const searchParams = useSearchParams();
  const albumId = searchParams.get('id');
  const persona = searchParams.get('persona') ?? 'Someone special';
  const scenesParam = searchParams.get('scenes');

  let scenes: AlbumScene[] = [];
  if (scenesParam) {
    try {
      scenes = JSON.parse(decodeURIComponent(scenesParam));
    } catch {
      // Invalid scenes data
    }
  }

  return (
    <main
      style={{
        minHeight: '100dvh',
        background: 'var(--color-bg)',
        padding: '2rem 1rem',
      }}
    >
      {/* Header */}
      <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>missless</h1>
        <p style={{ fontSize: '1.125rem', color: 'var(--color-muted)' }}>
          {persona}와의 추억 앨범
        </p>
        {albumId && (
          <p style={{ fontSize: '0.75rem', color: 'var(--color-muted)', marginTop: '0.5rem' }}>
            Album #{albumId.slice(0, 8)}
          </p>
        )}
      </div>

      {/* Scene gallery */}
      {scenes.length > 0 ? (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            gap: '1rem',
            maxWidth: '900px',
            margin: '0 auto',
          }}
        >
          {scenes.map((scene, i) => (
            <div
              key={i}
              style={{
                background: 'var(--color-surface)',
                borderRadius: '1rem',
                overflow: 'hidden',
              }}
            >
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={scene.imageUrl.startsWith('data:')
                  ? scene.imageUrl
                  : `data:image/jpeg;base64,${scene.imageUrl}`}
                alt={scene.caption || `Scene ${i + 1}`}
                style={{
                  width: '100%',
                  aspectRatio: '16/9',
                  objectFit: 'cover',
                  display: 'block',
                }}
              />
              {scene.caption && (
                <div style={{ padding: '0.75rem' }}>
                  <p style={{ fontSize: '0.875rem', color: 'var(--color-text)', margin: 0 }}>
                    {scene.caption}
                  </p>
                </div>
              )}
            </div>
          ))}
        </div>
      ) : (
        <div style={{ textAlign: 'center', marginTop: '4rem' }}>
          <p style={{ fontSize: '1rem', color: 'var(--color-muted)' }}>
            아직 추억이 없어요. 재회를 시작해보세요.
          </p>
        </div>
      )}

      {/* Share section */}
      <div style={{ textAlign: 'center', marginTop: '3rem' }}>
        <button
          onClick={() => {
            if (navigator.share) {
              navigator.share({
                title: `${persona}와의 추억 앨범 | missless`,
                text: `missless에서 ${persona}와 다시 만났어요.`,
                url: window.location.href,
              }).catch(() => {});
            } else {
              navigator.clipboard.writeText(window.location.href).catch(() => {});
            }
          }}
          style={{
            padding: '0.75rem 2rem',
            fontSize: '1rem',
            background: 'var(--color-primary)',
            color: 'white',
            border: 'none',
            borderRadius: '2rem',
            cursor: 'pointer',
          }}
        >
          공유하기
        </button>
      </div>
    </main>
  );
}

export default function AlbumPage() {
  return (
    <Suspense fallback={
      <main style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100dvh' }}>
        <p style={{ color: 'var(--color-muted)' }}>Loading...</p>
      </main>
    }>
      <AlbumContent />
    </Suspense>
  );
}
