import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAudio } from '../../hooks/useAudio';

// ---------------------------------------------------------------------------
// Mock AudioContext & related Web Audio API objects
// ---------------------------------------------------------------------------

const mockStart = vi.fn();
const mockConnect = vi.fn();
const mockClose = vi.fn();
const mockResume = vi.fn().mockResolvedValue(undefined);
const mockCopyToChannel = vi.fn();

let mockSource: {
  buffer: { copyToChannel: typeof mockCopyToChannel } | null;
  connect: typeof mockConnect;
  start: typeof mockStart;
  onended: (() => void) | null;
};

let latestContext: InstanceType<typeof MockAudioContext> | null = null;
let constructorCalls: Array<Record<string, unknown>> = [];
let initialState = 'running';

class MockAudioContext {
  state = initialState;
  sampleRate = 24000;
  destination = {};
  resume = mockResume;
  close = mockClose;
  createBuffer = vi.fn(() => ({
    copyToChannel: mockCopyToChannel,
  }));
  createBufferSource = vi.fn(() => mockSource);

  constructor(options?: { sampleRate?: number }) {
    constructorCalls.push(options ?? {});
    if (options?.sampleRate) {
      this.sampleRate = options.sampleRate;
    }
    latestContext = this;
  }
}

beforeEach(() => {
  vi.clearAllMocks();
  latestContext = null;
  constructorCalls = [];
  initialState = 'running';

  mockSource = {
    buffer: null,
    connect: mockConnect,
    start: mockStart,
    onended: null,
  };

  vi.stubGlobal('AudioContext', MockAudioContext);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useAudio', () => {
  // 1. Initial isPlaying is false ---------------------------------------------
  it('initial isPlaying is false', () => {
    const { result } = renderHook(() => useAudio());
    expect(result.current.isPlaying).toBe(false);
  });

  // 2. initAudioContext creates AudioContext with 24000 Hz --------------------
  it('creates AudioContext with sampleRate 24000', () => {
    const { result } = renderHook(() => useAudio());

    act(() => {
      result.current.initAudioContext();
    });

    expect(constructorCalls).toHaveLength(1);
    expect(constructorCalls[0]).toEqual({ sampleRate: 24000 });
  });

  // 3. initAudioContext resumes if suspended ----------------------------------
  it('resumes AudioContext when state is suspended', () => {
    // Set initial state to suspended before creating the context
    initialState = 'suspended';

    const { result } = renderHook(() => useAudio());

    act(() => {
      result.current.initAudioContext();
    });

    expect(mockResume).toHaveBeenCalledTimes(1);
  });

  // 4. initAudioContext reuses existing context --------------------------------
  it('reuses existing AudioContext on subsequent calls', () => {
    const { result } = renderHook(() => useAudio());

    let ctx1: AudioContext | undefined;
    let ctx2: AudioContext | undefined;

    act(() => {
      ctx1 = result.current.initAudioContext();
    });
    act(() => {
      ctx2 = result.current.initAudioContext();
    });

    // Constructor should only be called once
    expect(constructorCalls).toHaveLength(1);
    expect(ctx1).toBe(ctx2);
  });

  // 5. playPCM does nothing without context -----------------------------------
  it('playPCM is a no-op when AudioContext has not been initialized', () => {
    const { result } = renderHook(() => useAudio());

    const pcm = new ArrayBuffer(100);
    act(() => {
      result.current.playPCM(pcm);
    });

    // No context was created, so latestContext is null
    expect(latestContext).toBeNull();
  });

  // 6. playPCM creates buffer, source, connects and starts --------------------
  it('playPCM creates an audio buffer, connects to destination, and starts playback', () => {
    const { result } = renderHook(() => useAudio());

    // Initialize first
    act(() => {
      result.current.initAudioContext();
    });

    // Create 4-sample PCM payload (8 bytes for Int16)
    const pcm = new ArrayBuffer(8);
    const view = new Int16Array(pcm);
    view[0] = 16384; // 0.5 in float
    view[1] = -16384; // -0.5 in float
    view[2] = 0;
    view[3] = 32767; // ~1.0 in float

    act(() => {
      result.current.playPCM(pcm);
    });

    // createBuffer called with 1 channel, 4 samples, 24000 Hz
    expect(latestContext!.createBuffer).toHaveBeenCalledWith(1, 4, 24000);

    // copyToChannel called with Float32Array on channel 0
    expect(mockCopyToChannel).toHaveBeenCalledTimes(1);
    const float32Arg = mockCopyToChannel.mock.calls[0][0];
    expect(float32Arg).toBeInstanceOf(Float32Array);
    expect(float32Arg.length).toBe(4);
    // Verify Int16 -> Float32 conversion (16384 / 32768 = 0.5)
    expect(float32Arg[0]).toBeCloseTo(0.5, 4);
    expect(float32Arg[1]).toBeCloseTo(-0.5, 4);
    expect(mockCopyToChannel.mock.calls[0][1]).toBe(0);

    // createBufferSource called
    expect(latestContext!.createBufferSource).toHaveBeenCalledTimes(1);

    // source connected to destination
    expect(mockConnect).toHaveBeenCalledWith(latestContext!.destination);

    // source started
    expect(mockStart).toHaveBeenCalledTimes(1);

    // isPlaying is now true
    expect(result.current.isPlaying).toBe(true);
  });

  // 6b. playPCM sets isPlaying false when source ends -------------------------
  it('sets isPlaying to false when audio source ends', () => {
    const { result } = renderHook(() => useAudio());

    act(() => {
      result.current.initAudioContext();
    });

    const pcm = new ArrayBuffer(4);
    act(() => {
      result.current.playPCM(pcm);
    });
    expect(result.current.isPlaying).toBe(true);

    // Simulate the source ending
    act(() => {
      mockSource.onended?.();
    });
    expect(result.current.isPlaying).toBe(false);
  });

  // 7. cleanup closes context and nullifies ref --------------------------------
  it('cleanup closes AudioContext', () => {
    const { result } = renderHook(() => useAudio());

    act(() => {
      result.current.initAudioContext();
    });

    act(() => {
      result.current.cleanup();
    });

    expect(mockClose).toHaveBeenCalledTimes(1);

    // After cleanup, playPCM should be a no-op (ref is null)
    mockStart.mockClear();
    const pcm = new ArrayBuffer(4);
    act(() => {
      result.current.playPCM(pcm);
    });
    expect(mockStart).not.toHaveBeenCalled();
  });

  // 7b. cleanup is safe to call without initialization -------------------------
  it('cleanup is safe to call when AudioContext was never initialized', () => {
    const { result } = renderHook(() => useAudio());

    // Should not throw
    act(() => {
      result.current.cleanup();
    });

    expect(mockClose).not.toHaveBeenCalled();
  });
});
