import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { AIConfigPanel } from './AIConfigPanel'
import type { AIAction, AIModel } from '../../types/ai'
import type { Credential } from '../../api/credentials'

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

const mockCredentials: Credential[] = [
  {
    id: 'cred-1',
    tenantId: 'tenant-1',
    name: 'OpenAI API Key',
    type: 'api_key',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
]

describe('AIConfigPanel', () => {
  const defaultProps = {
    action: 'chat_completion' as AIAction,
    config: {},
    models: mockModels,
    credentials: mockCredentials,
    onChange: vi.fn(),
  }

  it('should render action selector', () => {
    render(<AIConfigPanel {...defaultProps} />)
    expect(screen.getByLabelText('Action Type')).toBeInTheDocument()
  })

  it('should show credential selector', () => {
    render(<AIConfigPanel {...defaultProps} />)
    expect(screen.getByText('Credential')).toBeInTheDocument()
  })

  it('should show model selector', () => {
    render(<AIConfigPanel {...defaultProps} />)
    expect(screen.getByText('Model')).toBeInTheDocument()
  })

  it('should render chat completion fields', () => {
    render(<AIConfigPanel {...defaultProps} action="chat_completion" />)
    expect(screen.getByText('System Prompt')).toBeInTheDocument()
  })

  it('should render summarization fields when action is summarization', () => {
    render(<AIConfigPanel {...defaultProps} action="summarization" />)
    expect(screen.getByText('Text to Summarize')).toBeInTheDocument()
    expect(screen.getByText('Summary Format')).toBeInTheDocument()
  })

  it('should render classification fields when action is classification', () => {
    render(<AIConfigPanel {...defaultProps} action="classification" />)
    expect(screen.getByText('Text to Classify')).toBeInTheDocument()
    expect(screen.getByText('Categories')).toBeInTheDocument()
  })

  it('should render entity extraction fields when action is entity_extraction', () => {
    render(<AIConfigPanel {...defaultProps} action="entity_extraction" />)
    expect(screen.getByText('Text to Analyze')).toBeInTheDocument()
    expect(screen.getByText('Entity Types')).toBeInTheDocument()
  })

  it('should render embedding fields when action is embedding', () => {
    render(<AIConfigPanel {...defaultProps} action="embedding" />)
    expect(screen.getByText('Texts to Embed')).toBeInTheDocument()
  })

  it('should call onChange when action type changes', () => {
    const onChange = vi.fn()
    render(<AIConfigPanel {...defaultProps} onChange={onChange} />)

    fireEvent.change(screen.getByLabelText('Action Type'), {
      target: { value: 'summarization' },
    })

    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ action: 'summarization' }))
  })

  it('should filter models based on action capability', () => {
    render(<AIConfigPanel {...defaultProps} action="embedding" />)
    fireEvent.click(screen.getByText('Select Model'))

    // Should show embedding model
    expect(screen.getByText('Text Embedding 3 Small')).toBeInTheDocument()
    // Should not show chat model
    expect(screen.queryByText('GPT-4o')).not.toBeInTheDocument()
  })

  it('should show chat models for chat_completion action', () => {
    render(<AIConfigPanel {...defaultProps} action="chat_completion" />)
    fireEvent.click(screen.getByText('Select Model'))

    // Should show chat model
    expect(screen.getByText('GPT-4o')).toBeInTheDocument()
    // Should not show embedding model
    expect(screen.queryByText('Text Embedding 3 Small')).not.toBeInTheDocument()
  })

  it('should show advanced options when expanded', () => {
    render(<AIConfigPanel {...defaultProps} />)

    fireEvent.click(screen.getByText('Advanced Options'))

    expect(screen.getByText('Temperature')).toBeInTheDocument()
    expect(screen.getByText('Max Tokens')).toBeInTheDocument()
  })

  it('should call onChange when model is selected', () => {
    const onChange = vi.fn()
    render(<AIConfigPanel {...defaultProps} onChange={onChange} />)

    fireEvent.click(screen.getByText('Select Model'))
    fireEvent.click(screen.getByText('GPT-4o'))

    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ model: 'gpt-4o' }))
  })

  it('should show error message when provided', () => {
    render(<AIConfigPanel {...defaultProps} error="Configuration error" />)
    expect(screen.getByText('Configuration error')).toBeInTheDocument()
  })

  it('should disable fields when disabled prop is true', () => {
    render(<AIConfigPanel {...defaultProps} disabled />)
    expect(screen.getByLabelText('Action Type')).toBeDisabled()
  })

  it('should populate form with existing config', () => {
    const config = {
      model: 'gpt-4o',
      systemPrompt: 'You are a helpful assistant',
      temperature: 0.7,
    }
    render(<AIConfigPanel {...defaultProps} config={config} />)

    expect(screen.getByDisplayValue('You are a helpful assistant')).toBeInTheDocument()
  })
})
