import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useMicrophone } from '../../hooks/useMicrophone';

// Mock getUserMedia
const mockGetUserMedia = vi.fn();
const mockTrackStop = vi.fn();

const mockStream = {
  getTracks: () => [{ stop: mockTrackStop }],
};

// Mock AudioContext + ScriptProcessorNode
const mockDisconnect = vi.fn();
const mockConnect = vi.fn();
const mockProcessorConnect = vi.fn();
const mockClose = vi.fn();

let audioProcessHandler: ((e: { inputBuffer: { getChannelData: (ch: number) => Float32Array } }) => void) | null = null;

const mockProcessor = {
  connect: mockProcessorConnect,
  disconnect: mockDisconnect,
  set onaudioprocess(fn: typeof audioProcessHandler) {
    audioProcessHandler = fn;
  },
  get onaudioprocess() {
    return audioProcessHandler;
  },
};

const mockSource = {
  connect: mockConnect,
};

class MockAudioContext {
  sampleRate = 16000;
  close = mockClose;
  destination = {};
  createMediaStreamSource = vi.fn(() => mockSource);
  createScriptProcessor = vi.fn(() => mockProcessor);
}

beforeEach(() => {
  vi.clearAllMocks();
  audioProcessHandler = null;
  mockGetUserMedia.mockResolvedValue(mockStream);
  vi.stubGlobal('AudioContext', MockAudioContext);
  vi.stubGlobal('navigator', {
    mediaDevices: { getUserMedia: mockGetUserMedia },
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('useMicrophone', () => {
  it('initial isRecording is false', () => {
    const { result } = renderHook(() => useMicrophone());
    expect(result.current.isRecording).toBe(false);
  });

  it('start requests microphone and sets isRecording to true', async () => {
    const { result } = renderHook(() => useMicrophone());
    const onData = vi.fn();

    await act(async () => {
      await result.current.start(onData);
    });

    expect(mockGetUserMedia).toHaveBeenCalledWith(
      expect.objectContaining({
        audio: expect.objectContaining({
          echoCancellation: true,
          noiseSuppression: true,
        }),
      }),
    );
    expect(result.current.isRecording).toBe(true);
  });

  it('sends PCM data via onData callback when audio is processed', async () => {
    const { result } = renderHook(() => useMicrophone());
    const onData = vi.fn();

    await act(async () => {
      await result.current.start(onData);
    });

    // Simulate audio processing
    const float32 = new Float32Array([0.5, -0.5, 0, 1.0]);
    act(() => {
      audioProcessHandler?.({
        inputBuffer: { getChannelData: () => float32 },
      });
    });

    expect(onData).toHaveBeenCalledTimes(1);
    const buffer = onData.mock.calls[0][0] as ArrayBuffer;
    expect(buffer).toBeInstanceOf(ArrayBuffer);

    // Verify Int16 conversion
    const int16 = new Int16Array(buffer);
    expect(int16.length).toBe(4);
    expect(int16[0]).toBeGreaterThan(0); // 0.5 → positive
    expect(int16[1]).toBeLessThan(0);    // -0.5 → negative
  });

  it('stop cleans up resources and sets isRecording to false', async () => {
    const { result } = renderHook(() => useMicrophone());
    const onData = vi.fn();

    await act(async () => {
      await result.current.start(onData);
    });
    expect(result.current.isRecording).toBe(true);

    act(() => {
      result.current.stop();
    });

    expect(mockDisconnect).toHaveBeenCalled();
    expect(mockClose).toHaveBeenCalled();
    expect(mockTrackStop).toHaveBeenCalled();
    expect(result.current.isRecording).toBe(false);
  });

  it('stop is safe to call without start', () => {
    const { result } = renderHook(() => useMicrophone());

    // Should not throw
    act(() => {
      result.current.stop();
    });

    expect(result.current.isRecording).toBe(false);
  });
});
