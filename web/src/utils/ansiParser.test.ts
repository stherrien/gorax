import { describe, it, expect } from 'vitest'
import { parseAnsiToElements, type AnsiElement } from './ansiParser'

describe('parseAnsiToElements', () => {
  describe('plain text', () => {
    it('should return plain text without ANSI codes', () => {
      const result = parseAnsiToElements('Hello World')
      expect(result).toEqual([
        { text: 'Hello World', styles: {} }
      ])
    })

    it('should handle empty string', () => {
      const result = parseAnsiToElements('')
      expect(result).toEqual([])
    })

    it('should handle multiline text', () => {
      const result = parseAnsiToElements('Line 1\nLine 2\nLine 3')
      expect(result).toEqual([
        { text: 'Line 1\nLine 2\nLine 3', styles: {} }
      ])
    })
  })

  describe('foreground colors', () => {
    it('should parse red text (31m)', () => {
      const result = parseAnsiToElements('\x1b[31mRed Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Red Text', styles: { color: '#ef4444' } }
      ])
    })

    it('should parse green text (32m)', () => {
      const result = parseAnsiToElements('\x1b[32mGreen Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Green Text', styles: { color: '#22c55e' } }
      ])
    })

    it('should parse yellow text (33m)', () => {
      const result = parseAnsiToElements('\x1b[33mYellow Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Yellow Text', styles: { color: '#eab308' } }
      ])
    })

    it('should parse blue text (34m)', () => {
      const result = parseAnsiToElements('\x1b[34mBlue Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Blue Text', styles: { color: '#3b82f6' } }
      ])
    })

    it('should parse magenta text (35m)', () => {
      const result = parseAnsiToElements('\x1b[35mMagenta Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Magenta Text', styles: { color: '#a855f7' } }
      ])
    })

    it('should parse cyan text (36m)', () => {
      const result = parseAnsiToElements('\x1b[36mCyan Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Cyan Text', styles: { color: '#06b6d4' } }
      ])
    })

    it('should parse white text (37m)', () => {
      const result = parseAnsiToElements('\x1b[37mWhite Text\x1b[0m')
      expect(result).toEqual([
        { text: 'White Text', styles: { color: '#f5f5f5' } }
      ])
    })

    it('should parse bright colors (90-97m)', () => {
      const result = parseAnsiToElements('\x1b[91mBright Red\x1b[0m')
      expect(result).toEqual([
        { text: 'Bright Red', styles: { color: '#f87171' } }
      ])
    })
  })

  describe('background colors', () => {
    it('should parse red background (41m)', () => {
      const result = parseAnsiToElements('\x1b[41mRed BG\x1b[0m')
      expect(result).toEqual([
        { text: 'Red BG', styles: { backgroundColor: '#ef4444' } }
      ])
    })

    it('should parse green background (42m)', () => {
      const result = parseAnsiToElements('\x1b[42mGreen BG\x1b[0m')
      expect(result).toEqual([
        { text: 'Green BG', styles: { backgroundColor: '#22c55e' } }
      ])
    })

    it('should parse bright backgrounds (100-107m)', () => {
      const result = parseAnsiToElements('\x1b[101mBright Red BG\x1b[0m')
      expect(result).toEqual([
        { text: 'Bright Red BG', styles: { backgroundColor: '#f87171' } }
      ])
    })
  })

  describe('text styles', () => {
    it('should parse bold text (1m)', () => {
      const result = parseAnsiToElements('\x1b[1mBold Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Bold Text', styles: { fontWeight: 'bold' } }
      ])
    })

    it('should parse dim text (2m)', () => {
      const result = parseAnsiToElements('\x1b[2mDim Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Dim Text', styles: { opacity: '0.6' } }
      ])
    })

    it('should parse italic text (3m)', () => {
      const result = parseAnsiToElements('\x1b[3mItalic Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Italic Text', styles: { fontStyle: 'italic' } }
      ])
    })

    it('should parse underlined text (4m)', () => {
      const result = parseAnsiToElements('\x1b[4mUnderlined Text\x1b[0m')
      expect(result).toEqual([
        { text: 'Underlined Text', styles: { textDecoration: 'underline' } }
      ])
    })
  })

  describe('combined styles', () => {
    it('should handle multiple styles in one code (31;1m)', () => {
      const result = parseAnsiToElements('\x1b[31;1mBold Red\x1b[0m')
      expect(result).toEqual([
        { text: 'Bold Red', styles: { color: '#ef4444', fontWeight: 'bold' } }
      ])
    })

    it('should handle foreground and background (31;42m)', () => {
      const result = parseAnsiToElements('\x1b[31;42mRed on Green\x1b[0m')
      expect(result).toEqual([
        { text: 'Red on Green', styles: { color: '#ef4444', backgroundColor: '#22c55e' } }
      ])
    })

    it('should handle nested style changes', () => {
      const result = parseAnsiToElements('\x1b[31mRed \x1b[1mBold Red\x1b[0m')
      expect(result).toEqual([
        { text: 'Red ', styles: { color: '#ef4444' } },
        { text: 'Bold Red', styles: { color: '#ef4444', fontWeight: 'bold' } }
      ])
    })
  })

  describe('reset codes', () => {
    it('should reset all styles with 0m', () => {
      const result = parseAnsiToElements('\x1b[31;1mRed Bold\x1b[0m Normal')
      expect(result).toEqual([
        { text: 'Red Bold', styles: { color: '#ef4444', fontWeight: 'bold' } },
        { text: ' Normal', styles: {} }
      ])
    })

    it('should handle multiple resets', () => {
      const result = parseAnsiToElements('\x1b[31mRed\x1b[0m \x1b[32mGreen\x1b[0m')
      expect(result).toEqual([
        { text: 'Red', styles: { color: '#ef4444' } },
        { text: ' ', styles: {} },
        { text: 'Green', styles: { color: '#22c55e' } }
      ])
    })
  })

  describe('edge cases', () => {
    it('should handle invalid ANSI codes gracefully', () => {
      const result = parseAnsiToElements('\x1b[999mInvalid\x1b[0m')
      expect(result).toEqual([
        { text: 'Invalid', styles: {} }
      ])
    })

    it('should handle incomplete ANSI codes', () => {
      const result = parseAnsiToElements('\x1b[31mNo Reset')
      expect(result).toEqual([
        { text: 'No Reset', styles: { color: '#ef4444' } }
      ])
    })

    it('should handle empty ANSI codes', () => {
      const result = parseAnsiToElements('\x1b[mEmpty Code\x1b[0m')
      expect(result).toEqual([
        { text: 'Empty Code', styles: {} }
      ])
    })

    it('should preserve whitespace', () => {
      const result = parseAnsiToElements('  \x1b[31mIndented\x1b[0m  ')
      expect(result).toEqual([
        { text: '  ', styles: {} },
        { text: 'Indented', styles: { color: '#ef4444' } },
        { text: '  ', styles: {} }
      ])
    })

    it('should handle consecutive ANSI codes', () => {
      const result = parseAnsiToElements('\x1b[31m\x1b[1mRed Bold\x1b[0m')
      expect(result).toEqual([
        { text: 'Red Bold', styles: { color: '#ef4444', fontWeight: 'bold' } }
      ])
    })
  })

  describe('real-world examples', () => {
    it('should parse Docker-style logs', () => {
      const result = parseAnsiToElements('[2024-01-01 12:00:00] \x1b[32mINFO\x1b[0m Starting server...')
      expect(result).toEqual([
        { text: '[2024-01-01 12:00:00] ', styles: {} },
        { text: 'INFO', styles: { color: '#22c55e' } },
        { text: ' Starting server...', styles: {} }
      ])
    })

    it('should parse error logs with multiple colors', () => {
      const result = parseAnsiToElements('\x1b[31mERROR\x1b[0m: \x1b[33mWarning\x1b[0m occurred at line \x1b[36m42\x1b[0m')
      expect(result).toEqual([
        { text: 'ERROR', styles: { color: '#ef4444' } },
        { text: ': ', styles: {} },
        { text: 'Warning', styles: { color: '#eab308' } },
        { text: ' occurred at line ', styles: {} },
        { text: '42', styles: { color: '#06b6d4' } }
      ])
    })

    it('should handle kubectl-style output', () => {
      const result = parseAnsiToElements('NAME\t\t\x1b[32mREADY\x1b[0m\t\x1b[36mSTATUS\x1b[0m')
      expect(result).toEqual([
        { text: 'NAME\t\t', styles: {} },
        { text: 'READY', styles: { color: '#22c55e' } },
        { text: '\t', styles: {} },
        { text: 'STATUS', styles: { color: '#06b6d4' } }
      ])
    })
  })
})
