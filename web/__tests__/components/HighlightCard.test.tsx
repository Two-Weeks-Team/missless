import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import HighlightCard, { type Highlight } from '../../components/HighlightCard';

describe('HighlightCard', () => {
  it('returns null for empty highlights', () => {
    const { container } = render(<HighlightCard highlights={[]} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders all highlight items', () => {
    const highlights: Highlight[] = [
      { timestamp: '0:10', description: 'First moment', expression: 'happy' },
      { timestamp: '1:30', description: 'Second moment', expression: 'excited' },
      { timestamp: '3:00', description: 'Third moment', expression: 'nostalgic' },
    ];
    render(<HighlightCard highlights={highlights} />);

    expect(screen.getByText('First moment')).toBeInTheDocument();
    expect(screen.getByText('Second moment')).toBeInTheDocument();
    expect(screen.getByText('Third moment')).toBeInTheDocument();
  });

  it('shows timestamp, description, and expression for each item', () => {
    const highlights: Highlight[] = [
      { timestamp: '2:45', description: 'A key scene', expression: 'surprised' },
    ];
    render(<HighlightCard highlights={highlights} />);

    expect(screen.getByText('2:45')).toBeInTheDocument();
    expect(screen.getByText('A key scene')).toBeInTheDocument();
    expect(screen.getByText('surprised')).toBeInTheDocument();
  });
});
