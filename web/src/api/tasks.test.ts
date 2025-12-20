import { describe, it, expect, vi, beforeEach } from 'vitest'
import { tasksApi } from './tasks'
import { apiClient } from './client'

vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

const mockedApiClient = vi.mocked(apiClient)

describe('tasksApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should fetch tasks with filters', async () => {
      const mockResponse = {
        tasks: [
          {
            id: '123',
            title: 'Test Task',
            status: 'pending',
          },
        ],
        count: 1,
      }

      mockedApiClient.get.mockResolvedValue(mockResponse)

      const params = { status: 'pending', limit: 10 }
      const result = await tasksApi.list(params)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v1/tasks', { params })
      expect(result).toEqual(mockResponse)
    })

    it('should fetch tasks without filters', async () => {
      const mockResponse = { tasks: [], count: 0 }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      const result = await tasksApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v1/tasks', { params: undefined })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('get', () => {
    it('should fetch a single task by id', async () => {
      const mockTask = {
        id: '123',
        title: 'Test Task',
        status: 'pending',
      }
      mockedApiClient.get.mockResolvedValue(mockTask)

      const result = await tasksApi.get('123')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v1/tasks/123')
      expect(result).toEqual(mockTask)
    })
  })

  describe('approve', () => {
    it('should approve a task with comment', async () => {
      const mockTask = {
        id: '123',
        title: 'Test Task',
        status: 'approved',
      }
      mockedApiClient.post.mockResolvedValue(mockTask)

      const result = await tasksApi.approve('123', { comment: 'Looks good!' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/tasks/123/approve', {
        comment: 'Looks good!',
      })
      expect(result).toEqual(mockTask)
    })

    it('should approve a task without comment', async () => {
      const mockTask = { id: '123', status: 'approved' }
      mockedApiClient.post.mockResolvedValue(mockTask)

      const result = await tasksApi.approve('123')

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/tasks/123/approve', {})
      expect(result).toEqual(mockTask)
    })
  })

  describe('reject', () => {
    it('should reject a task with reason', async () => {
      const mockTask = {
        id: '123',
        title: 'Test Task',
        status: 'rejected',
      }
      mockedApiClient.post.mockResolvedValue(mockTask)

      const result = await tasksApi.reject('123', { reason: 'Not ready' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/tasks/123/reject', {
        reason: 'Not ready',
      })
      expect(result).toEqual(mockTask)
    })

    it('should reject a task without reason', async () => {
      const mockTask = { id: '123', status: 'rejected' }
      mockedApiClient.post.mockResolvedValue(mockTask)

      const result = await tasksApi.reject('123')

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/tasks/123/reject', {})
      expect(result).toEqual(mockTask)
    })
  })

  describe('submit', () => {
    it('should submit task data', async () => {
      const mockTask = {
        id: '123',
        title: 'Test Task',
        status: 'approved',
      }
      mockedApiClient.post.mockResolvedValue(mockTask)

      const submitData = { data: { field1: 'value1', field2: 42 } }
      const result = await tasksApi.submit('123', submitData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/tasks/123/submit', submitData)
      expect(result).toEqual(mockTask)
    })
  })
})
