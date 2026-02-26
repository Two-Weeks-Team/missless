/**
 * Strip common markdown formatting from text for clean voice-transcript display.
 * Handles: bold, italic, headers, links, code blocks, bullet lists, etc.
 */
export function stripMarkdown(text: string): string {
  return text
    // Remove code blocks (``` ... ```)
    .replace(/```[\s\S]*?```/g, '')
    // Remove inline code (`...`)
    .replace(/`([^`]+)`/g, '$1')
    // Remove bold+italic (***text*** or ___text___)
    .replace(/(\*{3}|_{3})(.*?)\1/g, '$2')
    // Remove bold (**text** or __text__)
    .replace(/(\*{2}|_{2})(.*?)\1/g, '$2')
    // Remove italic (*text* or _text_)
    .replace(/(\*|_)(.*?)\1/g, '$2')
    // Remove headers (# ... ##)
    .replace(/^#{1,6}\s+/gm, '')
    // Remove links [text](url) → text
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    // Remove images ![alt](url)
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    // Remove bullet list markers
    .replace(/^[\s]*[-*+]\s+/gm, '')
    // Remove numbered list markers
    .replace(/^[\s]*\d+\.\s+/gm, '')
    // Remove horizontal rules
    .replace(/^[-*_]{3,}\s*$/gm, '')
    // Remove blockquotes
    .replace(/^>\s+/gm, '')
    // Collapse multiple spaces/newlines
    .replace(/\n{3,}/g, '\n\n')
    .trim();
}
