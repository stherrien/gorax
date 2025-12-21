import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ModelSelector } from './ModelSelector'
import type { AIModel, AIProvider } from '../../types/ai'

const mockModels: AIModel[] = [
  {
    id: 'gpt-4o',
    name: 'GPT-4o',
    provider: 'openai',
    maxTokens: 4096,
    contextWindow: 128000,
    inputCostPer1M: 500,
    outputCostPer1M: 1500,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  {
    id: 'gpt-4o-mini',
    name: 'GPT-4o Mini',
    provider: 'openai',
    maxTokens: 16384,
    contextWindow: 128000,
    inputCostPer1M: 15,
    outputCostPer1M: 60,
    capabilities: ['chat', 'function_calling'],
  },
  {
    id: 'claude-3-5-sonnet-20241022',
    name: 'Claude 3.5 Sonnet',
    provider: 'anthropic',
    maxTokens: 8192,
    contextWindow: 200000,
    inputCostPer1M: 300,
    outputCostPer1M: 1500,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  {
    id: 'text-embedding-3-small',
    name: 'Text Embedding 3 Small',
    provider: 'openai',
    maxTokens: 0,
    contextWindow: 8191,
    inputCostPer1M: 2,
    outputCostPer1M: 0,
    capabilities: ['embedding'],
  },
]

describe('ModelSelector', () => {
  const defaultProps = {
    models: mockModels,
    value: '',
    onChange: vi.fn(),
  }

  it('should render with placeholder when no model selected', () => {
    render(<ModelSelector {...defaultProps} />)
    expect(screen.getByText('Select Model')).toBeInTheDocument()
  })

  it('should display selected model name', () => {
    render(<ModelSelector {...defaultProps} value="gpt-4o" />)
    expect(screen.getByText('GPT-4o')).toBeInTheDocument()
  })

  it('should open dropdown when clicked', () => {
    render(<ModelSelector {...defaultProps} />)
    fireEvent.click(screen.getByRole('button'))
    expect(screen.getByText('GPT-4o Mini')).toBeInTheDocument()
  })

  it('should call onChange when model is selected', () => {
    const onChange = vi.fn()
    render(<ModelSelector {...defaultProps} onChange={onChange} />)

    fireEvent.click(screen.getByRole('button'))
    fireEvent.click(screen.getByText('GPT-4o Mini'))

    expect(onChange).toHaveBeenCalledWith('gpt-4o-mini')
  })

  it('should filter models by provider when specified', () => {
    render(<ModelSelector {...defaultProps} filterProvider="anthropic" />)
    fireEvent.click(screen.getByRole('button'))

    expect(screen.getByText('Claude 3.5 Sonnet')).toBeInTheDocument()
    expect(screen.queryByText('GPT-4o')).not.toBeInTheDocument()
  })

  it('should filter models by capability when specified', () => {
    render(<ModelSelector {...defaultProps} filterCapability="embedding" />)
    fireEvent.click(screen.getByRole('button'))

    expect(screen.getByText('Text Embedding 3 Small')).toBeInTheDocument()
    expect(screen.queryByText('GPT-4o')).not.toBeInTheDocument()
  })

  it('should filter chat models (excluding embedding)', () => {
    render(<ModelSelector {...defaultProps} filterCapability="chat" />)
    fireEvent.click(screen.getByRole('button'))

    expect(screen.getByText('GPT-4o')).toBeInTheDocument()
    expect(screen.getByText('Claude 3.5 Sonnet')).toBeInTheDocument()
    expect(screen.queryByText('Text Embedding 3 Small')).not.toBeInTheDocument()
  })

  it('should show pricing info when showPricing is true', () => {
    render(<ModelSelector {...defaultProps} showPricing />)
    fireEvent.click(screen.getByRole('button'))

    // Look for pricing info in the dropdown (500 cents per 1M = $5.00)
    expect(screen.getByText(/\$5\.00 input/i)).toBeInTheDocument()
  })

  it('should show context window info when showContext is true', () => {
    render(<ModelSelector {...defaultProps} showContext />)
    fireEvent.click(screen.getByRole('button'))

    // Look for context window info (multiple models have 128K, so use getAllByText)
    const contextElements = screen.getAllByText(/128K context/i)
    expect(contextElements.length).toBeGreaterThan(0)
  })

  it('should be disabled when disabled prop is true', () => {
    render(<ModelSelector {...defaultProps} disabled />)
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('should show loading state', () => {
    render(<ModelSelector {...defaultProps} loading />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('should show custom placeholder', () => {
    render(<ModelSelector {...defaultProps} placeholder="Choose an AI model" />)
    expect(screen.getByText('Choose an AI model')).toBeInTheDocument()
  })

  it('should close dropdown when model is selected', () => {
    render(<ModelSelector {...defaultProps} />)

    fireEvent.click(screen.getByRole('button'))
    expect(screen.getByText('GPT-4o Mini')).toBeInTheDocument()

    fireEvent.click(screen.getByText('GPT-4o'))

    // Dropdown should close
    expect(screen.queryByText('GPT-4o Mini')).not.toBeInTheDocument()
  })

  it('should group models by provider', () => {
    render(<ModelSelector {...defaultProps} groupByProvider />)
    fireEvent.click(screen.getByRole('button'))

    expect(screen.getByText('OpenAI')).toBeInTheDocument()
    expect(screen.getByText('Anthropic')).toBeInTheDocument()
  })

  it('should display provider badge for selected model', () => {
    render(<ModelSelector {...defaultProps} value="gpt-4o" showProviderBadge />)
    expect(screen.getByText('openai')).toBeInTheDocument()
  })

  it('should search models when searchable is enabled', () => {
    render(<ModelSelector {...defaultProps} searchable />)
    fireEvent.click(screen.getByRole('button'))

    const searchInput = screen.getByPlaceholderText('Search models...')
    fireEvent.change(searchInput, { target: { value: 'claude' } })

    expect(screen.getByText('Claude 3.5 Sonnet')).toBeInTheDocument()
    expect(screen.queryByText('GPT-4o')).not.toBeInTheDocument()
  })
})
