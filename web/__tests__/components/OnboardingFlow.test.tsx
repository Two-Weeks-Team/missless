import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import OnboardingFlow from '../../components/OnboardingFlow';
import type { YouTubeVideo } from '../../components/YouTubeGrid';
import type { Highlight } from '../../components/HighlightCard';

const sampleVideos: YouTubeVideo[] = [
  { id: 'v1', title: 'Video One', thumbnail: 'https://example.com/1.jpg', privacyStatus: 'public' },
  { id: 'v2', title: 'Video Two', thumbnail: 'https://example.com/2.jpg', privacyStatus: 'public' },
];

const sampleHighlights: Highlight[] = [
  { timestamp: '0:15', description: 'Laughing together', expression: 'joyful' },
];

const sampleCrops = [
  'https://example.com/crop1.jpg',
  'https://example.com/crop2.jpg',
];

const makeDefaultProps = () => ({
  stage: 'welcome' as const,
  videos: sampleVideos,
  personCrops: sampleCrops,
  highlights: sampleHighlights,
  analysisStep: 'Extracting features...',
  analysisPercent: 30,
  onSelectVideo: vi.fn(),
  onSelectPerson: vi.fn(),
});

let defaultProps: ReturnType<typeof makeDefaultProps>;

beforeEach(() => {
  defaultProps = makeDefaultProps();
});

describe('OnboardingFlow', () => {
  it('renders YouTubeGrid when stage is youtube_grid', () => {
    render(<OnboardingFlow {...defaultProps} stage="youtube_grid" />);
    expect(screen.getByText('Select a video')).toBeInTheDocument();
    expect(screen.getByText('Video One')).toBeInTheDocument();
    expect(screen.getByText('Video Two')).toBeInTheDocument();
  });

  it('renders person select UI when stage is person_select with crops', () => {
    render(<OnboardingFlow {...defaultProps} stage="person_select" />);
    expect(screen.getByText('Select the person to analyze')).toBeInTheDocument();
    expect(screen.getByAltText('Person 1')).toBeInTheDocument();
    expect(screen.getByAltText('Person 2')).toBeInTheDocument();
  });

  it('does not render person select when personCrops is empty', () => {
    render(<OnboardingFlow {...defaultProps} stage="person_select" personCrops={[]} />);
    expect(screen.queryByText('Select the person to analyze')).not.toBeInTheDocument();
  });

  it('renders ProgressBar when stage is analyzing', () => {
    render(<OnboardingFlow {...defaultProps} stage="analyzing" />);
    expect(screen.getByText('Extracting features...')).toBeInTheDocument();
    expect(screen.getByText('30%')).toBeInTheDocument();
  });

  it('calls onSelectPerson when person button clicked', () => {
    const onSelectPerson = vi.fn();
    render(
      <OnboardingFlow {...defaultProps} stage="person_select" onSelectPerson={onSelectPerson} />
    );
    fireEvent.click(screen.getByAltText('Person 2'));
    expect(onSelectPerson).toHaveBeenCalledWith(1);
  });
});
