import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ReactFlowProvider } from '@xyflow/react'
import AINode from './AINode'

// Wrapper component for ReactFlow context
const TestWrapper = ({ children }: { children: React.ReactNode }) => (
  <ReactFlowProvider>{children}</ReactFlowProvider>
)

describe('AINode', () => {
  const defaultProps = {
    id: 'ai-node-1',
    data: {
      label: 'Test AI Node',
      nodeType: 'ai_chat',
    },
    selected: false,
  }

  it('should render the node with label', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} />
      </TestWrapper>
    )

    expect(screen.getByText('Test AI Node')).toBeInTheDocument()
  })

  it('should display chat completion action type', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_chat' }} />
      </TestWrapper>
    )

    expect(screen.getByText('Chat Completion')).toBeInTheDocument()
  })

  it('should display summarization action type', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_summarize' }} />
      </TestWrapper>
    )

    expect(screen.getByText('Summarize')).toBeInTheDocument()
  })

  it('should display classification action type', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_classify' }} />
      </TestWrapper>
    )

    expect(screen.getByText('Classify')).toBeInTheDocument()
  })

  it('should display entity extraction action type', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_extract' }} />
      </TestWrapper>
    )

    expect(screen.getByText('Extract Entities')).toBeInTheDocument()
  })

  it('should display embedding action type', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_embed' }} />
      </TestWrapper>
    )

    expect(screen.getByText('Embeddings')).toBeInTheDocument()
  })

  it('should show robot icon for chat completion', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_chat' }} />
      </TestWrapper>
    )

    // Should have the robot emoji
    expect(screen.getByText(/🤖/)).toBeInTheDocument()
  })

  it('should show document icon for summarization', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_summarize' }} />
      </TestWrapper>
    )

    expect(screen.getByText(/📝/)).toBeInTheDocument()
  })

  it('should show tag icon for classification', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_classify' }} />
      </TestWrapper>
    )

    expect(screen.getByText(/🏷️/)).toBeInTheDocument()
  })

  it('should show search icon for entity extraction', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_extract' }} />
      </TestWrapper>
    )

    expect(screen.getByText(/🔍/)).toBeInTheDocument()
  })

  it('should show chart icon for embeddings', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_embed' }} />
      </TestWrapper>
    )

    expect(screen.getByText(/📊/)).toBeInTheDocument()
  })

  it('should apply selection styles when selected', () => {
    const { container } = render(
      <TestWrapper>
        <AINode {...defaultProps} selected={true} />
      </TestWrapper>
    )

    // Should have ring class when selected
    const nodeElement = container.querySelector('.ring-2')
    expect(nodeElement).toBeInTheDocument()
  })

  it('should not apply selection styles when not selected', () => {
    const { container } = render(
      <TestWrapper>
        <AINode {...defaultProps} selected={false} />
      </TestWrapper>
    )

    // Should not have ring class when not selected
    const nodeElement = container.querySelector('.ring-2')
    expect(nodeElement).not.toBeInTheDocument()
  })

  it('should display model name when provided', () => {
    render(
      <TestWrapper>
        <AINode
          {...defaultProps}
          data={{
            ...defaultProps.data,
            aiConfig: { model: 'gpt-4o' },
          }}
        />
      </TestWrapper>
    )

    expect(screen.getByText('gpt-4o')).toBeInTheDocument()
  })

  it('should display fallback text when no model configured', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} />
      </TestWrapper>
    )

    expect(screen.getByText('No model selected')).toBeInTheDocument()
  })

  it('should have gradient background with AI-specific colors', () => {
    const { container } = render(
      <TestWrapper>
        <AINode {...defaultProps} />
      </TestWrapper>
    )

    // Check for AI-specific gradient (violet/purple theme)
    const nodeElement = container.querySelector('.from-violet-500')
    expect(nodeElement).toBeInTheDocument()
  })

  it('should render input handle at top', () => {
    const { container } = render(
      <TestWrapper>
        <AINode {...defaultProps} />
      </TestWrapper>
    )

    // ReactFlow handles have specific data attributes
    const inputHandle = container.querySelector('[data-handlepos="top"]')
    expect(inputHandle).toBeInTheDocument()
  })

  it('should render output handle at bottom', () => {
    const { container } = render(
      <TestWrapper>
        <AINode {...defaultProps} />
      </TestWrapper>
    )

    // ReactFlow handles have specific data attributes
    const outputHandle = container.querySelector('[data-handlepos="bottom"]')
    expect(outputHandle).toBeInTheDocument()
  })

  it('should handle unknown node type gracefully', () => {
    render(
      <TestWrapper>
        <AINode {...defaultProps} data={{ ...defaultProps.data, nodeType: 'ai_unknown' }} />
      </TestWrapper>
    )

    // Should show fallback
    expect(screen.getByText('AI Action')).toBeInTheDocument()
    expect(screen.getByText(/🧠/)).toBeInTheDocument()
  })
})
