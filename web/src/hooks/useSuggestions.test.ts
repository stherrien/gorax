import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { suggestionsApi } from '../api/suggestions'
import {
  useSuggestions,
  useSuggestion,
  useApplySuggestion,
  useDismissSuggestion,
  useSuggestionActions,
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
