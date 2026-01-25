import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import { auditAPI } from '../api/audit'
import {
  useAuditEvents,
  useAuditEvent,
  useAuditStats,
  useExportAudit,
  useAuditCategories,
  useAuditEventTypes,
} from './useAudit'
import {
  AuditCategory,
  AuditEventType,
  AuditSeverity,
  AuditStatus,
  ExportFormat,
} from '../types/audit'

vi.mock('../api/audit')

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useAuditEvents', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch audit events with filter', async () => {
    const mockResponse = {
      events: [
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
      ],
      total: 1,
      page: 1,
      limit: 50,
    }

    vi.mocked(auditAPI.queryAuditEvents).mockResolvedValue(mockResponse)

    const { result } = renderHook(
      () => useAuditEvents({ limit: 50, offset: 0 }),
      { wrapper: createWrapper() }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.events).toHaveLength(1)
    expect(result.current.data?.total).toBe(1)
  })

  it('should not fetch when disabled', async () => {
    const { result } = renderHook(
      () => useAuditEvents({ limit: 50 }, false),
      { wrapper: createWrapper() }
    )

    expect(result.current.isLoading).toBe(false)
    expect(auditAPI.queryAuditEvents).not.toHaveBeenCalled()
  })

  it('should handle errors', async () => {
    vi.mocked(auditAPI.queryAuditEvents).mockRejectedValue(
      new Error('Failed to fetch')
    )

    const { result } = renderHook(() => useAuditEvents({ limit: 50 }), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))

    expect(result.current.error?.message).toBe('Failed to fetch')
  })
})

describe('useAuditEvent', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch a single audit event', async () => {
    const mockEvent = {
      id: 'event1',
      tenantId: 'tenant1',
      userId: 'user1',
      userEmail: 'user@example.com',
      category: AuditCategory.Workflow,
      eventType: AuditEventType.Execute,
      action: 'workflow.execute',
      resourceType: 'workflow',
      resourceId: 'wf1',
      resourceName: 'Test Workflow',
      ipAddress: '192.168.1.1',
      userAgent: 'Mozilla/5.0',
      severity: AuditSeverity.Info,
      status: AuditStatus.Success,
      metadata: {},
      createdAt: '2024-01-01T00:00:00Z',
    }

    vi.mocked(auditAPI.getAuditEvent).mockResolvedValue(mockEvent)

    const { result } = renderHook(() => useAuditEvent('event1'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toEqual(mockEvent)
  })

  it('should not fetch when eventId is empty', async () => {
    const { result } = renderHook(() => useAuditEvent(''), {
      wrapper: createWrapper(),
    })

    expect(result.current.isLoading).toBe(false)
    expect(auditAPI.getAuditEvent).not.toHaveBeenCalled()
  })
})

describe('useAuditStats', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch audit statistics', async () => {
    const mockStats = {
      totalEvents: 100,
      eventsByCategory: {
        [AuditCategory.Authentication]: 30,
        [AuditCategory.Workflow]: 40,
      },
      eventsBySeverity: {
        [AuditSeverity.Info]: 80,
        [AuditSeverity.Warning]: 15,
      },
      eventsByStatus: {
        [AuditStatus.Success]: 90,
        [AuditStatus.Failure]: 10,
      },
      topUsers: [],
      topActions: [],
      criticalEvents: 5,
      failedEvents: 10,
      recentCritical: [],
      timeRange: {
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-31T23:59:59Z',
      },
    }

    vi.mocked(auditAPI.getAuditStats).mockResolvedValue(mockStats)

    const { result } = renderHook(
      () =>
        useAuditStats({
          startDate: '2024-01-01T00:00:00Z',
          endDate: '2024-01-31T23:59:59Z',
        }),
      { wrapper: createWrapper() }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.totalEvents).toBe(100)
    expect(result.current.data?.criticalEvents).toBe(5)
  })
})

describe('useExportAudit', () => {
  let mockLink: { click: ReturnType<typeof vi.fn>; href: string; download: string }
  let createElementSpy: ReturnType<typeof vi.spyOn>
  let appendChildSpy: ReturnType<typeof vi.spyOn>
  let removeChildSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    vi.clearAllMocks()

    // Create mock link element
    mockLink = {
      click: vi.fn(),
      href: '',
      download: '',
    }

    // Only mock createElement for 'a' elements, let others through
    const originalCreateElement = document.createElement.bind(document)
    createElementSpy = vi.spyOn(document, 'createElement').mockImplementation((tagName: string) => {
      if (tagName === 'a') {
        return mockLink as any
      }
      return originalCreateElement(tagName)
    })

    // Only mock appendChild/removeChild for our mock link
    const originalAppendChild = document.body.appendChild.bind(document.body)
    appendChildSpy = vi.spyOn(document.body, 'appendChild').mockImplementation((node: Node) => {
      if (node === mockLink) {
        return mockLink as any
      }
      return originalAppendChild(node)
    })

    const originalRemoveChild = document.body.removeChild.bind(document.body)
    removeChildSpy = vi.spyOn(document.body, 'removeChild').mockImplementation((node: Node) => {
      if (node === mockLink) {
        return mockLink as any
      }
      return originalRemoveChild(node)
    })

    vi.spyOn(window.URL, 'createObjectURL').mockReturnValue('blob:mock-url')
    vi.spyOn(window.URL, 'revokeObjectURL').mockImplementation(() => {})
  })

  afterEach(() => {
    createElementSpy?.mockRestore()
    appendChildSpy?.mockRestore()
    removeChildSpy?.mockRestore()
  })

  it('should export audit events and trigger download', async () => {
    const mockBlob = new Blob(['csv,data'], { type: 'text/csv' })
    vi.mocked(auditAPI.exportAuditEvents).mockResolvedValue(mockBlob)

    const { result } = renderHook(() => useExportAudit(), {
      wrapper: createWrapper(),
    })

    result.current.mutate({
      filter: { limit: 100 },
      format: ExportFormat.CSV,
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(auditAPI.exportAuditEvents).toHaveBeenCalledWith(
      { limit: 100 },
      ExportFormat.CSV
    )
    expect(document.createElement).toHaveBeenCalledWith('a')
    expect(document.body.appendChild).toHaveBeenCalled()
  })

  it('should use custom filename when provided', async () => {
    const mockBlob = new Blob(['json data'], { type: 'application/json' })
    vi.mocked(auditAPI.exportAuditEvents).mockResolvedValue(mockBlob)

    // Use the shared mockLink from beforeEach
    mockLink.click = vi.fn()
    mockLink.href = ''
    mockLink.download = ''

    const { result } = renderHook(() => useExportAudit(), {
      wrapper: createWrapper(),
    })

    result.current.mutate({
      filter: {},
      format: ExportFormat.JSON,
      filename: 'custom-export.json',
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(mockLink.download).toBe('custom-export.json')
  })
})

describe('useAuditCategories', () => {
  it('should return available categories', () => {
    vi.mocked(auditAPI.getCategories).mockReturnValue([
      'authentication',
      'workflow',
      'credential',
    ])

    const { result } = renderHook(() => useAuditCategories(), { wrapper: createWrapper() })

    expect(result.current.data).toContain('authentication')
    expect(result.current.data).toContain('workflow')
    expect(result.current.isLoading).toBe(false)
  })
})

describe('useAuditEventTypes', () => {
  it('should return available event types', () => {
    vi.mocked(auditAPI.getEventTypes).mockReturnValue([
      'create',
      'login',
      'execute',
    ])

    const { result } = renderHook(() => useAuditEventTypes(), { wrapper: createWrapper() })

    expect(result.current.data).toContain('create')
    expect(result.current.data).toContain('login')
    expect(result.current.isLoading).toBe(false)
  })
})
