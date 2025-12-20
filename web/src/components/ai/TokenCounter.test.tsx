import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { TokenCounter } from './TokenCounter'
import type { AIModel } from '../../types/ai'

const mockModel: AIModel = {
  id: 'gpt-4o',
  name: 'GPT-4o',
  provider: 'openai',
  maxTokens: 4096,
  contextWindow: 128000,
  inputCostPer1M: 500,
  outputCostPer1M: 1500,
  capabilities: ['chat', 'function_calling', 'vision'],
}

describe('TokenCounter', () => {
  it('should display token count', () => {
    render(<TokenCounter tokens={100} />)
    expect(screen.getByText('100 tokens')).toBeInTheDocument()
  })

  it('should display singular for 1 token', () => {
    render(<TokenCounter tokens={1} />)
    expect(screen.getByText('1 token')).toBeInTheDocument()
  })

  it('should format large token counts with K suffix', () => {
    render(<TokenCounter tokens={1500} />)
    expect(screen.getByText('1.5K tokens')).toBeInTheDocument()
  })

  it('should format very large token counts with M suffix', () => {
    render(<TokenCounter tokens={1500000} />)
    expect(screen.getByText('1.5M tokens')).toBeInTheDocument()
  })

  it('should display estimated cost when model is provided', () => {
    render(<TokenCounter tokens={1000} model={mockModel} />)
    // 1000 tokens at $5/1M = $0.005 = less than $0.01
    expect(screen.getByText('<$0.01')).toBeInTheDocument()
  })

  it('should display cost in dollars for larger token counts', () => {
    render(<TokenCounter tokens={1000000} model={mockModel} />)
    // 1M tokens at $5/1M input = $5.00
    expect(screen.getByText('$5.00')).toBeInTheDocument()
  })

  it('should include output tokens in cost calculation', () => {
    render(<TokenCounter tokens={1000000} outputTokens={500000} model={mockModel} />)
    // 1M input at $5/1M = $5.00 + 500K output at $15/1M = $7.50 = $12.50
    expect(screen.getByText('$12.50')).toBeInTheDocument()
  })

  it('should show loading state', () => {
    render(<TokenCounter tokens={0} loading />)
    expect(screen.getByText('Counting...')).toBeInTheDocument()
  })

  it('should show context window usage when model is provided', () => {
    render(<TokenCounter tokens={64000} model={mockModel} showContextUsage />)
    // 64000 / 128000 = 50%
    expect(screen.getByText('50%')).toBeInTheDocument()
  })

  it('should show warning when approaching context limit', () => {
    render(<TokenCounter tokens={115200} model={mockModel} showContextUsage />)
    // 115200 / 128000 = 90%
    const percentElement = screen.getByText('90%')
    expect(percentElement).toHaveClass('text-yellow-600')
  })

  it('should show error when exceeding context limit', () => {
    render(<TokenCounter tokens={140000} model={mockModel} showContextUsage />)
    // 140000 / 128000 = 109%
    const percentElement = screen.getByText('109%')
    expect(percentElement).toHaveClass('text-red-600')
  })

  it('should display label when provided', () => {
    render(<TokenCounter tokens={100} label="Input Tokens" />)
    expect(screen.getByText('Input Tokens:')).toBeInTheDocument()
  })

  it('should use compact display when size is small', () => {
    render(<TokenCounter tokens={1500} size="small" />)
    expect(screen.getByText('1.5K')).toBeInTheDocument()
  })

  it('should show breakdown when showBreakdown is true', () => {
    render(
      <TokenCounter tokens={500} outputTokens={200} model={mockModel} showBreakdown />
    )
    expect(screen.getByText(/500.*input/i)).toBeInTheDocument()
    expect(screen.getByText(/200.*output/i)).toBeInTheDocument()
  })

  it('should handle zero tokens', () => {
    render(<TokenCounter tokens={0} />)
    expect(screen.getByText('0 tokens')).toBeInTheDocument()
  })

  it('should handle undefined outputTokens', () => {
    render(<TokenCounter tokens={100} model={mockModel} />)
    expect(screen.getByText('<$0.01')).toBeInTheDocument()
  })
})
