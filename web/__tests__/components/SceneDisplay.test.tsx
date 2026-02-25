import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import SceneDisplay from '../../components/SceneDisplay';

describe('SceneDisplay', () => {
  it('returns null when both previewSrc and finalSrc are null', () => {
    const { container } = render(<SceneDisplay previewSrc={null} finalSrc={null} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders preview image when previewSrc provided', () => {
    render(<SceneDisplay previewSrc="data:image/png;base64,preview" finalSrc={null} />);
    const img = screen.getByAltText('Scene preview');
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute('src', 'data:image/png;base64,preview');
  });

  it('renders final image element when finalSrc provided', () => {
    render(<SceneDisplay previewSrc={null} finalSrc="data:image/png;base64,final" />);
    const img = screen.getByAltText('Scene final');
    expect(img).toBeInTheDocument();
  });
});
