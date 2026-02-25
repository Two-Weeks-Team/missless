import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import ProgressBar from '../../components/ProgressBar';

describe('ProgressBar', () => {
  it('renders step text', () => {
    render(<ProgressBar step="Analyzing video..." percent={50} />);
    expect(screen.getByText('Analyzing video...')).toBeInTheDocument();
  });

  it('renders percent text', () => {
    render(<ProgressBar step="Loading" percent={42} />);
    expect(screen.getByText('42%')).toBeInTheDocument();
  });

  it('clamps width to 0% for negative values', () => {
    const { container } = render(<ProgressBar step="Step" percent={-20} />);
    // The inner bar is the child div inside the track div
    const track = container.querySelectorAll('div > div > div');
    // track[0] is the bar track, its child is the fill bar
    const fillBar = track[0]?.querySelector('div');
    expect(fillBar).toBeTruthy();
    expect(fillBar!.style.width).toBe('0%');
  });

  it('clamps width to 100% for values greater than 100', () => {
    const { container } = render(<ProgressBar step="Step" percent={150} />);
    const track = container.querySelectorAll('div > div > div');
    const fillBar = track[0]?.querySelector('div');
    expect(fillBar).toBeTruthy();
    expect(fillBar!.style.width).toBe('100%');
  });

  it('shows correct width for normal values', () => {
    const { container } = render(<ProgressBar step="Step" percent={65} />);
    const track = container.querySelectorAll('div > div > div');
    const fillBar = track[0]?.querySelector('div');
    expect(fillBar).toBeTruthy();
    expect(fillBar!.style.width).toBe('65%');
  });
});
