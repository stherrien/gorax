import { parseAnsiToElements } from '../../utils/ansiParser'

interface AnsiTextProps {
  text: string
  inline?: boolean
  className?: string
}

/**
 * AnsiText - Renders text with ANSI escape codes as styled elements
 *
 * Features:
 * - Parses ANSI color codes (foreground and background)
 * - Supports text styles (bold, italic, underline, dim)
 * - Preserves whitespace and line breaks
 * - Monospace font for code-like appearance
 * - Can render as block (pre) or inline (code)
 *
 * @example
 * <AnsiText text="\x1b[31mError:\x1b[0m Connection failed" />
 * <AnsiText text="\x1b[32mSUCCESS\x1b[0m" inline />
 */
export function AnsiText({ text, inline = false, className = '' }: AnsiTextProps) {
  if (!text) {
    return inline ? <code className="ansi-text-inline" /> : <pre className="ansi-text" />
  }

  const elements = parseAnsiToElements(text)

  const renderedElements = elements.map((element, index) => (
    <span key={index} style={element.styles}>
      {element.text}
    </span>
  ))

  if (inline) {
    return (
      <code className={`ansi-text-inline ${className}`}>
        {renderedElements}
      </code>
    )
  }

  return (
    <pre
      className={`ansi-text ${className}`}
      role="code"
      tabIndex={0}
      style={{
        fontFamily: 'monospace',
        whiteSpace: 'pre-wrap',
        wordBreak: 'break-word',
        margin: 0,
      }}
    >
      {renderedElements}
    </pre>
  )
}
