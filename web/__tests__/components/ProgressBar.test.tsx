import { describe, it, expect } from 'vitest';
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
    render(<ProgressBar step="Step" percent={-20} />);
    const fillBar = screen.getByTestId('progress-fill');
    expect(fillBar).toBeInTheDocument();
    expect(fillBar).toHaveStyle({ width: '0%' });
  });

  it('clamps width to 100% for values greater than 100', () => {
    render(<ProgressBar step="Step" percent={150} />);
    const fillBar = screen.getByTestId('progress-fill');
    expect(fillBar).toHaveStyle({ width: '100%' });
  });

  it('shows correct width for normal values', () => {
    render(<ProgressBar step="Step" percent={65} />);
    const fillBar = screen.getByTestId('progress-fill');
    expect(fillBar).toHaveStyle({ width: '65%' });
  });
});
