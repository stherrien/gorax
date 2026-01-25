import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useSchedules, useSchedule, useScheduleMutations } from './useSchedules'
import { scheduleAPI } from '../api/schedules'
import { createQueryWrapper } from '../test/test-utils'

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

describe('useSchedules', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch schedules on mount', async () => {
    const mockSchedules = [
      {
        id: 'sched-1',
        name: 'Daily Backup',
        cronExpression: '0 0 * * *',
        enabled: true,
      },
    ]

    vi.mocked(scheduleAPI.list).mockResolvedValue({
      schedules: mockSchedules,
      total: 1,
    })

    const { result } = renderHook(() => useSchedules(), { wrapper: createQueryWrapper() })

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

    const { result } = renderHook(() => useSchedules(), { wrapper: createQueryWrapper() })

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
      { initialProps: { params: { page: 1 } }, wrapper: createQueryWrapper() }
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
  // Valid RFC 4122 UUID (version 4, variant 1)
  const validUUID = '12345678-1234-4234-8234-123456789abc'

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch single schedule', async () => {
    const mockSchedule = {
      id: validUUID,
      name: 'Daily Backup',
      cronExpression: '0 0 * * *',
      enabled: true,
    }

    vi.mocked(scheduleAPI.get).mockResolvedValue(mockSchedule)

    const { result } = renderHook(() => useSchedule(validUUID), { wrapper: createQueryWrapper() })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.schedule).toEqual(mockSchedule)
    expect(result.current.error).toBeNull()
  })

  it('should not fetch when id is null', () => {
    const { result } = renderHook(() => useSchedule(null), { wrapper: createQueryWrapper() })

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

    const mockSchedule = { id: 'sched-new', ...input }
    vi.mocked(scheduleAPI.create).mockResolvedValue(mockSchedule)

    const { result } = renderHook(() => useScheduleMutations(), { wrapper: createQueryWrapper() })

    expect(result.current.creating).toBe(false)

    const promise = result.current.createSchedule('wf-1', input)

    await waitFor(() => {
      expect(result.current.creating).toBe(false)
    })

    const created = await promise
    expect(created).toEqual(mockSchedule)
  })

  it('should update schedule', async () => {
    const updates = { name: 'Updated' }
    const mockSchedule = { id: 'sched-1', ...updates }

    vi.mocked(scheduleAPI.update).mockResolvedValue(mockSchedule)

    const { result } = renderHook(() => useScheduleMutations(), { wrapper: createQueryWrapper() })

    const updated = await result.current.updateSchedule('sched-1', updates)

    expect(updated).toEqual(mockSchedule)
  })

  it('should delete schedule', async () => {
    vi.mocked(scheduleAPI.delete).mockResolvedValue(undefined)

    const { result } = renderHook(() => useScheduleMutations(), { wrapper: createQueryWrapper() })

    await result.current.deleteSchedule('sched-1')

    expect(scheduleAPI.delete).toHaveBeenCalledWith('sched-1')
  })

  it('should toggle schedule', async () => {
    const mockSchedule = { id: 'sched-1', enabled: false }
    vi.mocked(scheduleAPI.toggle).mockResolvedValue(mockSchedule)

    const { result } = renderHook(() => useScheduleMutations(), { wrapper: createQueryWrapper() })

    const toggled = await result.current.toggleSchedule('sched-1', false)

    expect(toggled.enabled).toBe(false)
  })
})
