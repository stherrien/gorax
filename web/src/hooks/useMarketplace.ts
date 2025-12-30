import { useState, useEffect } from 'react'
import { marketplaceAPI } from '../api/marketplace'
import type {
  MarketplaceTemplate,
  TemplateReview,
  SearchFilter,
  InstallTemplateInput,
  InstallTemplateResult,
  RateTemplateInput,
} from '../types/marketplace'

/**
 * Hook to fetch and manage marketplace templates
 */
export function useMarketplace(filter?: SearchFilter) {
  const [templates, setTemplates] = useState<MarketplaceTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchTemplates = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.list(filter)
        setTemplates(data)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    fetchTemplates()
  }, [filter])

  const refresh = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await marketplaceAPI.list(filter)
      setTemplates(data)
    } catch (err) {
      setError(err as Error)
    } finally {
      setLoading(false)
    }
  }

  return { templates, loading, error, refresh }
}

/**
 * Hook to fetch and manage a single marketplace template
 */
export function useMarketplaceTemplate(templateId: string) {
  const [template, setTemplate] = useState<MarketplaceTemplate | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchTemplate = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.get(templateId)
        setTemplate(data)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    if (templateId) {
      fetchTemplate()
    }
  }, [templateId])

  const install = async (input: InstallTemplateInput): Promise<InstallTemplateResult> => {
    return await marketplaceAPI.install(templateId, input)
  }

  const rate = async (input: RateTemplateInput): Promise<TemplateReview> => {
    return await marketplaceAPI.rate(templateId, input)
  }

  const refresh = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await marketplaceAPI.get(templateId)
      setTemplate(data)
    } catch (err) {
      setError(err as Error)
    } finally {
      setLoading(false)
    }
  }

  return { template, loading, error, install, rate, refresh }
}

/**
 * Hook to fetch and manage template reviews
 */
export function useMarketplaceReviews(templateId: string, limit = 10) {
  const [reviews, setReviews] = useState<TemplateReview[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchReviews = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.getReviews(templateId, limit)
        setReviews(data)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    if (templateId) {
      fetchReviews()
    }
  }, [templateId, limit])

  const deleteReview = async (reviewId: string) => {
    try {
      await marketplaceAPI.deleteReview(templateId, reviewId)
      setReviews((prev) => prev.filter((r) => r.id !== reviewId))
    } catch (err) {
      throw err
    }
  }

  const refresh = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await marketplaceAPI.getReviews(templateId, limit)
      setReviews(data)
    } catch (err) {
      setError(err as Error)
    } finally {
      setLoading(false)
    }
  }

  return { reviews, loading, error, deleteReview, refresh }
}

/**
 * Hook to fetch trending templates
 */
export function useTrendingTemplates(limit = 10) {
  const [templates, setTemplates] = useState<MarketplaceTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchTrending = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.getTrending(limit)
        setTemplates(data)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    fetchTrending()
  }, [limit])

  return { templates, loading, error }
}

/**
 * Hook to fetch popular templates
 */
export function usePopularTemplates(limit = 10) {
  const [templates, setTemplates] = useState<MarketplaceTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchPopular = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.getPopular(limit)
        setTemplates(data)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    fetchPopular()
  }, [limit])

  return { templates, loading, error }
}

/**
 * Hook to fetch template categories
 */
export function useMarketplaceCategories() {
  const [categories, setCategories] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.getCategories()
        setCategories(data)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    fetchCategories()
  }, [])

  return { categories, loading, error }
}
