/* eslint-disable @next/next/no-img-element */
'use client';

import { useEffect, useRef, useState } from 'react';

interface SceneDisplayProps {
  previewSrc: string | null;
  finalSrc: string | null;
}

// Using <img> because scenes are base64 data URLs from Gemini, not optimizable by next/image
export default function SceneDisplay({ previewSrc, finalSrc }: SceneDisplayProps) {
  const [showFinal, setShowFinal] = useState(false);
  const finalRef = useRef<HTMLImageElement>(null);

  useEffect(() => {
    const img = finalRef.current;
    if (!finalSrc || !img) return;

    let mounted = true;

    const handleLoad = () => {
      if (mounted) setShowFinal(true);
    };
    img.addEventListener('load', handleLoad);
    img.src = finalSrc;

    return () => {
      mounted = false;
      img.removeEventListener('load', handleLoad);
    };
  }, [finalSrc]);

  // Reset crossfade when new preview arrives
  useEffect(() => {
    if (previewSrc) {
      setShowFinal(false);
    }
  }, [previewSrc]);

  const currentSrc = previewSrc || finalSrc;
  if (!currentSrc) return null;

  return (
    <div style={{ position: 'relative', width: '100%', height: '100%' }}>
      {previewSrc && (
        <img
          src={previewSrc}
          alt="Scene preview"
          style={{
            position: 'absolute',
            inset: 0,
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            opacity: showFinal ? 0 : 1,
            transition: 'opacity 1.2s ease-in-out',
          }}
        />
      )}
      {finalSrc && (
        <img
          ref={finalRef}
          alt="Scene final"
          style={{
            position: 'absolute',
            inset: 0,
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            opacity: showFinal ? 1 : 0,
            transition: 'opacity 1.2s ease-in-out',
          }}
        />
      )}
    </div>
  );
}
