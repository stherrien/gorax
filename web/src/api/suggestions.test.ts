import { describe, it, expect, beforeEach, vi } from 'vitest'
import { suggestionsApi } from './suggestions'
import type { Suggestion, AnalyzeRequest, SuggestionsListResponse } from '../types/suggestions'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('Suggestions API', () => {
  const mockSuggestion: Suggestion = {
    id: 'sugg-123',
    tenant_id: 'tenant-1',
    execution_id: 'exec-456',
    node_id: 'node-1',
    category: 'network',
    type: 'retry',
    confidence: 'high',
    title: 'Retry the request',
    description: 'The request failed due to a temporary network issue',
    details: 'Consider retrying with exponential backoff',
    fix: {
      action_type: 'retry',
      retry_config: {
        max_retries: 3,
        backoff_ms: 1000,
        backoff_factor: 2,
      },
    },
    source: 'pattern',
    status: 'pending',
    created_at: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should fetch suggestions for execution', async () => {
      const response: SuggestionsListResponse = {
        data: [mockSuggestion],
      }
      ;(apiClient.get as any).mockResolvedValueOnce(response)

      const result = await suggestionsApi.list('exec-456')

      expect(apiClient.get).toHaveBeenCalledWith('/executions/exec-456/suggestions')
      expect(result).toEqual([mockSuggestion])
    })

    it('should handle empty suggestions list', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ data: [] })

      const result = await suggestionsApi.list('exec-456')

      expect(result).toEqual([])
    })

    it('should return suggestions with different categories', async () => {
      const suggestions: Suggestion[] = [
        { ...mockSuggestion, id: 'sugg-1', category: 'network' },
        { ...mockSuggestion, id: 'sugg-2', category: 'auth' },
        { ...mockSuggestion, id: 'sugg-3', category: 'timeout' },
      ]
      ;(apiClient.get as any).mockResolvedValueOnce({ data: suggestions })

      const result = await suggestionsApi.list('exec-456')

      expect(result).toHaveLength(3)
      expect(result.map((s) => s.category)).toEqual(['network', 'auth', 'timeout'])
    })

    it('should return suggestions with different statuses', async () => {
      const suggestions: Suggestion[] = [
        { ...mockSuggestion, id: 'sugg-1', status: 'pending' },
        { ...mockSuggestion, id: 'sugg-2', status: 'applied', applied_at: '2024-01-15T11:00:00Z' },
        { ...mockSuggestion, id: 'sugg-3', status: 'dismissed', dismissed_at: '2024-01-15T12:00:00Z' },
      ]
      ;(apiClient.get as any).mockResolvedValueOnce({ data: suggestions })

      const result = await suggestionsApi.list('exec-456')

      expect(result.map((s) => s.status)).toEqual(['pending', 'applied', 'dismissed'])
    })

    it('should handle API error', async () => {
      const error = new Error('Execution not found')
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(suggestionsApi.list('invalid')).rejects.toThrow('Execution not found')
    })
  })

  describe('get', () => {
    it('should fetch single suggestion by ID', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ data: mockSuggestion })

      const result = await suggestionsApi.get('sugg-123')

      expect(apiClient.get).toHaveBeenCalledWith('/suggestions/sugg-123')
      expect(result).toEqual(mockSuggestion)
    })

    it('should return suggestion with fix data', async () => {
      const suggestionWithFix: Suggestion = {
        ...mockSuggestion,
        fix: {
          action_type: 'config_change',
          config_path: 'timeout',
          old_value: 5000,
          new_value: 30000,
        },
      }
      ;(apiClient.get as any).mockResolvedValueOnce({ data: suggestionWithFix })

      const result = await suggestionsApi.get('sugg-123')

      expect(result.fix).toBeDefined()
      expect(result.fix?.config_path).toBe('timeout')
    })

    it('should handle not found error', async () => {
      const error = new Error('Suggestion not found')
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(suggestionsApi.get('invalid')).rejects.toThrow('Suggestion not found')
    })
  })

  describe('analyze', () => {
    it('should analyze error and generate suggestions', async () => {
      const request: AnalyzeRequest = {
        workflow_id: 'wf-123',
        node_id: 'node-1',
        node_type: 'action:http',
        error_message: 'Connection timeout',
        http_status: 504,
      }
      ;(apiClient.post as any).mockResolvedValueOnce({ data: [mockSuggestion] })

      const result = await suggestionsApi.analyze('exec-456', request)

      expect(apiClient.post).toHaveBeenCalledWith('/executions/exec-456/analyze', request)
      expect(result).toEqual([mockSuggestion])
    })

    it('should analyze with full context', async () => {
      const request: AnalyzeRequest = {
        workflow_id: 'wf-123',
        node_id: 'node-1',
        node_type: 'action:http',
        error_message: 'Authentication failed',
        error_code: 'AUTH_FAILED',
        http_status: 401,
        retry_count: 2,
        input_data: { api_key: '***' },
        node_config: { url: 'https://api.example.com' },
      }
      ;(apiClient.post as any).mockResolvedValueOnce({ data: [mockSuggestion] })

      await suggestionsApi.analyze('exec-456', request)

      expect(apiClient.post).toHaveBeenCalledWith('/executions/exec-456/analyze', request)
    })

    it('should return multiple suggestions for complex errors', async () => {
      const suggestions: Suggestion[] = [
        { ...mockSuggestion, id: 'sugg-1', type: 'retry', confidence: 'high' },
        { ...mockSuggestion, id: 'sugg-2', type: 'config_change', confidence: 'medium' },
      ]
      ;(apiClient.post as any).mockResolvedValueOnce({ data: suggestions })

      const request: AnalyzeRequest = {
        workflow_id: 'wf-123',
        node_id: 'node-1',
        node_type: 'action:http',
        error_message: 'Request failed',
      }

      const result = await suggestionsApi.analyze('exec-456', request)

      expect(result).toHaveLength(2)
    })

    it('should handle LLM-generated suggestions', async () => {
      const llmSuggestion: Suggestion = {
        ...mockSuggestion,
        source: 'llm',
        confidence: 'medium',
        description: 'AI-generated suggestion based on error analysis',
      }
      ;(apiClient.post as any).mockResolvedValueOnce({ data: [llmSuggestion] })

      const request: AnalyzeRequest = {
        workflow_id: 'wf-123',
        node_id: 'node-1',
        node_type: 'action:http',
        error_message: 'Complex error',
      }

      const result = await suggestionsApi.analyze('exec-456', request)

      expect(result[0].source).toBe('llm')
    })
  })

  describe('apply', () => {
    it('should mark suggestion as applied', async () => {
      (apiClient.post as any).mockResolvedValueOnce({})

      await suggestionsApi.apply('sugg-123')

      expect(apiClient.post).toHaveBeenCalledWith('/suggestions/sugg-123/apply', {})
    })

    it('should handle already applied error', async () => {
      const error = new Error('Suggestion already applied')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(suggestionsApi.apply('sugg-123')).rejects.toThrow('Suggestion already applied')
    })

    it('should handle not found error', async () => {
      const error = new Error('Suggestion not found')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(suggestionsApi.apply('invalid')).rejects.toThrow('Suggestion not found')
    })
  })

  describe('dismiss', () => {
    it('should mark suggestion as dismissed', async () => {
      (apiClient.post as any).mockResolvedValueOnce({})

      await suggestionsApi.dismiss('sugg-123')

      expect(apiClient.post).toHaveBeenCalledWith('/suggestions/sugg-123/dismiss', {})
    })

    it('should handle already dismissed error', async () => {
      const error = new Error('Suggestion already dismissed')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(suggestionsApi.dismiss('sugg-123')).rejects.toThrow(
        'Suggestion already dismissed'
      )
    })

    it('should handle not found error', async () => {
      const error = new Error('Suggestion not found')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(suggestionsApi.dismiss('invalid')).rejects.toThrow('Suggestion not found')
    })
  })
})
