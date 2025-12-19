import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  APIClient,
  APIError,
  AuthError,
  NotFoundError,
  ValidationError,
  ServerError,
  NetworkError,
} from './client'

describe('APIClient', () => {
  let client: APIClient

  beforeEach(() => {
    client = new APIClient('http://localhost:8080')
    vi.clearAllMocks()
    ;(global.fetch as any).mockClear()
  })

  describe('GET requests', () => {
    it('should make successful GET request', async () => {
      const mockData = { id: '123', name: 'Test' }
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => mockData,
      })

      const result = await client.get('/test')

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/test',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      )
      expect(result).toEqual(mockData)
    })

    it('should include auth token in header when available', async () => {
      ;(localStorage.getItem as any).mockReturnValueOnce('test-token')
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({}),
      })

      await client.get('/test')

      expect(global.fetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer test-token',
          }),
        })
      )
    })

    it('should handle query parameters', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({}),
      })

      await client.get('/test', { params: { page: 1, limit: 10 } })

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/test?page=1&limit=10',
        expect.any(Object)
      )
    })
  })

  describe('POST requests', () => {
    it('should make successful POST request with body', async () => {
      const requestBody = { name: 'New Item' }
      const responseData = { id: '123', ...requestBody }
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => responseData,
      })

      const result = await client.post('/items', requestBody)

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/items',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(requestBody),
        })
      )
      expect(result).toEqual(responseData)
    })
  })

  describe('PUT requests', () => {
    it('should make successful PUT request', async () => {
      const updates = { name: 'Updated' }
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => updates,
      })

      const result = await client.put('/items/123', updates)

      expect(result).toEqual(updates)
    })
  })

  describe('DELETE requests', () => {
    it('should make successful DELETE request', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: async () => ({}),
      })

      await client.delete('/items/123')

      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/items/123',
        expect.objectContaining({ method: 'DELETE' })
      )
    })
  })

  describe('Error handling', () => {
    it('should throw AuthError on 401 response', async () => {
      ;(global.fetch as any).mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({ error: 'Unauthorized' }),
      })

      await expect(client.get('/test')).rejects.toThrow(AuthError)
      await expect(client.get('/test')).rejects.toThrow('Unauthorized')
    })

    it('should throw AuthError on 403 response', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 403,
        json: async () => ({ error: 'Forbidden' }),
      })

      await expect(client.get('/test')).rejects.toThrow(AuthError)
    })

    it('should throw NotFoundError on 404 response', async () => {
      ;(global.fetch as any).mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({ error: 'Not found' }),
      })

      await expect(client.get('/test')).rejects.toThrow(NotFoundError)
      await expect(client.get('/test')).rejects.toThrow('Not found')
    })

    it('should throw ValidationError on 400 response', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({ error: 'Invalid input' }),
      })

      await expect(client.post('/test', {})).rejects.toThrow(ValidationError)
    })

    it('should throw ValidationError on 422 response', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 422,
        json: async () => ({ error: 'Validation failed' }),
      })

      await expect(client.post('/test', {})).rejects.toThrow(ValidationError)
    })

    it('should throw ServerError on 500 response', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({ error: 'Internal server error' }),
      })

      await expect(client.get('/test')).rejects.toThrow(ServerError)
    })

    it('should throw ServerError on 503 response', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 503,
        json: async () => ({ error: 'Service unavailable' }),
      })

      await expect(client.get('/test')).rejects.toThrow(ServerError)
    })

    it('should throw NetworkError when fetch fails', async () => {
      ;(global.fetch as any).mockRejectedValue(new Error('Network failure'))

      await expect(client.get('/test')).rejects.toThrow(NetworkError)
      await expect(client.get('/test')).rejects.toThrow('Network failure')
    })

    it('should handle non-JSON error responses', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => {
          throw new Error('Invalid JSON')
        },
        text: async () => 'Internal Server Error',
      })

      await expect(client.get('/test')).rejects.toThrow(ServerError)
    })
  })

  describe('Retry logic', () => {
    it('should retry on 503 errors', async () => {
      ;(global.fetch as any)
        .mockResolvedValueOnce({
          ok: false,
          status: 503,
          json: async () => ({ error: 'Service unavailable' }),
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => ({ success: true }),
        })

      const result = await client.get('/test', { retries: 1 })

      expect(global.fetch).toHaveBeenCalledTimes(2)
      expect(result).toEqual({ success: true })
    })

    it('should not retry on 4xx errors', async () => {
      ;(global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: async () => ({ error: 'Not found' }),
      })

      await expect(client.get('/test', { retries: 3 })).rejects.toThrow(NotFoundError)
      expect(global.fetch).toHaveBeenCalledTimes(1)
    })

    it('should exhaust retry attempts and throw error', async () => {
      ;(global.fetch as any).mockResolvedValue({
        ok: false,
        status: 503,
        json: async () => ({ error: 'Service unavailable' }),
      })

      await expect(client.get('/test', { retries: 2 })).rejects.toThrow(ServerError)
      expect(global.fetch).toHaveBeenCalledTimes(3) // initial + 2 retries
    })
  })

  describe('Timeout handling', () => {
    // Skip: happy-dom doesn't fully support AbortController
    it.skip('should timeout long requests', async () => {
      ;(global.fetch as any).mockImplementationOnce(
        () =>
          new Promise((resolve) =>
            setTimeout(() => resolve({ ok: true, json: async () => ({}) }), 5000)
          )
      )

      await expect(client.get('/test', { timeout: 50 })).rejects.toThrow('Request timeout')
    })
  })
})
