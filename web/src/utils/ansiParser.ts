/**
 * ANSI Parser - Converts ANSI escape codes to styled React elements
 *
 * Supports:
 * - Foreground colors (30-37, 90-97)
 * - Background colors (40-47, 100-107)
 * - Text styles: bold, dim, italic, underline
 * - Reset codes
 */

export interface AnsiElement {
  text: string
  styles: React.CSSProperties
}

// Color mapping for ANSI codes
const ANSI_COLORS: Record<number, string> = {
  // Standard colors (30-37)
  30: '#000000', // black
  31: '#ef4444', // red
  32: '#22c55e', // green
  33: '#eab308', // yellow
  34: '#3b82f6', // blue
  35: '#a855f7', // magenta
  36: '#06b6d4', // cyan
  37: '#f5f5f5', // white

  // Bright colors (90-97)
  90: '#404040', // bright black (gray)
  91: '#f87171', // bright red
  92: '#4ade80', // bright green
  93: '#fde047', // bright yellow
  94: '#60a5fa', // bright blue
  95: '#c084fc', // bright magenta
  96: '#22d3ee', // bright cyan
  97: '#ffffff', // bright white
}

interface StyleState {
  color?: string
  backgroundColor?: string
  fontWeight?: string
  opacity?: string
  fontStyle?: string
  textDecoration?: string
}

/**
 * Parse a single ANSI code and update style state
 */
function applyAnsiCode(code: number, state: StyleState): void {
  // Reset
  if (code === 0) {
    Object.keys(state).forEach(key => delete state[key as keyof StyleState])
    return
  }

  // Text styles
  if (code === 1) {
    state.fontWeight = 'bold'
    return
  }
  if (code === 2) {
    state.opacity = '0.6'
    return
  }
  if (code === 3) {
    state.fontStyle = 'italic'
    return
  }
  if (code === 4) {
    state.textDecoration = 'underline'
    return
  }

  // Foreground colors (30-37, 90-97)
  if ((code >= 30 && code <= 37) || (code >= 90 && code <= 97)) {
    state.color = ANSI_COLORS[code]
    return
  }

  // Background colors (40-47, 100-107)
  if ((code >= 40 && code <= 47) || (code >= 100 && code <= 107)) {
    // Map background codes to foreground equivalents
    // 40-47 -> 30-37 (standard)
    // 100-107 -> 90-97 (bright)
    const colorCode = code >= 100 ? code - 10 : code - 10
    state.backgroundColor = ANSI_COLORS[colorCode]
    return
  }
}

/**
 * Parse ANSI escape codes and convert to styled elements
 *
 * @param text - Text containing ANSI escape codes
 * @returns Array of elements with text and styles
 */
export function parseAnsiToElements(text: string): AnsiElement[] {
  if (!text) {
    return []
  }

  const elements: AnsiElement[] = []
  const currentState: StyleState = {}

  // Regex to match ANSI escape codes: \x1b[...m
  const ansiRegex = /\x1b\[([0-9;]*)m/g
  let lastIndex = 0
  let match: RegExpExecArray | null

  while ((match = ansiRegex.exec(text)) !== null) {
    const textBefore = text.slice(lastIndex, match.index)

    // Add text before ANSI code if present
    if (textBefore) {
      elements.push({
        text: textBefore,
        styles: { ...currentState }
      })
    }

    // Apply the ANSI codes to current state
    processAnsiCodes(match[1], currentState)
    lastIndex = ansiRegex.lastIndex
  }

  // Add remaining text after last ANSI code
  const remainingText = text.slice(lastIndex)
  if (remainingText) {
    elements.push({
      text: remainingText,
      styles: { ...currentState }
    })
  }

  return elements
}

/**
 * Process ANSI codes and update style state
 */
function processAnsiCodes(codes: string, state: StyleState): void {
  if (!codes) {
    resetStyles(state)
    return
  }

  const codeNumbers = codes.split(';').map(c => parseInt(c, 10)).filter(n => !isNaN(n))
  codeNumbers.forEach(code => applyAnsiCode(code, state))
}

/**
 * Reset all styles in the state
 */
function resetStyles(state: StyleState): void {
  Object.keys(state).forEach(key => delete state[key as keyof StyleState])
}
