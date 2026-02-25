'use client';

import { useState, useCallback } from 'react';
import ProgressBar from './ProgressBar';
import HighlightCard, { type Highlight } from './HighlightCard';
import YouTubeGrid, { type YouTubeVideo } from './YouTubeGrid';
import PrivateVideoPopup from './PrivateVideoPopup';

export type OnboardingStage =
  | 'welcome'
  | 'youtube_grid'
  | 'person_select'
  | 'analyzing'
  | 'transition'
  | 'reunion';

interface OnboardingFlowProps {
  stage: OnboardingStage;
  videos: YouTubeVideo[];
  personCrops: string[];
  highlights: Highlight[];
  analysisStep: string;
  analysisPercent: number;
  onSelectVideo: (videoId: string) => void;
  onSelectPerson: (index: number) => void;
}

export default function OnboardingFlow({
  stage,
  videos,
  personCrops,
  highlights,
  analysisStep,
  analysisPercent,
  onSelectVideo,
  onSelectPerson,
}: OnboardingFlowProps) {
  const [showPrivatePopup, setShowPrivatePopup] = useState(false);

  const handlePrivateClick = useCallback(() => {
    setShowPrivatePopup(true);
  }, []);

  return (
    <>
      {/* YouTube video selection grid */}
      {stage === 'youtube_grid' && (
        <YouTubeGrid
          videos={videos}
          onSelect={onSelectVideo}
          onPrivateClick={handlePrivateClick}
        />
      )}

      {/* Person crop selection grid */}
      {stage === 'person_select' && personCrops.length > 0 && (
        <div
          style={{
            position: 'absolute',
            inset: 0,
            background: 'var(--color-bg)',
            zIndex: 30,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '2rem',
          }}
        >
          <h2
            style={{
              fontSize: '1.25rem',
              marginBottom: '1.5rem',
              color: 'var(--color-text)',
            }}
          >
            분석할 인물을 선택하세요
          </h2>
          <div
            style={{
              display: 'flex',
              gap: '1rem',
              flexWrap: 'wrap',
              justifyContent: 'center',
            }}
          >
            {personCrops.map((crop, i) => (
              <button
                key={crop}
                className="person-crop-btn"
                onClick={() => onSelectPerson(i)}
                style={{
                  width: 100,
                  height: 100,
                  borderRadius: '50%',
                  overflow: 'hidden',
                  border: '3px solid transparent',
                  cursor: 'pointer',
                  padding: 0,
                  background: 'var(--color-surface)',
                  transition: 'border-color 0.2s',
                }}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={crop}
                  alt={`Person ${i + 1}`}
                  style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                />
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Analysis progress and highlights */}
      {stage === 'analyzing' && (
        <>
          <ProgressBar step={analysisStep} percent={analysisPercent} />
          <HighlightCard highlights={highlights} />
        </>
      )}

      {/* Private video popup */}
      {showPrivatePopup && (
        <PrivateVideoPopup onClose={() => setShowPrivatePopup(false)} />
      )}
    </>
  );
}
