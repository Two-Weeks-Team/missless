'use client';

export interface YouTubeVideo {
  id: string;
  title: string;
  thumbnail: string;
  privacyStatus: string;
}

interface YouTubeGridProps {
  videos: YouTubeVideo[];
  onSelect: (videoId: string) => void;
  onPrivateClick: () => void;
}

export default function YouTubeGrid({ videos, onSelect, onPrivateClick }: YouTubeGridProps) {
  if (videos.length === 0) return null;

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        background: 'var(--color-bg)',
        zIndex: 30,
        overflowY: 'auto',
        padding: '1rem',
      }}
    >
      <h2
        style={{
          fontSize: '1.25rem',
          marginBottom: '1rem',
          textAlign: 'center',
          color: 'var(--color-text)',
        }}
      >
        Select a video
      </h2>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))',
          gap: '0.75rem',
          maxWidth: '600px',
          margin: '0 auto',
        }}
      >
        {videos.map((video) => {
          const isPrivate = video.privacyStatus !== 'public';
          return (
            <button
              key={video.id}
              onClick={() => isPrivate ? onPrivateClick() : onSelect(video.id)}
              style={{
                position: 'relative',
                background: 'var(--color-surface)',
                border: 'none',
                borderRadius: '0.75rem',
                overflow: 'hidden',
                cursor: 'pointer',
                padding: 0,
                opacity: isPrivate ? 0.5 : 1,
                transition: 'transform 0.2s',
              }}
              onMouseEnter={(e) => {
                if (!isPrivate) e.currentTarget.style.transform = 'scale(1.03)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'scale(1)';
              }}
            >
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={video.thumbnail}
                alt={video.title}
                style={{
                  width: '100%',
                  aspectRatio: '16/9',
                  objectFit: 'cover',
                  display: 'block',
                }}
              />
              <div style={{ padding: '0.5rem' }}>
                <p
                  style={{
                    fontSize: '0.75rem',
                    color: 'var(--color-text)',
                    margin: 0,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {video.title}
                </p>
              </div>
              {isPrivate && (
                <div
                  style={{
                    position: 'absolute',
                    top: '0.25rem',
                    right: '0.25rem',
                    background: 'rgba(0,0,0,0.7)',
                    color: 'var(--color-muted)',
                    fontSize: '0.625rem',
                    padding: '0.125rem 0.375rem',
                    borderRadius: '0.25rem',
                  }}
                >
                  Private
                </div>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
