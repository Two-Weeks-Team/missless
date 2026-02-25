import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import YouTubeGrid, { type YouTubeVideo } from '../../components/YouTubeGrid';

const publicVideo: YouTubeVideo = {
  id: 'vid-1',
  title: 'Public Video Title',
  thumbnail: 'https://example.com/thumb1.jpg',
  privacyStatus: 'public',
};

const privateVideo: YouTubeVideo = {
  id: 'vid-2',
  title: 'Private Video Title',
  thumbnail: 'https://example.com/thumb2.jpg',
  privacyStatus: 'private',
};

const unlistedVideo: YouTubeVideo = {
  id: 'vid-3',
  title: 'Unlisted Video Title',
  thumbnail: 'https://example.com/thumb3.jpg',
  privacyStatus: 'unlisted',
};

describe('YouTubeGrid', () => {
  it('returns null for empty videos', () => {
    const { container } = render(
      <YouTubeGrid videos={[]} onSelect={vi.fn()} onPrivateClick={vi.fn()} />
    );
    expect(container.innerHTML).toBe('');
  });

  it('renders all video items with titles', () => {
    const videos = [publicVideo, privateVideo, unlistedVideo];
    render(
      <YouTubeGrid videos={videos} onSelect={vi.fn()} onPrivateClick={vi.fn()} />
    );

    expect(screen.getByText('Public Video Title')).toBeInTheDocument();
    expect(screen.getByText('Private Video Title')).toBeInTheDocument();
    expect(screen.getByText('Unlisted Video Title')).toBeInTheDocument();
  });

  it('clicking public video calls onSelect with video id', () => {
    const onSelect = vi.fn();
    const onPrivateClick = vi.fn();
    render(
      <YouTubeGrid videos={[publicVideo]} onSelect={onSelect} onPrivateClick={onPrivateClick} />
    );

    fireEvent.click(screen.getByText('Public Video Title'));
    expect(onSelect).toHaveBeenCalledWith('vid-1');
    expect(onPrivateClick).not.toHaveBeenCalled();
  });

  it('clicking private video calls onPrivateClick', () => {
    const onSelect = vi.fn();
    const onPrivateClick = vi.fn();
    render(
      <YouTubeGrid videos={[privateVideo]} onSelect={onSelect} onPrivateClick={onPrivateClick} />
    );

    fireEvent.click(screen.getByText('Private Video Title'));
    expect(onPrivateClick).toHaveBeenCalledTimes(1);
    expect(onSelect).not.toHaveBeenCalled();
  });

  it('shows "Private" badge for non-public videos', () => {
    const videos = [publicVideo, privateVideo, unlistedVideo];
    render(
      <YouTubeGrid videos={videos} onSelect={vi.fn()} onPrivateClick={vi.fn()} />
    );

    const privateBadges = screen.getAllByText('Private');
    // privateVideo and unlistedVideo both get the badge
    expect(privateBadges).toHaveLength(2);
  });
});
