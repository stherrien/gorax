import { describe, it, expect, beforeEach, vi } from 'vitest'
import { marketplaceAPI } from './marketplace'
import * as client from './client'

import { createQueryWrapper } from "../test/test-utils"
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('MarketplaceAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should list all templates', async () => {
      const mockTemplates = [
        { id: '1', name: 'Template 1' },
        { id: '2', name: 'Template 2' },
      ]

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockTemplates)

      const result = await marketplaceAPI.list()

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/templates', undefined)
      expect(result).toEqual(mockTemplates)
    })

    it('should list templates with filters', async () => {
      const mockTemplates = [{ id: '1', name: 'Template 1' }]
      const filter = { category: 'automation', limit: 10 }

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockTemplates)

      const result = await marketplaceAPI.list(filter)

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/templates', { params: filter })
      expect(result).toEqual(mockTemplates)
    })
  })

  describe('get', () => {
    it('should get a single template', async () => {
      const mockTemplate = { id: '1', name: 'Template 1' }

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockTemplate)

      const result = await marketplaceAPI.get('1')

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/templates/1')
      expect(result).toEqual(mockTemplate)
    })
  })

  describe('publish', () => {
    it('should publish a template', async () => {
      const input = {
        name: 'Test Template',
        description: 'Test description that is long enough',
        category: 'automation',
        definition: { nodes: [], edges: [] },
        version: '1.0.0',
      }
      const mockTemplate = { id: '1', ...input }

      vi.spyOn(client.apiClient, 'post').mockResolvedValue(mockTemplate)

      const result = await marketplaceAPI.publish(input)

      expect(client.apiClient.post).toHaveBeenCalledWith('/api/v1/marketplace/templates', input)
      expect(result).toEqual(mockTemplate)
    })
  })

  describe('install', () => {
    it('should install a template', async () => {
      const input = { workflowName: 'My Workflow' }
      const mockResult = {
        workflowId: 'workflow-1',
        workflowName: 'My Workflow',
        definition: { nodes: [], edges: [] },
      }

      vi.spyOn(client.apiClient, 'post').mockResolvedValue(mockResult)

      const result = await marketplaceAPI.install('template-1', input)

      expect(client.apiClient.post).toHaveBeenCalledWith('/api/v1/marketplace/templates/template-1/install', input)
      expect(result).toEqual(mockResult)
    })
  })

  describe('rate', () => {
    it('should rate a template', async () => {
      const input = { rating: 5, comment: 'Great!' }
      const mockReview = { id: 'review-1', rating: 5, comment: 'Great!' }

      vi.spyOn(client.apiClient, 'post').mockResolvedValue(mockReview)

      const result = await marketplaceAPI.rate('template-1', input)

      expect(client.apiClient.post).toHaveBeenCalledWith('/api/v1/marketplace/templates/template-1/rate', input)
      expect(result).toEqual(mockReview)
    })
  })

  describe('getReviews', () => {
    it('should get reviews for a template', async () => {
      const mockReviews = [
        { id: '1', rating: 5 },
        { id: '2', rating: 4 },
      ]

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockReviews)

      const result = await marketplaceAPI.getReviews('template-1')

      // Default sortBy is 'recent', so params always includes sort_by
      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/templates/template-1/reviews', {
        params: { sort_by: 'recent' },
      })
      expect(result).toEqual(mockReviews)
    })

    it('should get reviews with pagination', async () => {
      const mockReviews = [{ id: '1', rating: 5 }]

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockReviews)

      // Function signature: getReviews(templateId, sortBy, limit, offset)
      const result = await marketplaceAPI.getReviews('template-1', 'recent', 10, 20)

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/templates/template-1/reviews', {
        params: { sort_by: 'recent', limit: 10, offset: 20 },
      })
      expect(result).toEqual(mockReviews)
    })
  })

  describe('deleteReview', () => {
    it('should delete a review', async () => {
      vi.spyOn(client.apiClient, 'delete').mockResolvedValue(undefined)

      await marketplaceAPI.deleteReview('template-1', 'review-1')

      expect(client.apiClient.delete).toHaveBeenCalledWith('/api/v1/marketplace/templates/template-1/reviews/review-1')
    })
  })

  describe('getTrending', () => {
    it('should get trending templates', async () => {
      const mockTemplates = [{ id: '1', name: 'Trending 1' }]

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockTemplates)

      const result = await marketplaceAPI.getTrending()

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/trending', undefined)
      expect(result).toEqual(mockTemplates)
    })

    it('should get trending templates with limit', async () => {
      const mockTemplates = [{ id: '1', name: 'Trending 1' }]

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockTemplates)

      const result = await marketplaceAPI.getTrending(5)

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/trending', { params: { limit: 5 } })
      expect(result).toEqual(mockTemplates)
    })
  })

  describe('getPopular', () => {
    it('should get popular templates', async () => {
      const mockTemplates = [{ id: '1', name: 'Popular 1' }]

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockTemplates)

      const result = await marketplaceAPI.getPopular()

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/popular', undefined)
      expect(result).toEqual(mockTemplates)
    })

    it('should get popular templates with limit', async () => {
      const mockTemplates = [{ id: '1', name: 'Popular 1' }]

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockTemplates)

      const result = await marketplaceAPI.getPopular(5)

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/popular', { params: { limit: 5 } })
      expect(result).toEqual(mockTemplates)
    })
  })

  describe('getCategories', () => {
    it('should get all categories', async () => {
      const mockCategories = ['security', 'automation', 'monitoring']

      vi.spyOn(client.apiClient, 'get').mockResolvedValue(mockCategories)

      const result = await marketplaceAPI.getCategories()

      expect(client.apiClient.get).toHaveBeenCalledWith('/api/v1/marketplace/categories')
      expect(result).toEqual(mockCategories)
    })
  })
})
