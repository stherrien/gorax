import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useCredentialStore } from './credentialStore'
import { credentialAPI } from '../api/credentials'
import type { Credential } from '../api/credentials'

vi.mock('../api/credentials')

describe('credentialStore', () => {
  const mockCredential: Credential = {
    id: 'cred-123',
    tenantId: 'tenant-1',
    name: 'Test Credential',
    type: 'api_key',
    description: 'Test description',
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
  }

  const mockCredential2: Credential = {
    id: 'cred-456',
    tenantId: 'tenant-1',
    name: 'OAuth Credential',
    type: 'oauth2',
    createdAt: '2024-01-15T11:00:00Z',
    updatedAt: '2024-01-15T11:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Reset store state
    const { result } = renderHook(() => useCredentialStore())
    act(() => {
      result.current.reset()
    })
  })

  describe('Initial state', () => {
    it('should have empty credentials array', () => {
      const { result } = renderHook(() => useCredentialStore())
      expect(result.current.credentials).toEqual([])
    })

    it('should not be loading initially', () => {
      const { result } = renderHook(() => useCredentialStore())
      expect(result.current.loading).toBe(false)
    })

    it('should have no error initially', () => {
      const { result } = renderHook(() => useCredentialStore())
      expect(result.current.error).toBeNull()
    })

    it('should have no selected credential initially', () => {
      const { result } = renderHook(() => useCredentialStore())
      expect(result.current.selectedCredential).toBeNull()
    })
  })

  describe('fetchCredentials', () => {
    it('should fetch and set credentials', async () => {
      const mockResponse = {
        credentials: [mockCredential, mockCredential2],
        total: 2,
      }
      ;(credentialAPI.list as any).mockResolvedValueOnce(mockResponse)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.fetchCredentials()
      })

      expect(credentialAPI.list).toHaveBeenCalledWith(undefined)
      expect(result.current.credentials).toEqual([mockCredential, mockCredential2])
      expect(result.current.loading).toBe(false)
      expect(result.current.error).toBeNull()
    })

    it('should set loading state during fetch', async () => {
      (credentialAPI.list as any).mockImplementationOnce(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      )

      const { result } = renderHook(() => useCredentialStore())

      act(() => {
        result.current.fetchCredentials()
      })

      expect(result.current.loading).toBe(true)
    })

    it('should handle fetch errors', async () => {
      const error = new Error('Failed to fetch')
      ;(credentialAPI.list as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.fetchCredentials()
      })

      expect(result.current.error).toBe('Failed to fetch')
      expect(result.current.loading).toBe(false)
      expect(result.current.credentials).toEqual([])
    })

    it('should support filter parameters', async () => {
      const mockResponse = { credentials: [mockCredential], total: 1 }
      ;(credentialAPI.list as any).mockResolvedValueOnce(mockResponse)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.fetchCredentials({ type: 'api_key', search: 'test' })
      })

      expect(credentialAPI.list).toHaveBeenCalledWith({ type: 'api_key', search: 'test' })
    })
  })

  describe('fetchCredential', () => {
    it('should fetch and set selected credential', async () => {
      (credentialAPI.get as any).mockResolvedValueOnce(mockCredential)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.fetchCredential('cred-123')
      })

      expect(credentialAPI.get).toHaveBeenCalledWith('cred-123')
      expect(result.current.selectedCredential).toEqual(mockCredential)
      expect(result.current.error).toBeNull()
    })

    it('should handle fetch errors', async () => {
      const error = new Error('Credential not found')
      ;(credentialAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.fetchCredential('invalid-id')
      })

      expect(result.current.error).toBe('Credential not found')
      expect(result.current.selectedCredential).toBeNull()
    })
  })

  describe('createCredential', () => {
    it('should create new credential', async () => {
      const newCredential = {
        name: 'New Credential',
        type: 'api_key' as const,
        value: { apiKey: 'key-123' },
      }

      ;(credentialAPI.create as any).mockResolvedValueOnce(mockCredential)
      ;(credentialAPI.list as any).mockResolvedValueOnce({
        credentials: [mockCredential],
        total: 1,
      })

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.createCredential(newCredential)
      })

      expect(credentialAPI.create).toHaveBeenCalledWith(newCredential)
      expect(credentialAPI.list).toHaveBeenCalled()
    })

    it('should handle creation errors', async () => {
      const error = new Error('Validation failed')
      ;(credentialAPI.create as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.createCredential({
          name: 'Invalid',
          type: 'api_key',
          value: {},
        })
      })

      expect(result.current.error).toBe('Validation failed')
    })
  })

  describe('updateCredential', () => {
    it('should update existing credential', async () => {
      const updates = {
        name: 'Updated Name',
        description: 'Updated description',
      }

      const updatedCredential = { ...mockCredential, ...updates }
      ;(credentialAPI.update as any).mockResolvedValueOnce(updatedCredential)
      ;(credentialAPI.list as any).mockResolvedValueOnce({
        credentials: [updatedCredential],
        total: 1,
      })

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.updateCredential('cred-123', updates)
      })

      expect(credentialAPI.update).toHaveBeenCalledWith('cred-123', updates)
      expect(credentialAPI.list).toHaveBeenCalled()
    })

    it('should handle update errors', async () => {
      const error = new Error('Not found')
      ;(credentialAPI.update as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.updateCredential('invalid-id', { name: 'Test' })
      })

      expect(result.current.error).toBe('Not found')
    })
  })

  describe('deleteCredential', () => {
    it('should delete credential', async () => {
      (credentialAPI.delete as any).mockResolvedValueOnce(undefined)
      ;(credentialAPI.list as any).mockResolvedValueOnce({
        credentials: [],
        total: 0,
      })

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.deleteCredential('cred-123')
      })

      expect(credentialAPI.delete).toHaveBeenCalledWith('cred-123')
      expect(credentialAPI.list).toHaveBeenCalled()
    })

    it('should handle deletion errors', async () => {
      const error = new Error('Not found')
      ;(credentialAPI.delete as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.deleteCredential('invalid-id')
      })

      expect(result.current.error).toBe('Not found')
    })
  })

  describe('rotateCredential', () => {
    it('should rotate credential value', async () => {
      const newValue = { apiKey: 'new-key-456' }
      const rotatedCredential = {
        ...mockCredential,
        updatedAt: '2024-01-16T10:00:00Z',
      }

      ;(credentialAPI.rotate as any).mockResolvedValueOnce(rotatedCredential)
      ;(credentialAPI.list as any).mockResolvedValueOnce({
        credentials: [rotatedCredential],
        total: 1,
      })

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.rotateCredential('cred-123', newValue)
      })

      expect(credentialAPI.rotate).toHaveBeenCalledWith('cred-123', newValue)
      expect(credentialAPI.list).toHaveBeenCalled()
    })

    it('should handle rotation errors', async () => {
      const error = new Error('Rotation failed')
      ;(credentialAPI.rotate as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.rotateCredential('cred-123', { apiKey: 'key' })
      })

      expect(result.current.error).toBe('Rotation failed')
    })
  })

  describe('testCredential', () => {
    it('should test credential and return result', async () => {
      const testResult = {
        success: true,
        message: 'Connection successful',
        testedAt: '2024-01-15T10:00:00Z',
      }

      ;(credentialAPI.test as any).mockResolvedValueOnce(testResult)

      const { result } = renderHook(() => useCredentialStore())

      let testResponse
      await act(async () => {
        testResponse = await result.current.testCredential('cred-123')
      })

      expect(credentialAPI.test).toHaveBeenCalledWith('cred-123')
      expect(testResponse).toEqual(testResult)
      expect(result.current.error).toBeNull()
    })

    it('should handle test failures', async () => {
      const testResult = {
        success: false,
        message: 'Authentication failed',
        testedAt: '2024-01-15T10:00:00Z',
      }

      ;(credentialAPI.test as any).mockResolvedValueOnce(testResult)

      const { result } = renderHook(() => useCredentialStore())

      let testResponse
      await act(async () => {
        testResponse = await result.current.testCredential('cred-123')
      })

      expect(testResponse.success).toBe(false)
      expect(testResponse.message).toBe('Authentication failed')
    })

    it('should handle test errors', async () => {
      const error = new Error('Network error')
      ;(credentialAPI.test as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        try {
          await result.current.testCredential('cred-123')
        } catch (err) {
          // Expected to throw
        }
      })

      expect(result.current.error).toBe('Network error')
    })
  })

  describe('selectCredential', () => {
    it('should set selected credential', () => {
      const { result } = renderHook(() => useCredentialStore())

      act(() => {
        result.current.setCredentials([mockCredential, mockCredential2])
        result.current.selectCredential('cred-123')
      })

      expect(result.current.selectedCredential).toEqual(mockCredential)
    })

    it('should clear selected credential when null', () => {
      const { result } = renderHook(() => useCredentialStore())

      act(() => {
        result.current.setCredentials([mockCredential])
        result.current.selectCredential('cred-123')
      })

      expect(result.current.selectedCredential).toEqual(mockCredential)

      act(() => {
        result.current.selectCredential(null)
      })

      expect(result.current.selectedCredential).toBeNull()
    })

    it('should return null for non-existent credential', () => {
      const { result } = renderHook(() => useCredentialStore())

      act(() => {
        result.current.setCredentials([mockCredential])
        result.current.selectCredential('non-existent')
      })

      expect(result.current.selectedCredential).toBeNull()
    })
  })

  describe('setCredentials', () => {
    it('should set credentials array', () => {
      const { result } = renderHook(() => useCredentialStore())

      act(() => {
        result.current.setCredentials([mockCredential, mockCredential2])
      })

      expect(result.current.credentials).toEqual([mockCredential, mockCredential2])
    })
  })

  describe('clearError', () => {
    it('should clear error state', async () => {
      const error = new Error('Test error')
      ;(credentialAPI.list as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.fetchCredentials()
      })

      expect(result.current.error).toBe('Test error')

      act(() => {
        result.current.clearError()
      })

      expect(result.current.error).toBeNull()
    })
  })

  describe('reset', () => {
    it('should reset store to initial state', async () => {
      const mockResponse = {
        credentials: [mockCredential],
        total: 1,
      }
      ;(credentialAPI.list as any).mockResolvedValueOnce(mockResponse)

      const { result } = renderHook(() => useCredentialStore())

      await act(async () => {
        await result.current.fetchCredentials()
        result.current.selectCredential('cred-123')
      })

      expect(result.current.credentials).toEqual([mockCredential])
      expect(result.current.selectedCredential).toEqual(mockCredential)

      act(() => {
        result.current.reset()
      })

      expect(result.current.credentials).toEqual([])
      expect(result.current.selectedCredential).toBeNull()
      expect(result.current.error).toBeNull()
      expect(result.current.loading).toBe(false)
    })
  })
})
