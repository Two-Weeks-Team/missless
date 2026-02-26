import { useCallback, useRef, useState } from 'react';

// Target sample rate for Gemini Live API input.
const TARGET_SAMPLE_RATE = 16000;

// ScriptProcessorNode buffer size (4096 is a good balance of latency vs. efficiency).
const BUFFER_SIZE = 4096;

export function useMicrophone() {
  const streamRef = useRef<MediaStream | null>(null);
  const ctxRef = useRef<AudioContext | null>(null);
  const processorRef = useRef<ScriptProcessorNode | null>(null);
  const [isRecording, setIsRecording] = useState(false);
  const onDataRef = useRef<((pcm: ArrayBuffer) => void) | null>(null);

  const start = useCallback(async (onData: (pcm: ArrayBuffer) => void) => {
    if (streamRef.current) return; // already recording

    onDataRef.current = onData;

    const stream = await navigator.mediaDevices.getUserMedia({
      audio: {
        echoCancellation: true,
        noiseSuppression: true,
        sampleRate: TARGET_SAMPLE_RATE,
      },
    });
    streamRef.current = stream;

    const ctx = new AudioContext({ sampleRate: TARGET_SAMPLE_RATE });
    ctxRef.current = ctx;

    const source = ctx.createMediaStreamSource(stream);
    const processor = ctx.createScriptProcessor(BUFFER_SIZE, 1, 1);
    processorRef.current = processor;

    processor.onaudioprocess = (e) => {
      const float32 = e.inputBuffer.getChannelData(0);
      // Convert Float32 [-1, 1] → Int16 [-32768, 32767]
      const int16 = new Int16Array(float32.length);
      for (let i = 0; i < float32.length; i++) {
        const s = Math.max(-1, Math.min(1, float32[i]));
        int16[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
      }
      onDataRef.current?.(int16.buffer);
    };

    source.connect(processor);
    processor.connect(ctx.destination);
    setIsRecording(true);
  }, []);

  const stop = useCallback(() => {
    processorRef.current?.disconnect();
    processorRef.current = null;

    ctxRef.current?.close();
    ctxRef.current = null;

    streamRef.current?.getTracks().forEach((t) => t.stop());
    streamRef.current = null;

    onDataRef.current = null;
    setIsRecording(false);
  }, []);

  return { start, stop, isRecording };
}
