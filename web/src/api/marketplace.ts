import { apiClient } from './client'
import type {
  MarketplaceTemplate,
  PublishTemplateInput,
  InstallTemplateInput,
  InstallTemplateResult,
  RateTemplateInput,
  TemplateReview,
  SearchFilter,
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
   * Get reviews for a template
   */
  async getReviews(templateId: string, limit?: number, offset?: number): Promise<TemplateReview[]> {
    const params: Record<string, number> = {}
    if (limit !== undefined) params.limit = limit
    if (offset !== undefined) params.offset = offset

    const options = Object.keys(params).length > 0 ? { params } : undefined
    const response = await apiClient.get(`/api/v1/marketplace/templates/${templateId}/reviews`, options)
    return response.data || response
  }

  /**
   * Delete a review
   */
  async deleteReview(templateId: string, reviewId: string): Promise<void> {
    await apiClient.delete(`/api/v1/marketplace/templates/${templateId}/reviews/${reviewId}`)
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
   * Get all template categories
   */
  async getCategories(): Promise<string[]> {
    const response = await apiClient.get('/api/v1/marketplace/categories')
    return response.data || response
  }
}

export const marketplaceAPI = new MarketplaceAPI()
