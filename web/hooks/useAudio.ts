import { useCallback, useRef, useState } from 'react';

export function useAudio() {
  const audioCtxRef = useRef<AudioContext | null>(null);
  const nextStartTimeRef = useRef(0);
  const [isPlaying, setIsPlaying] = useState(false);

  // Must be called from user gesture (touch/click) on mobile
  const initAudioContext = useCallback(() => {
    if (!audioCtxRef.current) {
      audioCtxRef.current = new AudioContext({ sampleRate: 24000 });
      nextStartTimeRef.current = 0;
    }
    if (audioCtxRef.current.state === 'suspended') {
      audioCtxRef.current.resume();
    }
    return audioCtxRef.current;
  }, []);

  const playPCM = useCallback((pcmData: ArrayBuffer) => {
    const ctx = audioCtxRef.current;
    if (!ctx) return;

    const samples = new Int16Array(pcmData);
    const float32 = new Float32Array(samples.length);
    for (let i = 0; i < samples.length; i++) {
      float32[i] = samples[i] / 32768;
    }

    const buffer = ctx.createBuffer(1, float32.length, 24000);
    buffer.copyToChannel(float32, 0);

    const source = ctx.createBufferSource();
    source.buffer = buffer;
    source.connect(ctx.destination);

    // Schedule sequentially to prevent overlap.
    const now = ctx.currentTime;
    const startAt = Math.max(now, nextStartTimeRef.current);
    nextStartTimeRef.current = startAt + buffer.duration;

    source.onended = () => {
      // Mark not playing only when no more queued audio remains.
      if (ctx.currentTime >= nextStartTimeRef.current - 0.01) {
        setIsPlaying(false);
      }
    };
    source.start(startAt);
    setIsPlaying(true);
  }, []);

  const cleanup = useCallback(() => {
    audioCtxRef.current?.close();
    audioCtxRef.current = null;
    nextStartTimeRef.current = 0;
  }, []);

  return { initAudioContext, playPCM, isPlaying, cleanup };
}
