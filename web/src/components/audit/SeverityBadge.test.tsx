import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SeverityBadge } from './SeverityBadge'
import { AuditSeverity } from '../../types/audit'

describe('SeverityBadge', () => {
  it('should render info severity', () => {
    render(<SeverityBadge severity={AuditSeverity.Info} />)
    expect(screen.getByText('Info')).toBeInTheDocument()
  })

  it('should render warning severity', () => {
    render(<SeverityBadge severity={AuditSeverity.Warning} />)
    expect(screen.getByText('Warning')).toBeInTheDocument()
  })

  it('should render error severity', () => {
    render(<SeverityBadge severity={AuditSeverity.Error} />)
    expect(screen.getByText('Error')).toBeInTheDocument()
  })

  it('should render critical severity', () => {
    render(<SeverityBadge severity={AuditSeverity.Critical} />)
    expect(screen.getByText('Critical')).toBeInTheDocument()
  })

  it('should apply correct size classes', () => {
    const { rerender } = render(<SeverityBadge severity={AuditSeverity.Info} size="sm" />)
    let badge = screen.getByText('Info')
    expect(badge).toHaveClass('px-2', 'py-0.5', 'text-xs')

    rerender(<SeverityBadge severity={AuditSeverity.Info} size="md" />)
    badge = screen.getByText('Info')
    expect(badge).toHaveClass('px-2.5', 'py-1', 'text-sm')

    rerender(<SeverityBadge severity={AuditSeverity.Info} size="lg" />)
    badge = screen.getByText('Info')
    expect(badge).toHaveClass('px-3', 'py-1.5', 'text-base')
  })

  it('should have accessible label', () => {
    render(<SeverityBadge severity={AuditSeverity.Critical} />)
    const badge = screen.getByRole('status')
    expect(badge).toHaveAttribute('aria-label', 'Severity: Critical')
  })
})
