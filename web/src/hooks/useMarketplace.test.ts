import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import {
  useMarketplace,
  useMarketplaceTemplate,
  useMarketplaceReviews,
  useTrendingTemplates,
  usePopularTemplates,
  useMarketplaceCategories,
} from './useMarketplace'
import { marketplaceAPI } from '../api/marketplace'

vi.mock('../api/marketplace', () => ({
  marketplaceAPI: {
    list: vi.fn(),
    get: vi.fn(),
    publish: vi.fn(),
    install: vi.fn(),
    rate: vi.fn(),
    getReviews: vi.fn(),
    deleteReview: vi.fn(),
    getTrending: vi.fn(),
    getPopular: vi.fn(),
    getCategories: vi.fn(),
  },
}))

describe('useMarketplace', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch templates on mount', async () => {
    const mockTemplates = [
      { id: '1', name: 'Template 1' },
      { id: '2', name: 'Template 2' },
    ]

    vi.mocked(marketplaceAPI.list).mockResolvedValue(mockTemplates as any)

    const { result } = renderHook(() => useMarketplace())

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.templates).toEqual(mockTemplates)
    expect(result.current.error).toBeNull()
  })

  it('should handle errors when fetching templates', async () => {
    const mockError = new Error('Failed to fetch')
    vi.mocked(marketplaceAPI.list).mockRejectedValue(mockError)

    const { result } = renderHook(() => useMarketplace())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.templates).toEqual([])
    expect(result.current.error).toEqual(mockError)
  })

  it('should refetch templates when filter changes', async () => {
    const mockTemplates1 = [{ id: '1', name: 'Template 1' }]
    const mockTemplates2 = [{ id: '2', name: 'Template 2' }]

    vi.mocked(marketplaceAPI.list)
      .mockResolvedValueOnce(mockTemplates1 as any)
      .mockResolvedValueOnce(mockTemplates2 as any)

    const { result, rerender } = renderHook(
      ({ filter }) => useMarketplace(filter),
      { initialProps: { filter: { category: 'automation' } } }
    )

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.templates).toEqual(mockTemplates1)

    rerender({ filter: { category: 'security' } })

    await waitFor(() => {
      expect(result.current.templates).toEqual(mockTemplates2)
    })
  })
})

describe('useMarketplaceTemplate', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch a single template', async () => {
    const mockTemplate = { id: '1', name: 'Template 1' }

    vi.mocked(marketplaceAPI.get).mockResolvedValue(mockTemplate as any)

    const { result } = renderHook(() => useMarketplaceTemplate('1'))

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.template).toEqual(mockTemplate)
    expect(result.current.error).toBeNull()
  })

  it('should handle errors when fetching template', async () => {
    const mockError = new Error('Template not found')
    vi.mocked(marketplaceAPI.get).mockRejectedValue(mockError)

    const { result } = renderHook(() => useMarketplaceTemplate('1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.template).toBeNull()
    expect(result.current.error).toEqual(mockError)
  })

  it('should install template', async () => {
    const mockTemplate = { id: '1', name: 'Template 1' }
    const mockResult = { workflowId: 'workflow-1', workflowName: 'My Workflow' }

    vi.mocked(marketplaceAPI.get).mockResolvedValue(mockTemplate as any)
    vi.mocked(marketplaceAPI.install).mockResolvedValue(mockResult as any)

    const { result } = renderHook(() => useMarketplaceTemplate('1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    const installResult = await result.current.install({ workflowName: 'My Workflow' })

    expect(installResult).toEqual(mockResult)
    expect(marketplaceAPI.install).toHaveBeenCalledWith('1', { workflowName: 'My Workflow' })
  })

  it('should rate template', async () => {
    const mockTemplate = { id: '1', name: 'Template 1' }
    const mockReview = { id: 'review-1', rating: 5 }

    vi.mocked(marketplaceAPI.get).mockResolvedValue(mockTemplate as any)
    vi.mocked(marketplaceAPI.rate).mockResolvedValue(mockReview as any)

    const { result } = renderHook(() => useMarketplaceTemplate('1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    const review = await result.current.rate({ rating: 5, comment: 'Great!' })

    expect(review).toEqual(mockReview)
    expect(marketplaceAPI.rate).toHaveBeenCalledWith('1', { rating: 5, comment: 'Great!' })
  })
})

describe('useMarketplaceReviews', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch reviews for a template', async () => {
    const mockReviews = [
      { id: '1', rating: 5 },
      { id: '2', rating: 4 },
    ]

    vi.mocked(marketplaceAPI.getReviews).mockResolvedValue(mockReviews as any)

    const { result } = renderHook(() => useMarketplaceReviews('template-1'))

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.reviews).toEqual(mockReviews)
    expect(result.current.error).toBeNull()
  })

  it('should delete a review', async () => {
    const mockReviews = [
      { id: '1', rating: 5 },
      { id: '2', rating: 4 },
    ]

    vi.mocked(marketplaceAPI.getReviews).mockResolvedValue(mockReviews as any)
    vi.mocked(marketplaceAPI.deleteReview).mockResolvedValue(undefined)

    const { result } = renderHook(() => useMarketplaceReviews('template-1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    await result.current.deleteReview('review-1')

    expect(marketplaceAPI.deleteReview).toHaveBeenCalledWith('template-1', 'review-1')
  })

  it('should remove deleted review from state', async () => {
    const mockReviews = [
      { id: 'review-1', rating: 5 },
      { id: 'review-2', rating: 4 },
    ]

    vi.mocked(marketplaceAPI.getReviews).mockResolvedValue(mockReviews as any)
    vi.mocked(marketplaceAPI.deleteReview).mockResolvedValue(undefined)

    const { result } = renderHook(() => useMarketplaceReviews('template-1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.reviews).toHaveLength(2)

    await result.current.deleteReview('review-1')

    await waitFor(() => {
      expect(result.current.reviews).toHaveLength(1)
    })
    expect(result.current.reviews[0].id).toBe('review-2')
  })

  it('should handle delete review error', async () => {
    const mockReviews = [{ id: 'review-1', rating: 5 }]
    const mockError = new Error('Delete failed')

    vi.mocked(marketplaceAPI.getReviews).mockResolvedValue(mockReviews as any)
    vi.mocked(marketplaceAPI.deleteReview).mockRejectedValue(mockError)

    const { result } = renderHook(() => useMarketplaceReviews('template-1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    await expect(result.current.deleteReview('review-1')).rejects.toThrow('Delete failed')
  })

  it('should not fetch when templateId is empty', async () => {
    const { result } = renderHook(() => useMarketplaceReviews(''))

    // Give time for any async operations to run
    await new Promise((resolve) => setTimeout(resolve, 50))

    // API should not be called when templateId is empty
    expect(marketplaceAPI.getReviews).not.toHaveBeenCalled()
    // Note: loading stays true because the hook doesn't handle empty templateId by setting loading=false
    expect(result.current.loading).toBe(true)
    expect(result.current.reviews).toEqual([])
  })

  it('should handle errors when fetching reviews', async () => {
    const mockError = new Error('Failed to fetch reviews')
    vi.mocked(marketplaceAPI.getReviews).mockRejectedValue(mockError)

    const { result } = renderHook(() => useMarketplaceReviews('template-1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(mockError)
    expect(result.current.reviews).toEqual([])
  })

  it('should refresh reviews', async () => {
    const mockReviews1 = [{ id: '1', rating: 5 }]
    const mockReviews2 = [{ id: '1', rating: 5 }, { id: '2', rating: 4 }]

    vi.mocked(marketplaceAPI.getReviews)
      .mockResolvedValueOnce(mockReviews1 as any)
      .mockResolvedValueOnce(mockReviews2 as any)

    const { result } = renderHook(() => useMarketplaceReviews('template-1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.reviews).toEqual(mockReviews1)

    await result.current.refresh()

    await waitFor(() => {
      expect(result.current.reviews).toEqual(mockReviews2)
    })
  })
})

describe('useTrendingTemplates', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch trending templates', async () => {
    const mockTemplates = [
      { id: '1', name: 'Trending 1' },
      { id: '2', name: 'Trending 2' },
    ]

    vi.mocked(marketplaceAPI.getTrending).mockResolvedValue(mockTemplates as any)

    const { result } = renderHook(() => useTrendingTemplates())

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.templates).toEqual(mockTemplates)
    expect(result.current.error).toBeNull()
    expect(marketplaceAPI.getTrending).toHaveBeenCalledWith(10) // Default limit
  })

  it('should use custom limit', async () => {
    vi.mocked(marketplaceAPI.getTrending).mockResolvedValue([])

    renderHook(() => useTrendingTemplates(5))

    await waitFor(() => {
      expect(marketplaceAPI.getTrending).toHaveBeenCalledWith(5)
    })
  })

  it('should handle errors', async () => {
    const mockError = new Error('Failed to fetch trending')
    vi.mocked(marketplaceAPI.getTrending).mockRejectedValue(mockError)

    const { result } = renderHook(() => useTrendingTemplates())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(mockError)
    expect(result.current.templates).toEqual([])
  })
})

describe('usePopularTemplates', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch popular templates', async () => {
    const mockTemplates = [
      { id: '1', name: 'Popular 1' },
      { id: '2', name: 'Popular 2' },
    ]

    vi.mocked(marketplaceAPI.getPopular).mockResolvedValue(mockTemplates as any)

    const { result } = renderHook(() => usePopularTemplates())

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.templates).toEqual(mockTemplates)
    expect(result.current.error).toBeNull()
    expect(marketplaceAPI.getPopular).toHaveBeenCalledWith(10) // Default limit
  })

  it('should use custom limit', async () => {
    vi.mocked(marketplaceAPI.getPopular).mockResolvedValue([])

    renderHook(() => usePopularTemplates(20))

    await waitFor(() => {
      expect(marketplaceAPI.getPopular).toHaveBeenCalledWith(20)
    })
  })

  it('should handle errors', async () => {
    const mockError = new Error('Failed to fetch popular')
    vi.mocked(marketplaceAPI.getPopular).mockRejectedValue(mockError)

    const { result } = renderHook(() => usePopularTemplates())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(mockError)
    expect(result.current.templates).toEqual([])
  })
})

describe('useMarketplaceCategories', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch categories', async () => {
    const mockCategories = ['automation', 'security', 'devops']

    vi.mocked(marketplaceAPI.getCategories).mockResolvedValue(mockCategories)

    const { result } = renderHook(() => useMarketplaceCategories())

    expect(result.current.loading).toBe(true)

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.categories).toEqual(mockCategories)
    expect(result.current.error).toBeNull()
  })

  it('should handle errors', async () => {
    const mockError = new Error('Failed to fetch categories')
    vi.mocked(marketplaceAPI.getCategories).mockRejectedValue(mockError)

    const { result } = renderHook(() => useMarketplaceCategories())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toEqual(mockError)
    expect(result.current.categories).toEqual([])
  })
})

describe('useMarketplace refresh', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should refresh templates', async () => {
    const mockTemplates1 = [{ id: '1', name: 'Template 1' }]
    const mockTemplates2 = [{ id: '1', name: 'Template 1' }, { id: '2', name: 'Template 2' }]

    vi.mocked(marketplaceAPI.list)
      .mockResolvedValueOnce(mockTemplates1 as any)
      .mockResolvedValueOnce(mockTemplates2 as any)

    const { result } = renderHook(() => useMarketplace())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.templates).toEqual(mockTemplates1)

    await result.current.refresh()

    await waitFor(() => {
      expect(result.current.templates).toEqual(mockTemplates2)
    })
  })

  it('should handle refresh error', async () => {
    const mockTemplates = [{ id: '1', name: 'Template 1' }]
    const mockError = new Error('Refresh failed')

    vi.mocked(marketplaceAPI.list)
      .mockResolvedValueOnce(mockTemplates as any)
      .mockRejectedValueOnce(mockError)

    const { result } = renderHook(() => useMarketplace())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    await result.current.refresh()

    await waitFor(() => {
      expect(result.current.error).toEqual(mockError)
    })
  })
})

describe('useMarketplaceTemplate refresh', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should refresh template', async () => {
    const mockTemplate1 = { id: '1', name: 'Template 1', version: 1 }
    const mockTemplate2 = { id: '1', name: 'Template 1 Updated', version: 2 }

    vi.mocked(marketplaceAPI.get)
      .mockResolvedValueOnce(mockTemplate1 as any)
      .mockResolvedValueOnce(mockTemplate2 as any)

    const { result } = renderHook(() => useMarketplaceTemplate('1'))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.template).toEqual(mockTemplate1)

    await result.current.refresh()

    await waitFor(() => {
      expect(result.current.template).toEqual(mockTemplate2)
    })
  })

  it('should not fetch when templateId is empty', async () => {
    const { result } = renderHook(() => useMarketplaceTemplate(''))

    // Give time for any async operations to run
    await new Promise((resolve) => setTimeout(resolve, 50))

    // API should not be called when templateId is empty
    expect(marketplaceAPI.get).not.toHaveBeenCalled()
    // Note: loading stays true because the hook doesn't handle empty templateId by setting loading=false
    expect(result.current.loading).toBe(true)
    expect(result.current.template).toBeNull()
  })
})
