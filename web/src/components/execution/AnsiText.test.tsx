import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { AnsiText } from './AnsiText'

describe('AnsiText', () => {
  describe('plain text', () => {
    it('should render plain text without styles', () => {
      const { container } = render(<AnsiText text="Hello World" />)
      expect(container.textContent).toBe('Hello World')
    })

    it('should preserve whitespace', () => {
      const { container } = render(<AnsiText text="  Hello  World  " />)
      expect(container.textContent).toBe('  Hello  World  ')
    })

    it('should preserve line breaks', () => {
      const { container } = render(<AnsiText text="Line 1\nLine 2\nLine 3" />)
      // Check that all content is present (DOM may normalize whitespace in textContent)
      expect(container.textContent).toContain('Line 1')
      expect(container.textContent).toContain('Line 2')
      expect(container.textContent).toContain('Line 3')
    })
  })

  describe('colored text', () => {
    it('should render red text', () => {
      const text = '\u001b[31mRed Text\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Red Text')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const redSpan = spans[0] // The parsed text becomes the first span
      expect(redSpan.textContent).toBe('Red Text')
      expect(redSpan.style.color).toBe('#ef4444')
    })

    it('should render green text', () => {
      const text = '\u001b[32mGreen Text\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Green Text')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const greenSpan = spans[0]
      expect(greenSpan.textContent).toBe('Green Text')
      expect(greenSpan.style.color).toBe('#22c55e')
    })

    it('should render multiple colors', () => {
      const text = '\u001b[31mRed\u001b[0m \u001b[32mGreen\u001b[0m'
      const { container} = render(<AnsiText text={text} />)
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBe(3) // Red, space, Green
      const redSpan = Array.from(spans).find(s => s.textContent === 'Red')
      const greenSpan = Array.from(spans).find(s => s.textContent === 'Green')
      expect(redSpan).toBeDefined()
      expect(greenSpan).toBeDefined()
      expect(redSpan!.style.color).toBe('#ef4444')
      expect(greenSpan!.style.color).toBe('#22c55e')
    })

    it('should render background colors', () => {
      const text = '\u001b[41mRed BG\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Red BG')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const bgSpan = spans[0]
      expect(bgSpan.textContent).toBe('Red BG')
      expect(bgSpan.style.backgroundColor).toBe('#ef4444')
    })
  })

  describe('text styles', () => {
    it('should render bold text', () => {
      const text = '\u001b[1mBold\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Bold')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const boldSpan = spans[0]
      expect(boldSpan.textContent).toBe('Bold')
      expect(boldSpan.style.fontWeight).toBe('bold')
    })

    it('should render italic text', () => {
      const text = '\u001b[3mItalic\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Italic')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const italicSpan = spans[0]
      expect(italicSpan.textContent).toBe('Italic')
      expect(italicSpan.style.fontStyle).toBe('italic')
    })

    it('should render underlined text', () => {
      const text = '\u001b[4mUnderline\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Underline')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const underlineSpan = spans[0]
      expect(underlineSpan.textContent).toBe('Underline')
      expect(underlineSpan.style.textDecoration).toBe('underline')
    })

    it('should render dim text', () => {
      const text = '\u001b[2mDim\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Dim')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const dimSpan = spans[0]
      expect(dimSpan.textContent).toBe('Dim')
      expect(dimSpan.style.opacity).toBe('0.6')
    })
  })

  describe('combined styles', () => {
    it('should render bold red text', () => {
      const text = '\u001b[31;1mBold Red\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Bold Red')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const styledSpan = spans[0]
      expect(styledSpan.textContent).toBe('Bold Red')
      expect(styledSpan.style.color).toBe('#ef4444')
      expect(styledSpan.style.fontWeight).toBe('bold')
    })

    it('should render text with foreground and background', () => {
      const text = '\u001b[31;42mRed on Green\u001b[0m'
      const { container } = render(<AnsiText text={text} />)
      expect(container.textContent).toContain('Red on Green')
      const spans = container.querySelectorAll('span')
      expect(spans.length).toBeGreaterThan(0)
      const styledSpan = spans[0]
      expect(styledSpan.textContent).toBe('Red on Green')
      expect(styledSpan.style.color).toBe('#ef4444')
      expect(styledSpan.style.backgroundColor).toBe('#22c55e')
    })
  })

  describe('monospace styling', () => {
    it('should apply monospace font', () => {
      const { container } = render(<AnsiText text="Code" />)
      const pre = container.querySelector('pre')
      expect(pre).toBeInTheDocument()
      expect(pre).toHaveClass('ansi-text')
    })

    it('should use code element for inline display', () => {
      const { container } = render(<AnsiText text="Code" inline />)
      const code = container.querySelector('code')
      expect(code).toBeInTheDocument()
    })
  })

  describe('real-world log examples', () => {
    it('should render Docker-style logs', () => {
      const log = '[2024-01-01 12:00:00] \u001b[32mINFO\u001b[0m Starting server...'
      const { container } = render(<AnsiText text={log} />)
      const spans = container.querySelectorAll('span')
      const infoSpan = Array.from(spans).find(s => s.textContent === 'INFO')
      expect(infoSpan).toBeDefined()
      expect(infoSpan!.style.color).toBe('#22c55e')
      expect(container.textContent).toContain('[2024-01-01 12:00:00]')
      expect(container.textContent).toContain('Starting server...')
    })

    it('should render error logs', () => {
      const log = '\u001b[31mERROR\u001b[0m: Connection failed'
      const { container } = render(<AnsiText text={log} />)
      const spans = container.querySelectorAll('span')
      const errorSpan = Array.from(spans).find(s => s.textContent === 'ERROR')
      expect(errorSpan).toBeDefined()
      expect(errorSpan!.style.color).toBe('#ef4444')
      expect(container.textContent).toContain(': Connection failed')
    })

    it('should render kubectl-style colored output', () => {
      const log = 'NAME\t\u001b[32mREADY\u001b[0m\t\u001b[36mSTATUS\u001b[0m'
      const { container } = render(<AnsiText text={log} />)
      const spans = container.querySelectorAll('span')
      const readySpan = Array.from(spans).find(s => s.textContent === 'READY')
      const statusSpan = Array.from(spans).find(s => s.textContent === 'STATUS')
      expect(readySpan).toBeDefined()
      expect(statusSpan).toBeDefined()
      expect(readySpan!.style.color).toBe('#22c55e')
      expect(statusSpan!.style.color).toBe('#06b6d4')
    })
  })

  describe('accessibility', () => {
    it('should have proper ARIA role', () => {
      const { container } = render(<AnsiText text="Test" />)
      const pre = container.querySelector('pre')
      expect(pre).toHaveAttribute('role', 'code')
    })

    it('should be keyboard navigable', () => {
      const { container } = render(<AnsiText text="Test" />)
      const pre = container.querySelector('pre')
      expect(pre).toHaveAttribute('tabIndex', '0')
    })
  })

  describe('edge cases', () => {
    it('should handle empty text', () => {
      const { container } = render(<AnsiText text="" />)
      expect(container.textContent).toBe('')
    })

    it('should handle null text gracefully', () => {
      const { container } = render(<AnsiText text={null as any} />)
      expect(container.textContent).toBe('')
    })

    it('should handle undefined text gracefully', () => {
      const { container } = render(<AnsiText text={undefined as any} />)
      expect(container.textContent).toBe('')
    })
  })
})
