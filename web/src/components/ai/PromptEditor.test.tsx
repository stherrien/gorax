import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { PromptEditor } from './PromptEditor'

describe('PromptEditor', () => {
  const defaultProps = {
    value: '',
    onChange: vi.fn(),
  }

  it('should render with empty value', () => {
    render(<PromptEditor {...defaultProps} />)
    expect(screen.getByRole('textbox')).toHaveValue('')
  })

  it('should render with initial value', () => {
    render(<PromptEditor {...defaultProps} value="Hello world" />)
    expect(screen.getByRole('textbox')).toHaveValue('Hello world')
  })

  it('should call onChange when text changes', () => {
    const onChange = vi.fn()
    render(<PromptEditor {...defaultProps} onChange={onChange} />)

    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: 'New text' },
    })

    expect(onChange).toHaveBeenCalledWith('New text')
  })

  it('should display placeholder text', () => {
    render(<PromptEditor {...defaultProps} placeholder="Enter your prompt..." />)
    expect(screen.getByPlaceholderText('Enter your prompt...')).toBeInTheDocument()
  })

  it('should show label when provided', () => {
    render(<PromptEditor {...defaultProps} label="System Prompt" />)
    expect(screen.getByText('System Prompt')).toBeInTheDocument()
  })

  it('should show variable picker when showVariables is true', () => {
    const variables = ['trigger.data', 'steps.previous.output', 'env.api_url']
    render(<PromptEditor {...defaultProps} showVariables variables={variables} />)

    fireEvent.click(screen.getByText('Insert Variable'))
    expect(screen.getByText('trigger.data')).toBeInTheDocument()
    expect(screen.getByText('steps.previous.output')).toBeInTheDocument()
  })

  it('should insert variable at cursor position', () => {
    const onChange = vi.fn()
    const variables = ['trigger.data']
    render(
      <PromptEditor
        {...defaultProps}
        value="Hello "
        onChange={onChange}
        showVariables
        variables={variables}
      />
    )

    const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
    // Simulate clicking at the end of the text
    textarea.setSelectionRange(6, 6)
    fireEvent.click(textarea)

    // Click insert variable
    fireEvent.click(screen.getByText('Insert Variable'))
    fireEvent.click(screen.getByText('trigger.data'))

    // Should insert the variable template at cursor
    expect(onChange).toHaveBeenCalledWith('Hello {{trigger.data}}')
  })

  it('should be disabled when disabled prop is true', () => {
    render(<PromptEditor {...defaultProps} disabled />)
    expect(screen.getByRole('textbox')).toBeDisabled()
  })

  it('should apply custom rows', () => {
    render(<PromptEditor {...defaultProps} rows={10} />)
    expect(screen.getByRole('textbox')).toHaveAttribute('rows', '10')
  })

  it('should show error message when error prop is provided', () => {
    render(<PromptEditor {...defaultProps} error="Prompt is required" />)
    expect(screen.getByText('Prompt is required')).toBeInTheDocument()
  })

  it('should show helper text when provided', () => {
    render(<PromptEditor {...defaultProps} helperText="Use {{variable}} syntax" />)
    expect(screen.getByText('Use {{variable}} syntax')).toBeInTheDocument()
  })

  it('should close variable picker when variable is selected', () => {
    const variables = ['trigger.data']
    render(<PromptEditor {...defaultProps} showVariables variables={variables} />)

    fireEvent.click(screen.getByText('Insert Variable'))
    expect(screen.getByText('trigger.data')).toBeInTheDocument()

    fireEvent.click(screen.getByText('trigger.data'))
    expect(screen.queryByText('trigger.data')).not.toBeInTheDocument()
  })

  it('should show character count when maxLength is provided', () => {
    render(<PromptEditor {...defaultProps} value="Hello" maxLength={100} showCharCount />)
    expect(screen.getByText('5 / 100')).toBeInTheDocument()
  })

  it('should highlight warning when approaching maxLength', () => {
    render(<PromptEditor {...defaultProps} value={'x'.repeat(95)} maxLength={100} showCharCount />)
    expect(screen.getByText('95 / 100')).toHaveClass('text-yellow-600')
  })

  it('should highlight error when exceeding maxLength', () => {
    render(<PromptEditor {...defaultProps} value={'x'.repeat(105)} maxLength={100} showCharCount />)
    expect(screen.getByText('105 / 100')).toHaveClass('text-red-600')
  })

  it('should allow custom className', () => {
    render(<PromptEditor {...defaultProps} className="custom-class" />)
    // The className is applied to the outer wrapper div (grandparent of textarea)
    expect(screen.getByRole('textbox').parentElement?.parentElement).toHaveClass('custom-class')
  })

  it('should support read-only mode', () => {
    render(<PromptEditor {...defaultProps} value="Read only text" readOnly />)
    expect(screen.getByRole('textbox')).toHaveAttribute('readonly')
  })
})
