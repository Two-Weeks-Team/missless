import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import SessionTransition from '../../components/SessionTransition';

describe('SessionTransition', () => {
  it('returns null when phase is idle', () => {
    const { container } = render(<SessionTransition phase="idle" />);
    expect(container.innerHTML).toBe('');
  });

  it('shows "Please wait..." when transitioning', () => {
    render(<SessionTransition phase="transitioning" />);
    expect(screen.getByText('Please wait...')).toBeInTheDocument();
  });

  it('shows "Ready" when ready', () => {
    render(<SessionTransition phase="ready" />);
    expect(screen.getByText('Ready')).toBeInTheDocument();
  });
});
