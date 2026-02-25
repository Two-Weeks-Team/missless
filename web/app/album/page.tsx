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
      const parsed: AlbumScene[] = JSON.parse(decodeURIComponent(scenesParam));
      scenes = parsed.filter((s) => {
        if (!s.imageUrl) return false;
        const isBase64 = /^[A-Za-z0-9+/=]+$/.test(s.imageUrl);
        const isSafeDataUri = s.imageUrl.startsWith('data:image/') && !s.imageUrl.startsWith('data:image/svg+xml');
        const isHttps = s.imageUrl.startsWith('https://');
        return isBase64 || isSafeDataUri || isHttps;
      });
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
          Reunion Album with {persona}
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
            No memories yet. Start a reunion to create them.
          </p>
        </div>
      )}

      {/* Share section */}
      <div style={{ textAlign: 'center', marginTop: '3rem' }}>
        <button
          onClick={() => {
            if (navigator.share) {
              navigator.share({
                title: `Reunion Album with ${persona} | missless`,
                text: `I reunited with ${persona} on missless.`,
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
          Share
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
