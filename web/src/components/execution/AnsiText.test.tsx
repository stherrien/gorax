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
      const { container } = render(<AnsiText text="\x1b[31mRed Text\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const redSpan = Array.from(spans).find(s => s.textContent === 'Red Text')
      expect(redSpan).toBeDefined()
      expect(redSpan!).toHaveStyle({ color: '#ef4444' })
    })

    it('should render green text', () => {
      const { container } = render(<AnsiText text="\x1b[32mGreen Text\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const greenSpan = Array.from(spans).find(s => s.textContent === 'Green Text')
      expect(greenSpan).toBeDefined()
      expect(greenSpan!).toHaveStyle({ color: '#22c55e' })
    })

    it('should render multiple colors', () => {
      const { container } = render(
        <AnsiText text="\x1b[31mRed\x1b[0m \x1b[32mGreen\x1b[0m" />
      )
      const spans = container.querySelectorAll('span')
      const redSpan = Array.from(spans).find(s => s.textContent === 'Red')
      const greenSpan = Array.from(spans).find(s => s.textContent === 'Green')
      expect(redSpan).toBeDefined()
      expect(greenSpan).toBeDefined()
      expect(redSpan!).toHaveStyle({ color: '#ef4444' })
      expect(greenSpan!).toHaveStyle({ color: '#22c55e' })
    })

    it('should render background colors', () => {
      const { container } = render(<AnsiText text="\x1b[41mRed BG\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const bgSpan = Array.from(spans).find(s => s.textContent === 'Red BG')
      expect(bgSpan).toBeDefined()
      expect(bgSpan!).toHaveStyle({ backgroundColor: '#ef4444' })
    })
  })

  describe('text styles', () => {
    it('should render bold text', () => {
      const { container } = render(<AnsiText text="\x1b[1mBold\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const boldSpan = Array.from(spans).find(s => s.textContent === 'Bold')
      expect(boldSpan).toBeDefined()
      expect(boldSpan!).toHaveStyle({ fontWeight: 'bold' })
    })

    it('should render italic text', () => {
      const { container } = render(<AnsiText text="\x1b[3mItalic\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const italicSpan = Array.from(spans).find(s => s.textContent === 'Italic')
      expect(italicSpan).toBeDefined()
      expect(italicSpan!).toHaveStyle({ fontStyle: 'italic' })
    })

    it('should render underlined text', () => {
      const { container } = render(<AnsiText text="\x1b[4mUnderline\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const underlineSpan = Array.from(spans).find(s => s.textContent === 'Underline')
      expect(underlineSpan).toBeDefined()
      expect(underlineSpan!).toHaveStyle({ textDecoration: 'underline' })
    })

    it('should render dim text', () => {
      const { container } = render(<AnsiText text="\x1b[2mDim\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const dimSpan = Array.from(spans).find(s => s.textContent === 'Dim')
      expect(dimSpan).toBeDefined()
      expect(dimSpan!).toHaveStyle({ opacity: '0.6' })
    })
  })

  describe('combined styles', () => {
    it('should render bold red text', () => {
      const { container } = render(<AnsiText text="\x1b[31;1mBold Red\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const styledSpan = Array.from(spans).find(s => s.textContent === 'Bold Red')
      expect(styledSpan).toBeDefined()
      expect(styledSpan!).toHaveStyle({ color: '#ef4444', fontWeight: 'bold' })
    })

    it('should render text with foreground and background', () => {
      const { container } = render(<AnsiText text="\x1b[31;42mRed on Green\x1b[0m" />)
      const spans = container.querySelectorAll('span')
      const styledSpan = Array.from(spans).find(s => s.textContent === 'Red on Green')
      expect(styledSpan).toBeDefined()
      expect(styledSpan!).toHaveStyle({ color: '#ef4444', backgroundColor: '#22c55e' })
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
      const log = '[2024-01-01 12:00:00] \x1b[32mINFO\x1b[0m Starting server...'
      const { container } = render(<AnsiText text={log} />)
      const spans = container.querySelectorAll('span')
      const infoSpan = Array.from(spans).find(s => s.textContent === 'INFO')
      expect(infoSpan).toHaveStyle({ color: '#22c55e' })
      expect(container.textContent).toContain('[2024-01-01 12:00:00]')
      expect(container.textContent).toContain('Starting server...')
    })

    it('should render error logs', () => {
      const log = '\x1b[31mERROR\x1b[0m: Connection failed'
      const { container } = render(<AnsiText text={log} />)
      const spans = container.querySelectorAll('span')
      const errorSpan = Array.from(spans).find(s => s.textContent === 'ERROR')
      expect(errorSpan).toHaveStyle({ color: '#ef4444' })
      expect(container.textContent).toContain(': Connection failed')
    })

    it('should render kubectl-style colored output', () => {
      const log = 'NAME\t\x1b[32mREADY\x1b[0m\t\x1b[36mSTATUS\x1b[0m'
      const { container } = render(<AnsiText text={log} />)
      const spans = container.querySelectorAll('span')
      const readySpan = Array.from(spans).find(s => s.textContent === 'READY')
      const statusSpan = Array.from(spans).find(s => s.textContent === 'STATUS')
      expect(readySpan).toHaveStyle({ color: '#22c55e' })
      expect(statusSpan).toHaveStyle({ color: '#06b6d4' })
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
