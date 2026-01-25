import { describe, it, expect, beforeEach, vi } from 'vitest'
import { scheduleTemplatesApi } from './scheduleTemplates'
import type {
  ScheduleTemplate,
  ScheduleTemplateFilter,
  ApplyTemplateInput,
  Schedule,
} from './scheduleTemplates'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('Schedule Templates API', () => {
  const mockTemplate: ScheduleTemplate = {
    id: 'tpl-123',
    name: 'Daily at Midnight',
    description: 'Runs every day at midnight',
    category: 'daily',
    cron_expression: '0 0 * * *',
    timezone: 'UTC',
    tags: ['daily', 'midnight'],
    is_system: true,
    created_at: '2024-01-15T10:00:00Z',
  }

  const mockSchedule: Schedule = {
    id: 'sched-123',
    tenant_id: 'tenant-1',
    workflow_id: 'wf-123',
    name: 'Daily Data Sync',
    cron_expression: '0 0 * * *',
    timezone: 'America/New_York',
    enabled: true,
    next_run_at: '2024-01-16T05:00:00Z',
    created_by: 'user-1',
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should fetch all templates without filter', async () => {
      const templates = [mockTemplate]
      ;(apiClient.get as any).mockResolvedValueOnce(templates)

      const result = await scheduleTemplatesApi.list()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedule-templates')
      expect(result).toEqual(templates)
    })

    it('should fetch templates with category filter', async () => {
      (apiClient.get as any).mockResolvedValueOnce([mockTemplate])

      const filter: ScheduleTemplateFilter = { category: 'daily' }
      await scheduleTemplatesApi.list(filter)

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedule-templates?category=daily')
    })

    it('should fetch templates with tags filter', async () => {
      (apiClient.get as any).mockResolvedValueOnce([mockTemplate])

      const filter: ScheduleTemplateFilter = { tags: ['daily', 'midnight'] }
      await scheduleTemplatesApi.list(filter)

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/schedule-templates?tags=daily%2Cmidnight'
      )
    })

    it('should fetch system templates only', async () => {
      (apiClient.get as any).mockResolvedValueOnce([mockTemplate])

      const filter: ScheduleTemplateFilter = { is_system: true }
      await scheduleTemplatesApi.list(filter)

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedule-templates?is_system=true')
    })

    it('should fetch custom templates only', async () => {
      (apiClient.get as any).mockResolvedValueOnce([])

      const filter: ScheduleTemplateFilter = { is_system: false }
      await scheduleTemplatesApi.list(filter)

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedule-templates?is_system=false')
    })

    it('should fetch templates with search', async () => {
      (apiClient.get as any).mockResolvedValueOnce([mockTemplate])

      const filter: ScheduleTemplateFilter = { search: 'midnight' }
      await scheduleTemplatesApi.list(filter)

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedule-templates?search=midnight')
    })

    it('should fetch templates with all filters', async () => {
      (apiClient.get as any).mockResolvedValueOnce([mockTemplate])

      const filter: ScheduleTemplateFilter = {
        category: 'daily',
        tags: ['midnight'],
        is_system: true,
        search: 'daily',
      }
      await scheduleTemplatesApi.list(filter)

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/schedule-templates?category=daily&tags=midnight&is_system=true&search=daily'
      )
    })

    it('should handle empty list', async () => {
      (apiClient.get as any).mockResolvedValueOnce([])

      const result = await scheduleTemplatesApi.list()

      expect(result).toEqual([])
    })

    it('should return templates with different categories', async () => {
      const templates: ScheduleTemplate[] = [
        { ...mockTemplate, id: 'tpl-1', category: 'daily' },
        { ...mockTemplate, id: 'tpl-2', category: 'weekly', name: 'Weekly on Monday' },
        { ...mockTemplate, id: 'tpl-3', category: 'monthly', name: 'First of Month' },
      ]
      ;(apiClient.get as any).mockResolvedValueOnce(templates)

      const result = await scheduleTemplatesApi.list()

      expect(result).toHaveLength(3)
      expect(result.map((t) => t.category)).toEqual(['daily', 'weekly', 'monthly'])
    })

    it('should handle API error', async () => {
      const error = new Error('Network error')
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(scheduleTemplatesApi.list()).rejects.toThrow('Network error')
    })
  })

  describe('get', () => {
    it('should fetch single template by ID', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTemplate)

      const result = await scheduleTemplatesApi.get('tpl-123')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/schedule-templates/tpl-123')
      expect(result).toEqual(mockTemplate)
    })

    it('should return template with cron expression', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTemplate)

      const result = await scheduleTemplatesApi.get('tpl-123')

      expect(result.cron_expression).toBe('0 0 * * *')
      expect(result.timezone).toBe('UTC')
    })

    it('should handle not found error', async () => {
      const error = new Error('Template not found')
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(scheduleTemplatesApi.get('invalid')).rejects.toThrow('Template not found')
    })
  })

  describe('apply', () => {
    it('should apply template to create schedule', async () => {
      const input: ApplyTemplateInput = {
        workflow_id: 'wf-123',
      }
      ;(apiClient.post as any).mockResolvedValueOnce(mockSchedule)

      const result = await scheduleTemplatesApi.apply('tpl-123', input)

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/v1/schedule-templates/tpl-123/apply',
        input
      )
      expect(result).toEqual(mockSchedule)
    })

    it('should apply template with custom name', async () => {
      const input: ApplyTemplateInput = {
        workflow_id: 'wf-123',
        name: 'My Custom Schedule',
      }
      const schedule = { ...mockSchedule, name: 'My Custom Schedule' }
      ;(apiClient.post as any).mockResolvedValueOnce(schedule)

      const result = await scheduleTemplatesApi.apply('tpl-123', input)

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/v1/schedule-templates/tpl-123/apply',
        input
      )
      expect(result.name).toBe('My Custom Schedule')
    })

    it('should apply template with custom timezone', async () => {
      const input: ApplyTemplateInput = {
        workflow_id: 'wf-123',
        timezone: 'America/New_York',
      }
      const schedule = { ...mockSchedule, timezone: 'America/New_York' }
      ;(apiClient.post as any).mockResolvedValueOnce(schedule)

      const result = await scheduleTemplatesApi.apply('tpl-123', input)

      expect(result.timezone).toBe('America/New_York')
    })

    it('should apply template with all options', async () => {
      const input: ApplyTemplateInput = {
        workflow_id: 'wf-123',
        name: 'Daily Sync - East Coast',
        timezone: 'America/New_York',
      }
      ;(apiClient.post as any).mockResolvedValueOnce(mockSchedule)

      await scheduleTemplatesApi.apply('tpl-123', input)

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/v1/schedule-templates/tpl-123/apply',
        input
      )
    })

    it('should return schedule with next run time', async () => {
      (apiClient.post as any).mockResolvedValueOnce(mockSchedule)

      const result = await scheduleTemplatesApi.apply('tpl-123', { workflow_id: 'wf-123' })

      expect(result.next_run_at).toBeDefined()
      expect(result.enabled).toBe(true)
    })

    it('should handle template not found error', async () => {
      const error = new Error('Template not found')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(
        scheduleTemplatesApi.apply('invalid', { workflow_id: 'wf-123' })
      ).rejects.toThrow('Template not found')
    })

    it('should handle workflow not found error', async () => {
      const error = new Error('Workflow not found')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(
        scheduleTemplatesApi.apply('tpl-123', { workflow_id: 'invalid' })
      ).rejects.toThrow('Workflow not found')
    })

    it('should handle validation error', async () => {
      const error = new Error('Workflow ID is required')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(
        scheduleTemplatesApi.apply('tpl-123', { workflow_id: '' })
      ).rejects.toThrow('Workflow ID is required')
    })
  })
})
