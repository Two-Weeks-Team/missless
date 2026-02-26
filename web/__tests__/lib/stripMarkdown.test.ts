import { describe, it, expect } from 'vitest';
import { stripMarkdown } from '../../lib/stripMarkdown';

describe('stripMarkdown', () => {
  it('strips bold markers', () => {
    expect(stripMarkdown('**hello** world')).toBe('hello world');
  });

  it('strips italic markers', () => {
    expect(stripMarkdown('*hello* world')).toBe('hello world');
  });

  it('strips bold+italic markers', () => {
    expect(stripMarkdown('***hello*** world')).toBe('hello world');
  });

  it('strips headers', () => {
    expect(stripMarkdown('## Hello\nWorld')).toBe('Hello\nWorld');
  });

  it('strips inline code', () => {
    expect(stripMarkdown('use `npm install`')).toBe('use npm install');
  });

  it('strips links preserving text', () => {
    expect(stripMarkdown('[click here](https://example.com)')).toBe('click here');
  });

  it('strips bullet list markers', () => {
    expect(stripMarkdown('- item one\n- item two')).toBe('item one\nitem two');
  });

  it('strips blockquotes', () => {
    expect(stripMarkdown('> quoted text')).toBe('quoted text');
  });

  it('returns plain text unchanged', () => {
    expect(stripMarkdown('Hello there!')).toBe('Hello there!');
  });

  it('handles the exact model output from screenshot', () => {
    const input = '**Initiating Welcome Sequence** I\'ve just received the user\'s connection.';
    const output = stripMarkdown(input);
    expect(output).not.toContain('**');
    expect(output).toContain('Initiating Welcome Sequence');
  });
});
