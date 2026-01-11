import { useState, useEffect, useCallback } from 'react'
import { marketplaceAPI } from '../api/marketplace'
import type {
  MarketplaceTemplate,
  Category,
  TemplateReview,
  SearchFilter,
  InstallTemplateInput,
  InstallTemplateResult,
  RateTemplateInput,
  ReviewSortOption,
  RatingDistribution,
  ReportReviewInput,
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
 * Hook to fetch and manage template reviews with sorting
 */
export function useMarketplaceReviews(
  templateId: string,
  sortBy: ReviewSortOption = 'recent',
  limit = 10,
  offset = 0
) {
  const [reviews, setReviews] = useState<TemplateReview[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchReviews = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.getReviews(templateId, sortBy, limit, offset)
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
  }, [templateId, sortBy, limit, offset])

  const deleteReview = async (reviewId: string) => {
    await marketplaceAPI.deleteReview(templateId, reviewId)
    setReviews((prev) => prev.filter((r) => r.id !== reviewId))
  }

  const refresh = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await marketplaceAPI.getReviews(templateId, sortBy, limit, offset)
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
 * Hook to fetch template categories (legacy - returns strings)
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

/**
 * Hook to fetch detailed categories with hierarchy
 */
export function useCategories() {
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchCategories = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await marketplaceAPI.listCategories()
      setCategories(data)
    } catch (err) {
      setError(err as Error)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchCategories()
  }, [fetchCategories])

  return { categories, loading, error, refetch: fetchCategories }
}

/**
 * Hook to fetch featured templates
 */
export function useFeaturedTemplates(limit = 10) {
  const [templates, setTemplates] = useState<MarketplaceTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchFeatured = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.getFeatured(limit)
        setTemplates(data)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    fetchFeatured()
  }, [limit])

  return { templates, loading, error }
}

/**
 * Hook to fetch rating distribution for a template
 */
export function useRatingDistribution(templateId: string) {
  const [distribution, setDistribution] = useState<RatingDistribution | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const fetchDistribution = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await marketplaceAPI.getRatingDistribution(templateId)
        setDistribution(data)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    if (templateId) {
      fetchDistribution()
    }
  }, [templateId])

  return { distribution, loading, error }
}

/**
 * Hook to manage helpful votes on reviews
 */
export function useReviewHelpful(reviewId: string) {
  const [hasVoted, setHasVoted] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const checkVoteStatus = async () => {
      try {
        const voted = await marketplaceAPI.hasVotedHelpful(reviewId)
        setHasVoted(voted)
      } catch (err) {
        setError(err as Error)
      }
    }

    if (reviewId) {
      checkVoteStatus()
    }
  }, [reviewId])

  const vote = async () => {
    try {
      setLoading(true)
      setError(null)
      await marketplaceAPI.voteReviewHelpful(reviewId)
      setHasVoted(true)
    } catch (err) {
      setError(err as Error)
      throw err
    } finally {
      setLoading(false)
    }
  }

  const unvote = async () => {
    try {
      setLoading(true)
      setError(null)
      await marketplaceAPI.unvoteReviewHelpful(reviewId)
      setHasVoted(false)
    } catch (err) {
      setError(err as Error)
      throw err
    } finally {
      setLoading(false)
    }
  }

  const toggleVote = async () => {
    if (hasVoted) {
      await unvote()
    } else {
      await vote()
    }
  }

  return { hasVoted, loading, error, vote, unvote, toggleVote }
}

/**
 * Hook to report a review
 */
export function useReportReview() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const report = async (reviewId: string, input: ReportReviewInput) => {
    try {
      setLoading(true)
      setError(null)
      await marketplaceAPI.reportReview(reviewId, input)
    } catch (err) {
      setError(err as Error)
      throw err
    } finally {
      setLoading(false)
    }
  }

  return { report, loading, error }
}

/**
 * Hook to create or update a review
 */
export function useCreateReview(templateId: string) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const createReview = async (input: RateTemplateInput): Promise<TemplateReview> => {
    try {
      setLoading(true)
      setError(null)
      const review = await marketplaceAPI.rate(templateId, input)
      return review
    } catch (err) {
      setError(err as Error)
      throw err
    } finally {
      setLoading(false)
    }
  }

  return { createReview, loading, error }
}
