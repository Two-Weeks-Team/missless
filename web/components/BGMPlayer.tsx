'use client';

import { useCallback, useEffect, useRef } from 'react';

/** Volume level for BGM (-20dB relative to full scale). */
const BGM_VOLUME = 0.1;
/** Crossfade duration in milliseconds. */
const CROSSFADE_MS = 2000;

interface BGMPlayerProps {
  /** Current BGM URL to play. null stops playback. */
  bgmUrl: string | null;
}

export default function BGMPlayer({ bgmUrl }: BGMPlayerProps) {
  const currentAudioRef = useRef<HTMLAudioElement | null>(null);
  const nextAudioRef = useRef<HTMLAudioElement | null>(null);
  const fadeIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const clearFade = useCallback(() => {
    if (fadeIntervalRef.current) {
      clearInterval(fadeIntervalRef.current);
      fadeIntervalRef.current = null;
    }
  }, []);

  const crossfadeTo = useCallback((url: string) => {
    clearFade();

    const next = new Audio(url);
    next.loop = true;
    next.volume = 0;
    nextAudioRef.current = next;

    const prev = currentAudioRef.current;
    const steps = CROSSFADE_MS / 50;
    let step = 0;

    next.play().catch(() => {
      // Autoplay may be blocked; BGM is non-critical.
    });

    fadeIntervalRef.current = setInterval(() => {
      step++;
      const progress = Math.min(step / steps, 1);

      next.volume = BGM_VOLUME * progress;
      if (prev) {
        prev.volume = BGM_VOLUME * (1 - progress);
      }

      if (progress >= 1) {
        clearFade();
        if (prev) {
          prev.pause();
          prev.src = '';
        }
        currentAudioRef.current = next;
        nextAudioRef.current = null;
      }
    }, 50);
  }, [clearFade]);

  const stopBGM = useCallback(() => {
    clearFade();
    if (currentAudioRef.current) {
      currentAudioRef.current.pause();
      currentAudioRef.current.src = '';
      currentAudioRef.current = null;
    }
    if (nextAudioRef.current) {
      nextAudioRef.current.pause();
      nextAudioRef.current.src = '';
      nextAudioRef.current = null;
    }
  }, [clearFade]);

  useEffect(() => {
    if (!bgmUrl) {
      stopBGM();
      return;
    }

    // If same URL is already playing, skip.
    if (currentAudioRef.current && currentAudioRef.current.src.endsWith(bgmUrl)) {
      return;
    }

    crossfadeTo(bgmUrl);
  }, [bgmUrl, crossfadeTo, stopBGM]);

  // Cleanup on unmount.
  useEffect(() => {
    return () => {
      stopBGM();
    };
  }, [stopBGM]);

  // This component renders nothing visible.
  return null;
}
