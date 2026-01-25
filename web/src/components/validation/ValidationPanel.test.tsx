/**
 * Tests for ValidationPanel component
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ValidationPanel, { ValidationBadge } from './ValidationPanel'
import type { ValidationResult, ValidationIssue } from '../../types/workflow'

// ============================================================================
// Test Fixtures
// ============================================================================

let issueIdCounter = 0

const createValidResult = (): ValidationResult => ({
  valid: true,
  issues: [],
  executionOrder: ['trigger-1', 'action-1'],
})

const createInvalidResult = (issues: ValidationIssue[]): ValidationResult => ({
  valid: false,
  issues,
  executionOrder: undefined,
})

const createErrorIssue = (message: string, nodeId?: string): ValidationIssue => ({
  id: `issue-error-${++issueIdCounter}`,
  severity: 'error',
  message,
  nodeId,
  suggestion: 'Fix this issue',
})

const createWarningIssue = (message: string, nodeId?: string): ValidationIssue => ({
  id: `issue-warning-${++issueIdCounter}`,
  severity: 'warning',
  message,
  nodeId,
  suggestion: 'Consider fixing this',
})

// ============================================================================
// ValidationPanel Tests
// ============================================================================

describe('ValidationPanel', () => {
  it('should render "No validation results" when result is null', () => {
    render(<ValidationPanel result={null} />)

    expect(screen.getByText(/run validation/i)).toBeInTheDocument()
  })

  it('should render valid state correctly', () => {
    const result = createValidResult()

    render(<ValidationPanel result={result} />)

    // Should show the valid message and ready to run text
    expect(screen.getByText(/ready to run/i)).toBeInTheDocument()
  })

  it('should render error count in summary', () => {
    const result = createInvalidResult([
      createErrorIssue('Error 1'),
      createErrorIssue('Error 2'),
    ])

    render(<ValidationPanel result={result} />)

    expect(screen.getByText(/2 errors/i)).toBeInTheDocument()
  })

  it('should render warning count in summary', () => {
    const result = createInvalidResult([
      createWarningIssue('Warning 1'),
      createWarningIssue('Warning 2'),
      createWarningIssue('Warning 3'),
    ])
    // Make it valid by setting valid: true
    result.valid = true

    render(<ValidationPanel result={result} />)

    expect(screen.getByText(/3 warnings/i)).toBeInTheDocument()
  })

  it('should call onNavigateToNode when clicking issue with nodeId', () => {
    const onNavigateToNode = vi.fn()
    const result = createInvalidResult([
      createErrorIssue('Node error', 'node-123'),
    ])

    render(<ValidationPanel result={result} onNavigateToNode={onNavigateToNode} />)

    // Find and click the issue
    const issueElement = screen.getByText('Node error')
    fireEvent.click(issueElement)

    expect(onNavigateToNode).toHaveBeenCalledWith('node-123')
  })

  it('should toggle collapse when clicking header', () => {
    const onToggleCollapse = vi.fn()
    const result = createInvalidResult([createErrorIssue('Error')])

    render(
      <ValidationPanel
        result={result}
        collapsed={false}
        onToggleCollapse={onToggleCollapse}
      />
    )

    // Click the header button
    const header = screen.getByRole('button')
    fireEvent.click(header)

    expect(onToggleCollapse).toHaveBeenCalled()
  })

  it('should not render issues when collapsed', () => {
    const result = createInvalidResult([createErrorIssue('Hidden error')])

    render(<ValidationPanel result={result} collapsed={true} />)

    expect(screen.queryByText('Hidden error')).not.toBeInTheDocument()
  })

  it('should render execution order for valid workflow', () => {
    const result = createValidResult()

    render(<ValidationPanel result={result} collapsed={false} />)

    expect(screen.getByText('trigger-1')).toBeInTheDocument()
    expect(screen.getByText('action-1')).toBeInTheDocument()
  })

  it('should show auto-fix button for auto-fixable issues', () => {
    const onAutoFix = vi.fn()
    const issue: ValidationIssue = {
      id: 'issue-1',
      severity: 'error',
      message: 'Fixable error',
      autoFixable: true,
    }
    const result = createInvalidResult([issue])

    render(<ValidationPanel result={result} onAutoFix={onAutoFix} />)

    const fixButton = screen.getByRole('button', { name: /fix/i })
    expect(fixButton).toBeInTheDocument()

    fireEvent.click(fixButton)
    expect(onAutoFix).toHaveBeenCalledWith(issue)
  })

  it('should display suggestion text', () => {
    const issue: ValidationIssue = {
      id: 'issue-1',
      severity: 'error',
      message: 'Error message',
      suggestion: 'Here is how to fix it',
    }
    const result = createInvalidResult([issue])

    render(<ValidationPanel result={result} />)

    expect(screen.getByText('Here is how to fix it')).toBeInTheDocument()
  })

  it('should display node ID badge', () => {
    const result = createInvalidResult([
      createErrorIssue('Node-specific error', 'my-node-id'),
    ])

    render(<ValidationPanel result={result} />)

    expect(screen.getByText('my-node-id')).toBeInTheDocument()
  })

  it('should display field name badge', () => {
    const issue: ValidationIssue = {
      id: 'issue-1',
      severity: 'error',
      message: 'Field error',
      nodeId: 'node-1',
      field: 'url',
    }
    const result = createInvalidResult([issue])

    render(<ValidationPanel result={result} />)

    expect(screen.getByText('url')).toBeInTheDocument()
  })
})

// ============================================================================
// ValidationBadge Tests
// ============================================================================

describe('ValidationBadge', () => {
  it('should render nothing when result is null', () => {
    const { container } = render(<ValidationBadge result={null} />)

    expect(container.firstChild).toBeNull()
  })

  it('should render "Valid" for valid workflow', () => {
    const result = createValidResult()

    render(<ValidationBadge result={result} />)

    expect(screen.getByText('Valid')).toBeInTheDocument()
  })

  it('should render error count for invalid workflow', () => {
    const result = createInvalidResult([
      createErrorIssue('Error 1'),
      createErrorIssue('Error 2'),
    ])

    render(<ValidationBadge result={result} />)

    expect(screen.getByText(/2 errors/i)).toBeInTheDocument()
  })

  it('should call onClick when clicked', () => {
    const onClick = vi.fn()
    const result = createValidResult()

    render(<ValidationBadge result={result} onClick={onClick} />)

    const badge = screen.getByRole('button')
    fireEvent.click(badge)

    expect(onClick).toHaveBeenCalled()
  })
})
