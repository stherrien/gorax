import { apiClient } from './client'
import type {
  MarketplaceTemplate,
  MarketplaceTemplateWithCategories,
  Category,
  PublishTemplateInput,
  InstallTemplateInput,
  InstallTemplateResult,
  RateTemplateInput,
  TemplateReview,
  SearchFilter,
  EnhancedSearchFilter,
  ReviewSortOption,
  RatingDistribution,
  ReportReviewInput,
  CreateCategoryInput,
  UpdateCategoryInput,
} from '../types/marketplace'

class MarketplaceAPI {
  /**
   * List marketplace templates with optional filters
   */
  async list(filter?: SearchFilter): Promise<MarketplaceTemplate[]> {
    const options = filter ? { params: filter } : undefined
    const response = await apiClient.get('/api/v1/marketplace/templates', options)
    return response.data || response
  }

  /**
   * Get a single marketplace template by ID
   */
  async get(id: string): Promise<MarketplaceTemplate> {
    const response = await apiClient.get(`/api/v1/marketplace/templates/${id}`)
    return response.data || response
  }

  /**
   * Publish a template to the marketplace
   */
  async publish(input: PublishTemplateInput): Promise<MarketplaceTemplate> {
    const response = await apiClient.post('/api/v1/marketplace/templates', input)
    return response.data || response
  }

  /**
   * Install a template as a workflow
   */
  async install(templateId: string, input: InstallTemplateInput): Promise<InstallTemplateResult> {
    const response = await apiClient.post(`/api/v1/marketplace/templates/${templateId}/install`, input)
    return response.data || response
  }

  /**
   * Rate a template
   */
  async rate(templateId: string, input: RateTemplateInput): Promise<TemplateReview> {
    const response = await apiClient.post(`/api/v1/marketplace/templates/${templateId}/rate`, input)
    return response.data || response
  }

  /**
   * Get reviews for a template with sorting
   */
  async getReviews(
    templateId: string,
    sortBy: ReviewSortOption = 'recent',
    limit?: number,
    offset?: number
  ): Promise<TemplateReview[]> {
    const params: Record<string, string | number> = { sort_by: sortBy }
    if (limit !== undefined) params.limit = limit
    if (offset !== undefined) params.offset = offset

    const response = await apiClient.get(`/api/v1/marketplace/templates/${templateId}/reviews`, { params })
    return response.data || response
  }

  /**
   * Delete a review
   */
  async deleteReview(templateId: string, reviewId: string): Promise<void> {
    await apiClient.delete(`/api/v1/marketplace/templates/${templateId}/reviews/${reviewId}`)
  }

  /**
   * Vote a review as helpful
   */
  async voteReviewHelpful(reviewId: string): Promise<void> {
    await apiClient.post(`/api/v1/marketplace/reviews/${reviewId}/helpful`)
  }

  /**
   * Remove helpful vote from a review
   */
  async unvoteReviewHelpful(reviewId: string): Promise<void> {
    await apiClient.delete(`/api/v1/marketplace/reviews/${reviewId}/helpful`)
  }

  /**
   * Check if user has voted a review as helpful
   */
  async hasVotedHelpful(reviewId: string): Promise<boolean> {
    const response = await apiClient.get(`/api/v1/marketplace/reviews/${reviewId}/helpful`)
    return response.data?.hasVoted || false
  }

  /**
   * Report a review
   */
  async reportReview(reviewId: string, input: ReportReviewInput): Promise<void> {
    await apiClient.post(`/api/v1/marketplace/reviews/${reviewId}/report`, input)
  }

  /**
   * Get rating distribution for a template
   */
  async getRatingDistribution(templateId: string): Promise<RatingDistribution> {
    const response = await apiClient.get(`/api/v1/marketplace/templates/${templateId}/rating-distribution`)
    return response.data || response
  }

  /**
   * Get trending templates
   */
  async getTrending(limit?: number): Promise<MarketplaceTemplate[]> {
    const options = limit ? { params: { limit } } : undefined
    const response = await apiClient.get('/api/v1/marketplace/trending', options)
    return response.data || response
  }

  /**
   * Get popular templates
   */
  async getPopular(limit?: number): Promise<MarketplaceTemplate[]> {
    const options = limit ? { params: { limit } } : undefined
    const response = await apiClient.get('/api/v1/marketplace/popular', options)
    return response.data || response
  }

  /**
   * Get all template categories (legacy - returns strings)
   */
  async getCategories(): Promise<string[]> {
    const response = await apiClient.get('/api/v1/marketplace/categories')
    return response.data || response
  }

  /**
   * List all categories with details
   */
  async listCategories(): Promise<Category[]> {
    const response = await apiClient.get('/api/v1/marketplace/categories/list')
    return response.data || response
  }

  /**
   * Get a single category by ID
   */
  async getCategory(categoryId: string): Promise<Category> {
    const response = await apiClient.get(`/api/v1/marketplace/categories/${categoryId}`)
    return response.data || response
  }

  /**
   * Create a new category
   */
  async createCategory(input: CreateCategoryInput): Promise<Category> {
    const response = await apiClient.post('/api/v1/marketplace/categories', input)
    return response.data || response
  }

  /**
   * Update a category
   */
  async updateCategory(categoryId: string, input: UpdateCategoryInput): Promise<Category> {
    const response = await apiClient.put(`/api/v1/marketplace/categories/${categoryId}`, input)
    return response.data || response
  }

  /**
   * Delete a category
   */
  async deleteCategory(categoryId: string): Promise<void> {
    await apiClient.delete(`/api/v1/marketplace/categories/${categoryId}`)
  }

  /**
   * Get featured templates
   */
  async getFeatured(limit?: number): Promise<MarketplaceTemplate[]> {
    const options = limit ? { params: { limit } } : undefined
    const response = await apiClient.get('/api/v1/marketplace/featured', options)
    return response.data || response
  }

  /**
   * Search templates with enhanced filter (supports category IDs)
   */
  async searchEnhanced(filter?: EnhancedSearchFilter): Promise<MarketplaceTemplate[]> {
    const options = filter ? { params: filter } : undefined
    const response = await apiClient.get('/api/v1/marketplace/templates/search', options)
    return response.data || response
  }

  /**
   * Get templates by category ID
   */
  async getByCategory(categoryId: string, limit?: number, offset?: number): Promise<MarketplaceTemplate[]> {
    const params: Record<string, number> = {}
    if (limit !== undefined) params.limit = limit
    if (offset !== undefined) params.offset = offset

    const options = Object.keys(params).length > 0 ? { params } : undefined
    const response = await apiClient.get(`/api/v1/marketplace/categories/${categoryId}/templates`, options)
    return response.data || response
  }
}

export const marketplaceAPI = new MarketplaceAPI()
