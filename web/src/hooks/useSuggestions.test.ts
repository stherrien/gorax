import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { suggestionsApi } from '../api/suggestions'
import {
  useSuggestions,
  useSuggestion,
  useAnalyzeError,
  useApplySuggestion,
  useDismissSuggestion,
  useSuggestionActions,
  useSuggestionsPanel,
  groupSuggestionsByCategory,
  sortSuggestionsByConfidence,
} from './useSuggestions'
import type { Suggestion } from '../types/suggestions'
import { createElement } from 'react'

// Mock the API
vi.mock('../api/suggestions', () => ({
  suggestionsApi: {
    list: vi.fn(),
    get: vi.fn(),
    analyze: vi.fn(),
    apply: vi.fn(),
    dismiss: vi.fn(),
  },
}))

// Helper to create test suggestions
const createTestSuggestion = (
  overrides: Partial<Suggestion> = {}
): Suggestion => ({
  id: 'sugg-123',
  tenant_id: 'tenant-456',
  execution_id: 'exec-789',
  node_id: 'node-abc',
  category: 'network',
  type: 'retry',
  confidence: 'high',
  title: 'Connection Error',
  description: 'Connection refused',
  source: 'pattern',
  status: 'pending',
  created_at: '2025-12-20T00:00:00Z',
  ...overrides,
})

// Wrapper for hooks that need QueryClient
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })
  return ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('useSuggestions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch suggestions for an execution', async () => {
    const mockSuggestions = [createTestSuggestion()]
    vi.mocked(suggestionsApi.list).mockResolvedValue(mockSuggestions)

    const { result } = renderHook(() => useSuggestions('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(suggestionsApi.list).toHaveBeenCalledWith('exec-123')
    expect(result.current.data).toEqual(mockSuggestions)
  })

  it('should not fetch when executionId is undefined', () => {
    const { result } = renderHook(() => useSuggestions(undefined), {
      wrapper: createWrapper(),
    })

    expect(suggestionsApi.list).not.toHaveBeenCalled()
    expect(result.current.isFetching).toBe(false)
  })
})

describe('useSuggestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch a single suggestion', async () => {
    const mockSuggestion = createTestSuggestion()
    vi.mocked(suggestionsApi.get).mockResolvedValue(mockSuggestion)

    const { result } = renderHook(() => useSuggestion('sugg-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(suggestionsApi.get).toHaveBeenCalledWith('sugg-123')
    expect(result.current.data).toEqual(mockSuggestion)
  })

  it('should not fetch when suggestionId is undefined', () => {
    const { result } = renderHook(() => useSuggestion(undefined), {
      wrapper: createWrapper(),
    })

    expect(suggestionsApi.get).not.toHaveBeenCalled()
    expect(result.current.isFetching).toBe(false)
  })
})

describe('useApplySuggestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should apply a suggestion', async () => {
    vi.mocked(suggestionsApi.apply).mockResolvedValue(undefined)

    const { result } = renderHook(() => useApplySuggestion(), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      await result.current.mutateAsync('sugg-123')
    })

    expect(suggestionsApi.apply).toHaveBeenCalledWith('sugg-123')
  })
})

describe('useDismissSuggestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should dismiss a suggestion', async () => {
    vi.mocked(suggestionsApi.dismiss).mockResolvedValue(undefined)

    const { result } = renderHook(() => useDismissSuggestion(), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      await result.current.mutateAsync('sugg-123')
    })

    expect(suggestionsApi.dismiss).toHaveBeenCalledWith('sugg-123')
  })
})

describe('useSuggestionActions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should provide apply and dismiss functions', async () => {
    vi.mocked(suggestionsApi.apply).mockResolvedValue(undefined)
    vi.mocked(suggestionsApi.dismiss).mockResolvedValue(undefined)

    const { result } = renderHook(() => useSuggestionActions(), {
      wrapper: createWrapper(),
    })

    // Test apply
    await act(async () => {
      await result.current.apply('sugg-123')
    })
    expect(suggestionsApi.apply).toHaveBeenCalledWith('sugg-123')

    // Test dismiss
    await act(async () => {
      await result.current.dismiss('sugg-456')
    })
    expect(suggestionsApi.dismiss).toHaveBeenCalledWith('sugg-456')
  })

  it('should track loading state during apply', async () => {
    vi.mocked(suggestionsApi.apply).mockResolvedValue(undefined)

    const { result } = renderHook(() => useSuggestionActions(), {
      wrapper: createWrapper(),
    })

    // Initial state should not be loading
    expect(result.current.isLoading).toBe(false)

    // After mutation completes, should not be loading
    await act(async () => {
      await result.current.apply('sugg-123')
    })

    expect(result.current.isLoading).toBe(false)
    expect(suggestionsApi.apply).toHaveBeenCalledWith('sugg-123')
  })
})

describe('groupSuggestionsByCategory', () => {
  it('should group suggestions by category', () => {
    const suggestions = [
      createTestSuggestion({ id: '1', category: 'network' }),
      createTestSuggestion({ id: '2', category: 'auth' }),
      createTestSuggestion({ id: '3', category: 'network' }),
    ]

    const grouped = groupSuggestionsByCategory(suggestions)

    expect(Object.keys(grouped)).toHaveLength(2)
    expect(grouped['network']).toHaveLength(2)
    expect(grouped['auth']).toHaveLength(1)
  })

  it('should handle empty array', () => {
    const grouped = groupSuggestionsByCategory([])
    expect(Object.keys(grouped)).toHaveLength(0)
  })
})

describe('sortSuggestionsByConfidence', () => {
  it('should sort by confidence (high first)', () => {
    const suggestions = [
      createTestSuggestion({ id: '1', confidence: 'low' }),
      createTestSuggestion({ id: '2', confidence: 'high' }),
      createTestSuggestion({ id: '3', confidence: 'medium' }),
    ]

    const sorted = sortSuggestionsByConfidence(suggestions)

    expect(sorted[0].confidence).toBe('high')
    expect(sorted[1].confidence).toBe('medium')
    expect(sorted[2].confidence).toBe('low')
  })

  it('should not mutate original array', () => {
    const suggestions = [
      createTestSuggestion({ id: '1', confidence: 'low' }),
      createTestSuggestion({ id: '2', confidence: 'high' }),
    ]

    const sorted = sortSuggestionsByConfidence(suggestions)

    expect(suggestions[0].confidence).toBe('low')
    expect(sorted).not.toBe(suggestions)
  })
})

describe('useAnalyzeError', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should analyze an execution error', async () => {
    const mockResult = { suggestionId: 'sugg-new' }
    vi.mocked(suggestionsApi.analyze).mockResolvedValue(mockResult)

    const { result } = renderHook(() => useAnalyzeError('exec-123'), {
      wrapper: createWrapper(),
    })

    await act(async () => {
      await result.current.mutateAsync({
        errorMessage: 'Connection refused',
        errorType: 'network',
        nodeId: 'node-1',
      })
    })

    expect(suggestionsApi.analyze).toHaveBeenCalledWith('exec-123', {
      errorMessage: 'Connection refused',
      errorType: 'network',
      nodeId: 'node-1',
    })
  })

  it('should handle analysis errors', async () => {
    const mockError = new Error('Analysis failed')
    vi.mocked(suggestionsApi.analyze).mockRejectedValue(mockError)

    const { result } = renderHook(() => useAnalyzeError('exec-123'), {
      wrapper: createWrapper(),
    })

    await expect(
      act(async () => {
        await result.current.mutateAsync({
          errorMessage: 'Error',
          errorType: 'unknown',
          nodeId: 'node-1',
        })
      })
    ).rejects.toThrow('Analysis failed')
  })
})

describe('useSuggestions error handling', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should handle fetch errors', async () => {
    const mockError = new Error('Failed to fetch')
    vi.mocked(suggestionsApi.list).mockRejectedValue(mockError)

    const { result } = renderHook(() => useSuggestions('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))

    expect(result.current.error).toEqual(mockError)
    expect(result.current.data).toBeUndefined()
  })
})

describe('useSuggestion error handling', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should handle fetch errors', async () => {
    const mockError = new Error('Suggestion not found')
    vi.mocked(suggestionsApi.get).mockRejectedValue(mockError)

    const { result } = renderHook(() => useSuggestion('sugg-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))

    expect(result.current.error).toEqual(mockError)
    expect(result.current.data).toBeUndefined()
  })
})

// Note: Error handling for mutations is implicitly tested via query error handling tests
// The mutation hooks simply propagate errors from the API calls

describe('useSuggestionsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch suggestions for an execution', async () => {
    const mockSuggestions = [
      createTestSuggestion({ id: '1', status: 'pending' }),
      createTestSuggestion({ id: '2', status: 'applied' }),
      createTestSuggestion({ id: '3', status: 'dismissed' }),
    ]
    vi.mocked(suggestionsApi.list).mockResolvedValue(mockSuggestions)

    const { result } = renderHook(() => useSuggestionsPanel('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.suggestions).toEqual(mockSuggestions)
    expect(result.current.pendingSuggestions).toHaveLength(1)
    expect(result.current.appliedSuggestions).toHaveLength(1)
    expect(result.current.dismissedSuggestions).toHaveLength(1)
  })

  it('should return empty arrays when no executionId', async () => {
    const { result } = renderHook(() => useSuggestionsPanel(undefined), {
      wrapper: createWrapper(),
    })

    expect(result.current.suggestions).toEqual([])
    expect(result.current.pendingSuggestions).toEqual([])
    expect(result.current.appliedSuggestions).toEqual([])
    expect(result.current.dismissedSuggestions).toEqual([])
    expect(suggestionsApi.list).not.toHaveBeenCalled()
  })

  it('should handle selection state', async () => {
    const mockSuggestions = [createTestSuggestion({ id: 'sugg-1' })]
    vi.mocked(suggestionsApi.list).mockResolvedValue(mockSuggestions)

    const { result } = renderHook(() => useSuggestionsPanel('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    // Initial state - no selection
    expect(result.current.selectedSuggestionId).toBeNull()
    expect(result.current.selectedSuggestion).toBeUndefined()

    // Select a suggestion
    act(() => {
      result.current.setSelectedSuggestionId('sugg-1')
    })

    expect(result.current.selectedSuggestionId).toBe('sugg-1')
    expect(result.current.selectedSuggestion).toEqual(mockSuggestions[0])
  })

  it('should apply suggestion and clear selection', async () => {
    const mockSuggestions = [createTestSuggestion({ id: 'sugg-1' })]
    vi.mocked(suggestionsApi.list).mockResolvedValue(mockSuggestions)
    vi.mocked(suggestionsApi.apply).mockResolvedValue(undefined)

    const { result } = renderHook(() => useSuggestionsPanel('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    // Select and apply
    act(() => {
      result.current.setSelectedSuggestionId('sugg-1')
    })

    await act(async () => {
      await result.current.apply('sugg-1')
    })

    expect(suggestionsApi.apply).toHaveBeenCalledWith('sugg-1')
    expect(result.current.selectedSuggestionId).toBeNull()
  })

  it('should dismiss suggestion and clear selection', async () => {
    const mockSuggestions = [createTestSuggestion({ id: 'sugg-1' })]
    vi.mocked(suggestionsApi.list).mockResolvedValue(mockSuggestions)
    vi.mocked(suggestionsApi.dismiss).mockResolvedValue(undefined)

    const { result } = renderHook(() => useSuggestionsPanel('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    // Select and dismiss
    act(() => {
      result.current.setSelectedSuggestionId('sugg-1')
    })

    await act(async () => {
      await result.current.dismiss('sugg-1')
    })

    expect(suggestionsApi.dismiss).toHaveBeenCalledWith('sugg-1')
    expect(result.current.selectedSuggestionId).toBeNull()
  })

  it('should not clear selection when applying different suggestion', async () => {
    const mockSuggestions = [
      createTestSuggestion({ id: 'sugg-1' }),
      createTestSuggestion({ id: 'sugg-2' }),
    ]
    vi.mocked(suggestionsApi.list).mockResolvedValue(mockSuggestions)
    vi.mocked(suggestionsApi.apply).mockResolvedValue(undefined)

    const { result } = renderHook(() => useSuggestionsPanel('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    // Select sugg-1, apply sugg-2
    act(() => {
      result.current.setSelectedSuggestionId('sugg-1')
    })

    await act(async () => {
      await result.current.apply('sugg-2')
    })

    // Selection should remain on sugg-1
    expect(result.current.selectedSuggestionId).toBe('sugg-1')
  })

  it('should handle multiple pending suggestions', async () => {
    const mockSuggestions = [
      createTestSuggestion({ id: '1', status: 'pending', confidence: 'high' }),
      createTestSuggestion({ id: '2', status: 'pending', confidence: 'medium' }),
      createTestSuggestion({ id: '3', status: 'pending', confidence: 'low' }),
      createTestSuggestion({ id: '4', status: 'applied' }),
    ]
    vi.mocked(suggestionsApi.list).mockResolvedValue(mockSuggestions)

    const { result } = renderHook(() => useSuggestionsPanel('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.pendingSuggestions).toHaveLength(3)
    expect(result.current.appliedSuggestions).toHaveLength(1)
  })

  it('should provide refetch function', async () => {
    const mockSuggestions1 = [createTestSuggestion({ id: '1' })]
    const mockSuggestions2 = [
      createTestSuggestion({ id: '1' }),
      createTestSuggestion({ id: '2' }),
    ]

    vi.mocked(suggestionsApi.list)
      .mockResolvedValueOnce(mockSuggestions1)
      .mockResolvedValueOnce(mockSuggestions2)

    const { result } = renderHook(() => useSuggestionsPanel('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(result.current.suggestions).toHaveLength(1)

    await act(async () => {
      await result.current.refetch()
    })

    await waitFor(() => expect(result.current.suggestions).toHaveLength(2))
  })

  it('should track loading states for apply and dismiss', async () => {
    const mockSuggestions = [createTestSuggestion({ id: 'sugg-1' })]
    vi.mocked(suggestionsApi.list).mockResolvedValue(mockSuggestions)
    vi.mocked(suggestionsApi.apply).mockResolvedValue(undefined)
    vi.mocked(suggestionsApi.dismiss).mockResolvedValue(undefined)

    const { result } = renderHook(() => useSuggestionsPanel('exec-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    // Initial states should be false
    expect(result.current.isApplying).toBe(false)
    expect(result.current.isDismissing).toBe(false)

    // After apply completes
    await act(async () => {
      await result.current.apply('sugg-1')
    })
    expect(result.current.isApplying).toBe(false)

    // After dismiss completes
    await act(async () => {
      await result.current.dismiss('sugg-1')
    })
    expect(result.current.isDismissing).toBe(false)
  })
})
