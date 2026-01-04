import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { AuditLogTable } from './AuditLogTable'
import {
  AuditCategory,
  AuditEventType,
  AuditSeverity,
  AuditStatus,
  AuditEvent,
} from '../../types/audit'

const mockEvents: AuditEvent[] = [
  {
    id: '1',
    tenantId: 'tenant1',
    userId: 'user1',
    userEmail: 'user@example.com',
    category: AuditCategory.Authentication,
    eventType: AuditEventType.Login,
    action: 'user.login',
    resourceType: 'user',
    resourceId: 'user1',
    resourceName: 'User 1',
    ipAddress: '192.168.1.1',
    userAgent: 'Mozilla/5.0',
    severity: AuditSeverity.Info,
    status: AuditStatus.Success,
    metadata: {},
    createdAt: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    tenantId: 'tenant1',
    userId: 'user2',
    userEmail: 'user2@example.com',
    category: AuditCategory.Workflow,
    eventType: AuditEventType.Execute,
    action: 'workflow.execute',
    resourceType: 'workflow',
    resourceId: 'wf1',
    resourceName: 'Test Workflow',
    ipAddress: '192.168.1.2',
    userAgent: 'Mozilla/5.0',
    severity: AuditSeverity.Error,
    status: AuditStatus.Failure,
    metadata: {},
    createdAt: '2024-01-01T01:00:00Z',
  },
]

describe('AuditLogTable', () => {
  const defaultProps = {
    events: mockEvents,
    total: 2,
    currentPage: 1,
    pageSize: 50,
    onPageChange: vi.fn(),
  }

  it('should render audit events', () => {
    render(<AuditLogTable {...defaultProps} />)

    expect(screen.getByText('user@example.com')).toBeInTheDocument()
    expect(screen.getByText('user2@example.com')).toBeInTheDocument()
    expect(screen.getByText('user.login')).toBeInTheDocument()
    expect(screen.getByText('workflow.execute')).toBeInTheDocument()
  })

  it('should show loading state', () => {
    render(<AuditLogTable {...defaultProps} isLoading={true} />)
    expect(screen.getByText('Loading audit logs...')).toBeInTheDocument()
  })

  it('should show empty state when no events', () => {
    render(<AuditLogTable {...defaultProps} events={[]} total={0} />)
    expect(screen.getByText('No audit events found')).toBeInTheDocument()
  })

  it('should handle page navigation', () => {
    const onPageChange = vi.fn()
    render(
      <AuditLogTable
        {...defaultProps}
        onPageChange={onPageChange}
        currentPage={1}
        total={100}
      />
    )

    const nextButton = screen.getAllByRole('button').find((btn) =>
      btn.querySelector('svg')
    )
    if (nextButton) {
      fireEvent.click(nextButton)
    }

    expect(onPageChange).toHaveBeenCalled()
  })

  it('should handle row click to open modal', () => {
    render(<AuditLogTable {...defaultProps} />)

    const firstRow = screen.getByText('user@example.com').closest('tr')
    if (firstRow) {
      fireEvent.click(firstRow)
    }

    // Modal should be opened (check for modal title)
    expect(screen.getByText('Audit Event Details')).toBeInTheDocument()
  })

  it('should show pagination info', () => {
    render(<AuditLogTable {...defaultProps} total={100} pageSize={50} currentPage={1} />)

    expect(screen.getByText(/Showing/)).toBeInTheDocument()
    expect(screen.getByText(/1/)).toBeInTheDocument()
    expect(screen.getByText(/50/)).toBeInTheDocument()
    expect(screen.getByText(/100/)).toBeInTheDocument()
  })

  it('should disable previous button on first page', () => {
    render(<AuditLogTable {...defaultProps} currentPage={1} />)

    const prevButton = screen.getAllByRole('button')[0]
    expect(prevButton).toBeDisabled()
  })

  it('should disable next button on last page', () => {
    render(<AuditLogTable {...defaultProps} currentPage={1} total={2} pageSize={50} />)

    const buttons = screen.getAllByRole('button')
    const nextButton = buttons[buttons.length - 1]
    expect(nextButton).toBeDisabled()
  })
})
