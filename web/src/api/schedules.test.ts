import { describe, it, expect, beforeEach, vi } from 'vitest'
import { scheduleAPI } from './schedules'
import { apiClient } from './client'

vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('scheduleAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should fetch schedules with params', async () => {
      const mockResponse = {
        data: [
          {
            id: 'sched-1',
            tenantId: 'tenant-1',
            workflowId: 'wf-1',
            name: 'Daily Backup',
            cronExpression: '0 0 * * *',
            timezone: 'UTC',
            enabled: true,
            nextRunAt: '2025-01-20T00:00:00Z',
            lastRunAt: '2025-01-19T00:00:00Z',
            createdAt: '2025-01-01T00:00:00Z',
            updatedAt: '2025-01-01T00:00:00Z',
          },
        ],
        limit: 20,
        offset: 0,
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const result = await scheduleAPI.list({ page: 1, limit: 20 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedules', {
        params: { page: 1, limit: 20 },
      })
      expect(result.schedules).toHaveLength(1)
      expect(result.schedules[0].name).toBe('Daily Backup')
    })

    it('should fetch schedules without params', async () => {
      const mockResponse = { data: [], limit: 20, offset: 0 }
      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      await scheduleAPI.list()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedules', undefined)
    })
  })

  describe('get', () => {
    it('should fetch a single schedule', async () => {
      const mockSchedule = {
        data: {
          id: 'sched-1',
          name: 'Daily Backup',
          cronExpression: '0 0 * * *',
          enabled: true,
        },
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockSchedule)

      const result = await scheduleAPI.get('sched-1')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedules/sched-1')
      expect(result.name).toBe('Daily Backup')
    })
  })

  describe('create', () => {
    it('should create a new schedule', async () => {
      const input = {
        workflowId: 'wf-1',
        name: 'New Schedule',
        cronExpression: '0 0 * * *',
        timezone: 'UTC',
        enabled: true,
      }

      const mockResponse = {
        data: { id: 'sched-new', ...input },
      }

      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await scheduleAPI.create('wf-1', input)

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/v1/workflows/wf-1/schedules',
        input
      )
      expect(result.id).toBe('sched-new')
    })
  })

  describe('update', () => {
    it('should update a schedule', async () => {
      const updates = { name: 'Updated Schedule', enabled: false }
      const mockResponse = {
        data: { id: 'sched-1', ...updates },
      }

      vi.mocked(apiClient.put).mockResolvedValue(mockResponse)

      const result = await scheduleAPI.update('sched-1', updates)

      expect(apiClient.put).toHaveBeenCalledWith('/api/v1/schedules/sched-1', updates)
      expect(result.name).toBe('Updated Schedule')
    })
  })

  describe('delete', () => {
    it('should delete a schedule', async () => {
      vi.mocked(apiClient.delete).mockResolvedValue(undefined)

      await scheduleAPI.delete('sched-1')

      expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/schedules/sched-1')
    })
  })

  describe('toggle', () => {
    it('should toggle schedule enabled state', async () => {
      const mockResponse = {
        data: { id: 'sched-1', enabled: false },
      }

      vi.mocked(apiClient.put).mockResolvedValue(mockResponse)

      const result = await scheduleAPI.toggle('sched-1', false)

      expect(apiClient.put).toHaveBeenCalledWith('/api/v1/schedules/sched-1', {
        enabled: false,
      })
      expect(result.enabled).toBe(false)
    })
  })

  describe('parseCron', () => {
    it('should parse cron expression', async () => {
      const mockResponse = {
        data: {
          valid: true,
          next_run: '2025-01-20T00:00:00Z',
        },
      }

      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await scheduleAPI.parseCron('0 0 * * *', 'UTC')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/schedules/parse-cron', {
        cron_expression: '0 0 * * *',
        timezone: 'UTC',
      })
      expect(result.valid).toBe(true)
    })
  })

  describe('preview', () => {
    it('should preview schedule with default count', async () => {
      const mockResponse = {
        data: {
          valid: true,
          next_runs: [
            '2025-01-20T09:00:00Z',
            '2025-01-21T09:00:00Z',
            '2025-01-22T09:00:00Z',
          ],
          count: 3,
          timezone: 'UTC',
        },
      }

      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await scheduleAPI.preview('0 9 * * *', 'UTC')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/schedules/preview', {
        cron_expression: '0 9 * * *',
        timezone: 'UTC',
        count: 10,
      })
      expect(result.valid).toBe(true)
      expect(result.next_runs).toHaveLength(3)
    })

    it('should preview schedule with custom count', async () => {
      const mockResponse = {
        data: {
          valid: true,
          next_runs: ['2025-01-20T09:00:00Z', '2025-01-21T09:00:00Z'],
          count: 2,
          timezone: 'America/New_York',
        },
      }

      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await scheduleAPI.preview('0 9 * * *', 'America/New_York', 2)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/schedules/preview', {
        cron_expression: '0 9 * * *',
        timezone: 'America/New_York',
        count: 2,
      })
      expect(result.count).toBe(2)
    })

    it('should handle invalid cron expression', async () => {
      vi.mocked(apiClient.post).mockRejectedValue(
        new Error('invalid cron expression')
      )

      await expect(scheduleAPI.preview('invalid', 'UTC')).rejects.toThrow(
        'invalid cron expression'
      )
    })
  })
})
