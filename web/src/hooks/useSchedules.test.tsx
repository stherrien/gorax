import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useSchedules, useSchedule, useScheduleMutations } from './useSchedules'
import { scheduleAPI } from '../api/schedules'

vi.mock('../api/schedules', () => ({
  scheduleAPI: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    toggle: vi.fn(),
  },
}))

// Helper to create a wrapper with QueryClient
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

// Valid UUIDs for tests (RFC 4122 compliant)
const mockScheduleId = 'a1b2c3d4-e5f6-4890-abcd-ef1234567890'
const mockWorkflowId = 'b2c3d4e5-f6a7-4901-bcde-f12345678901'

describe('useSchedules', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch schedules on mount', async () => {
    const mockSchedules = [
      {
        id: mockScheduleId,
        name: 'Daily Backup',
        cronExpression: '0 0 * * *',
        enabled: true,
      },
    ]

    vi.mocked(scheduleAPI.list).mockResolvedValue({
      schedules: mockSchedules,
      total: 1,
    })

    const { result } = renderHook(() => useSchedules(), { wrapper: createWrapper() })

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.schedules).toEqual(mockSchedules)
    expect(result.current.total).toBe(1)
    expect(result.current.error).toBeNull()
  })

  it('should handle fetch error', async () => {
    const error = new Error('Fetch failed')
    vi.mocked(scheduleAPI.list).mockRejectedValue(error)

    const { result } = renderHook(() => useSchedules(), { wrapper: createWrapper() })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.schedules).toEqual([])
    expect(result.current.error).toEqual(error)
  })

  it('should refetch schedules when params change', async () => {
    vi.mocked(scheduleAPI.list).mockResolvedValue({
      schedules: [],
      total: 0,
    })

    const { rerender } = renderHook(
      ({ params }) => useSchedules(params),
      { initialProps: { params: { page: 1 } }, wrapper: createWrapper() }
    )

    await waitFor(() => {
      expect(scheduleAPI.list).toHaveBeenCalledWith({ page: 1 })
    })

    rerender({ params: { page: 2 } })

    await waitFor(() => {
      expect(scheduleAPI.list).toHaveBeenCalledWith({ page: 2 })
    })
  })
})

describe('useSchedule', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch single schedule', async () => {
    const mockSchedule = {
      id: mockScheduleId,
      name: 'Daily Backup',
      cronExpression: '0 0 * * *',
      enabled: true,
    }

    vi.mocked(scheduleAPI.get).mockResolvedValue(mockSchedule)

    const { result } = renderHook(() => useSchedule(mockScheduleId), { wrapper: createWrapper() })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.schedule).toEqual(mockSchedule)
    expect(result.current.error).toBeNull()
  })

  it('should not fetch when id is null', () => {
    const { result } = renderHook(() => useSchedule(null), { wrapper: createWrapper() })

    expect(result.current.loading).toBe(false)
    expect(result.current.schedule).toBeNull()
    expect(scheduleAPI.get).not.toHaveBeenCalled()
  })
})

describe('useScheduleMutations', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should create schedule', async () => {
    const input = {
      name: 'New Schedule',
      cronExpression: '0 0 * * *',
      enabled: true,
    }

    const newScheduleId = 'c3d4e5f6-a7b8-4012-8def-123456789012'
    const mockSchedule = { id: newScheduleId, ...input }
    vi.mocked(scheduleAPI.create).mockResolvedValue(mockSchedule)

    const { result } = renderHook(() => useScheduleMutations(), { wrapper: createWrapper() })

    expect(result.current.creating).toBe(false)

    const promise = result.current.createSchedule(mockWorkflowId, input)

    await waitFor(() => {
      expect(result.current.creating).toBe(false)
    })

    const created = await promise
    expect(created).toEqual(mockSchedule)
  })

  it('should update schedule', async () => {
    const updates = { name: 'Updated' }
    const mockSchedule = { id: mockScheduleId, ...updates }

    vi.mocked(scheduleAPI.update).mockResolvedValue(mockSchedule)

    const { result } = renderHook(() => useScheduleMutations(), { wrapper: createWrapper() })

    const updated = await result.current.updateSchedule(mockScheduleId, updates)

    expect(updated).toEqual(mockSchedule)
  })

  it('should delete schedule', async () => {
    vi.mocked(scheduleAPI.delete).mockResolvedValue(undefined)

    const { result } = renderHook(() => useScheduleMutations(), { wrapper: createWrapper() })

    await result.current.deleteSchedule(mockScheduleId)

    expect(scheduleAPI.delete).toHaveBeenCalledWith(mockScheduleId)
  })

  it('should toggle schedule', async () => {
    const mockSchedule = { id: mockScheduleId, enabled: false }
    vi.mocked(scheduleAPI.toggle).mockResolvedValue(mockSchedule)

    const { result } = renderHook(() => useScheduleMutations(), { wrapper: createWrapper() })

    const toggled = await result.current.toggleSchedule(mockScheduleId, false)

    expect(toggled.enabled).toBe(false)
  })
})
