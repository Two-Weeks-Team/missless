'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useWebSocket, ServerMessage } from '../hooks/useWebSocket';
import { useAudio } from '../hooks/useAudio';
import SceneDisplay from '../components/SceneDisplay';
import SessionTransition from '../components/SessionTransition';
import OnboardingFlow, { type OnboardingStage } from '../components/OnboardingFlow';
import BGMPlayer from '../components/BGMPlayer';
import type { YouTubeVideo } from '../components/YouTubeGrid';
import type { Highlight } from '../components/HighlightCard';

type TransitionPhase = 'idle' | 'transitioning' | 'ready';

const CONNECTION_COLORS: Record<string, string> = {
  connected: '#4ade80',
  connecting: '#fbbf24',
  disconnected: '#ef4444',
  error: '#ef4444',
};

export default function Home() {
  const [started, setStarted] = useState(false);
  const [previewSrc, setPreviewSrc] = useState<string | null>(null);
  const [finalSrc, setFinalSrc] = useState<string | null>(null);
  const [transition, setTransition] = useState<TransitionPhase>('idle');
  const [transcript, setTranscript] = useState<string>('');
  const readyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Onboarding state
  const [onboardingStage, setOnboardingStage] = useState<OnboardingStage>('welcome');
  const [videos, setVideos] = useState<YouTubeVideo[]>([]);
  const [personCrops, setPersonCrops] = useState<string[]>([]);
  const [highlights, setHighlights] = useState<Highlight[]>([]);
  const [analysisStep, setAnalysisStep] = useState('');
  const [analysisPercent, setAnalysisPercent] = useState(0);
  const [bgmUrl, setBgmUrl] = useState<string | null>(null);

  const { initAudioContext, playPCM, cleanup: cleanupAudio } = useAudio();

  const handleMessage = useCallback((msg: ServerMessage) => {
    switch (msg.type) {
      case 'audio':
        playPCM(msg.data);
        break;
      case 'scene_preview':
        setPreviewSrc(msg.image);
        break;
      case 'scene_final':
        setFinalSrc(msg.image);
        break;
      case 'session_transition':
        setTransition('transitioning');
        setOnboardingStage('transition');
        break;
      case 'session_ready':
        setTransition('ready');
        setOnboardingStage('reunion');
        break;
      case 'transcript':
        setTranscript(msg.text);
        break;
      case 'youtube_videos':
        setVideos(msg.videos as YouTubeVideo[]);
        setOnboardingStage('youtube_grid');
        break;
      case 'person_detected':
        setPersonCrops(msg.crops);
        setOnboardingStage('person_select');
        break;
      case 'atmosphere_change':
        if (msg.bgm_url) {
          setBgmUrl(msg.bgm_url);
        }
        break;
      case 'analysis_progress':
        setOnboardingStage('analyzing');
        setAnalysisStep(msg.step);
        setAnalysisPercent(msg.percent);
        if (msg.highlight) {
          try {
            const h = JSON.parse(msg.highlight) as Highlight;
            setHighlights((prev) => [...prev, h]);
          } catch {
            // Highlight may be a plain string description
            setHighlights((prev) => [
              ...prev,
              { timestamp: '', description: msg.highlight as string, expression: '' },
            ]);
          }
        }
        break;
    }
  }, [playPCM]);

  // Clean up transition timer on unmount or phase change
  useEffect(() => {
    if (transition === 'ready') {
      readyTimerRef.current = setTimeout(() => setTransition('idle'), 1600);
    }
    return () => {
      if (readyTimerRef.current) {
        clearTimeout(readyTimerRef.current);
        readyTimerRef.current = null;
      }
    };
  }, [transition]);

  const wsUrl = typeof window !== 'undefined'
    ? `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`
    : '';

  const { state, connect, send, disconnect } = useWebSocket(wsUrl, handleMessage);

  const handleStart = () => {
    initAudioContext();
    connect();
    setStarted(true);
    document.body.classList.add('session-active');
  };

  const handleStop = () => {
    disconnect();
    cleanupAudio();
    setStarted(false);
    document.body.classList.remove('session-active');
    setPreviewSrc(null);
    setFinalSrc(null);
    setTransition('idle');
    setTranscript('');
    setOnboardingStage('welcome');
    setVideos([]);
    setPersonCrops([]);
    setHighlights([]);
    setAnalysisStep('');
    setAnalysisPercent(0);
    setBgmUrl(null);
  };

  const handleSelectVideo = useCallback((videoId: string) => {
    send({ type: 'select_video', videoId });
    setOnboardingStage('analyzing');
  }, [send]);

  const handleSelectPerson = useCallback((personIndex: number) => {
    send({ type: 'select_person', personIndex });
    setOnboardingStage('analyzing');
  }, [send]);

  if (!started) {
    return (
      <main
        style={{
          minHeight: '100dvh',
          overflowY: 'auto',
          background: 'var(--color-bg)',
        }}
      >
        {/* Hero Section */}
        <section
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '100dvh',
            textAlign: 'center',
            padding: '2rem',
          }}
        >
          <h1 style={{ fontSize: 'clamp(2.5rem, 6vw, 4rem)', marginBottom: '1rem', fontWeight: 700 }}>
            missless
          </h1>
          <p style={{ fontSize: 'clamp(1rem, 2.5vw, 1.25rem)', color: 'var(--color-muted)', maxWidth: '480px', lineHeight: 1.6 }}>
            A virtual reunion with someone you miss.
            <br />
            Powered by AI voice, real-time scenes, and shared memories.
          </p>
          <button
            onClick={handleStart}
            aria-label="Begin your virtual reunion"
            style={{
              marginTop: '2.5rem',
              padding: '1rem 3rem',
              fontSize: '1.125rem',
              background: 'var(--color-primary)',
              color: 'white',
              border: 'none',
              borderRadius: '2rem',
              cursor: 'pointer',
              transition: 'opacity 0.2s',
            }}
          >
            Begin Reunion
          </button>
          <p style={{ fontSize: '0.75rem', color: 'var(--color-muted)', marginTop: '1rem', opacity: 0.7 }}>
            Gemini Live Agent Challenge 2026 &middot; Creative Storyteller
          </p>
        </section>

        {/* Features Section */}
        <section
          style={{
            padding: '4rem 2rem',
            maxWidth: '900px',
            margin: '0 auto',
          }}
        >
          <h2 style={{ fontSize: '1.5rem', textAlign: 'center', marginBottom: '3rem', fontWeight: 600 }}>
            How It Works
          </h2>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
              gap: '2rem',
            }}
          >
            {[
              {
                step: '1',
                title: 'Share a Video',
                desc: 'Paste a YouTube link of the person you miss. We analyze their voice, expressions, and personality.',
              },
              {
                step: '2',
                title: 'AI Builds a Persona',
                desc: 'Our AI creates a realistic voice persona with matched speech patterns, memories, and emotions.',
              },
              {
                step: '3',
                title: 'Live Voice Reunion',
                desc: 'Have a real-time voice conversation. AI generates scenes and background music as you talk.',
              },
            ].map((f) => (
              <div
                key={f.step}
                style={{
                  background: 'var(--color-surface)',
                  borderRadius: '1rem',
                  padding: '1.5rem',
                  textAlign: 'center',
                }}
              >
                <div
                  style={{
                    width: 48,
                    height: 48,
                    borderRadius: '50%',
                    background: 'var(--color-primary)',
                    color: 'white',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '1.25rem',
                    fontWeight: 700,
                    margin: '0 auto 1rem',
                  }}
                >
                  {f.step}
                </div>
                <h3 style={{ fontSize: '1.125rem', marginBottom: '0.5rem', fontWeight: 600 }}>
                  {f.title}
                </h3>
                <p style={{ fontSize: '0.875rem', color: 'var(--color-muted)', lineHeight: 1.5, margin: 0 }}>
                  {f.desc}
                </p>
              </div>
            ))}
          </div>
        </section>

        {/* Highlights Section */}
        <section
          style={{
            padding: '3rem 2rem 4rem',
            maxWidth: '700px',
            margin: '0 auto',
            textAlign: 'center',
          }}
        >
          <h2 style={{ fontSize: '1.5rem', marginBottom: '2rem', fontWeight: 600 }}>
            Key Features
          </h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem', textAlign: 'left' }}>
            {[
              { label: 'Voice Reunion', detail: 'Real-time AI voice conversation with personality-matched personas' },
              { label: 'Scene Generation', detail: 'Progressive image generation — flash preview in seconds, high-quality final render' },
              { label: 'Memory Album', detail: 'Save and share reunion moments as a beautiful photo album' },
              { label: 'Background Music', detail: 'AI-selected atmospheric music that matches the mood of your conversation' },
            ].map((item) => (
              <div
                key={item.label}
                style={{
                  display: 'flex',
                  gap: '1rem',
                  alignItems: 'baseline',
                }}
              >
                <span style={{ color: 'var(--color-primary)', fontWeight: 600, minWidth: '160px', fontSize: '0.9375rem' }}>
                  {item.label}
                </span>
                <span style={{ color: 'var(--color-muted)', fontSize: '0.875rem', lineHeight: 1.5 }}>
                  {item.detail}
                </span>
              </div>
            ))}
          </div>
        </section>

        {/* Bottom CTA */}
        <section
          style={{
            padding: '3rem 2rem 5rem',
            textAlign: 'center',
          }}
        >
          <p style={{ fontSize: '1.125rem', color: 'var(--color-text)', marginBottom: '1.5rem' }}>
            Ready to reconnect?
          </p>
          <button
            onClick={handleStart}
            aria-label="Start your virtual reunion"
            style={{
              padding: '1rem 3rem',
              fontSize: '1.125rem',
              background: 'var(--color-primary)',
              color: 'white',
              border: 'none',
              borderRadius: '2rem',
              cursor: 'pointer',
              transition: 'opacity 0.2s',
            }}
          >
            Start Now
          </button>
        </section>
      </main>
    );
  }

  return (
    <main style={{ position: 'relative', height: '100dvh', width: '100dvw' }}>
      <SceneDisplay previewSrc={previewSrc} finalSrc={finalSrc} />
      <BGMPlayer bgmUrl={bgmUrl} />
      <SessionTransition phase={transition} />
      <OnboardingFlow
        stage={onboardingStage}
        videos={videos}
        personCrops={personCrops}
        highlights={highlights}
        analysisStep={analysisStep}
        analysisPercent={analysisPercent}
        onSelectVideo={handleSelectVideo}
        onSelectPerson={handleSelectPerson}
      />

      {/* Connection indicator */}
      <div
        style={{
          position: 'absolute',
          top: '1rem',
          right: '1rem',
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          zIndex: 10,
        }}
      >
        <div
          style={{
            width: 8,
            height: 8,
            borderRadius: '50%',
            background: CONNECTION_COLORS[state] ?? '#ef4444',
          }}
        />
        <span style={{ fontSize: '0.75rem', color: 'var(--color-muted)' }}>
          {state}
        </span>
      </div>

      {/* Transcript overlay */}
      {transcript && (
        <div
          style={{
            position: 'absolute',
            bottom: '6rem',
            left: '50%',
            transform: 'translateX(-50%)',
            maxWidth: '80%',
            padding: '0.75rem 1.5rem',
            background: 'rgba(0,0,0,0.6)',
            borderRadius: '1rem',
            color: 'var(--color-text)',
            fontSize: '1rem',
            textAlign: 'center',
            zIndex: 10,
          }}
        >
          {transcript}
        </div>
      )}

      {/* Stop button */}
      <button
        onClick={handleStop}
        style={{
          position: 'absolute',
          bottom: '2rem',
          left: '50%',
          transform: 'translateX(-50%)',
          padding: '0.75rem 2rem',
          fontSize: '1rem',
          background: 'rgba(255,255,255,0.1)',
          color: 'var(--color-text)',
          border: '1px solid rgba(255,255,255,0.2)',
          borderRadius: '2rem',
          cursor: 'pointer',
          zIndex: 10,
        }}
      >
        End Session
      </button>
    </main>
  );
}
