import { useCallback, useRef, useState } from 'react';

export function useAudio() {
  const audioCtxRef = useRef<AudioContext | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);

  // Must be called from user gesture (touch/click) on mobile
  const initAudioContext = useCallback(() => {
    if (!audioCtxRef.current) {
      audioCtxRef.current = new AudioContext({ sampleRate: 24000 });
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
    source.onended = () => setIsPlaying(false);
    source.start();
    setIsPlaying(true);
  }, []);

  const cleanup = useCallback(() => {
    audioCtxRef.current?.close();
    audioCtxRef.current = null;
  }, []);

  return { initAudioContext, playPCM, isPlaying, cleanup };
}
