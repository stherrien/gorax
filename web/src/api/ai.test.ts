import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { aiAPI } from './ai'

// Mock the apiClient
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('AI API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('listModels', () => {
    it('should list all models when no provider specified', async () => {
      const mockModels = [
        { id: 'gpt-4o', name: 'GPT-4o', provider: 'openai' },
        { id: 'claude-3-opus', name: 'Claude 3 Opus', provider: 'anthropic' },
      ]
      vi.mocked(apiClient.get).mockResolvedValue({ models: mockModels })

      const result = await aiAPI.listModels()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/ai/models', undefined)
      expect(result.models).toEqual(mockModels)
    })

    it('should filter models by provider', async () => {
      const mockModels = [{ id: 'gpt-4o', name: 'GPT-4o', provider: 'openai' }]
      vi.mocked(apiClient.get).mockResolvedValue({ models: mockModels })

      const result = await aiAPI.listModels('openai')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/ai/models', {
        params: { provider: 'openai' },
      })
      expect(result.models).toEqual(mockModels)
    })
  })

  describe('getModelPricing', () => {
    it('should return pricing for a model', async () => {
      const mockPricing = {
        provider: 'openai',
        model: 'gpt-4o',
        inputCostPer1M: 500,
        outputCostPer1M: 1500,
        contextWindow: 128000,
      }
      vi.mocked(apiClient.get).mockResolvedValue(mockPricing)

      const result = await aiAPI.getModelPricing('openai', 'gpt-4o')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/ai/models/openai/gpt-4o/pricing')
      expect(result).toEqual(mockPricing)
    })
  })

  describe('estimateTokens', () => {
    it('should estimate token count for text', async () => {
      const mockResponse = { tokens: 150, model: 'gpt-4o' }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await aiAPI.estimateTokens('Hello world', 'gpt-4o')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/tokens/estimate', {
        text: 'Hello world',
        model: 'gpt-4o',
      })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('chatCompletion', () => {
    it('should execute chat completion', async () => {
      const mockResponse = {
        id: 'chatcmpl-123',
        model: 'gpt-4o',
        role: 'assistant',
        content: 'Hello! How can I help you?',
        finishReason: 'stop',
        usage: {
          promptTokens: 10,
          completionTokens: 8,
          totalTokens: 18,
        },
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'openai' as const,
        model: 'gpt-4o',
        credentialId: 'cred-123',
        messages: [{ role: 'user' as const, content: 'Hello' }],
      }

      const result = await aiAPI.chatCompletion(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/chat', config)
      expect(result).toEqual(mockResponse)
    })

    it('should include optional parameters', async () => {
      const mockResponse = { id: 'chatcmpl-123', content: 'Response' }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'openai' as const,
        model: 'gpt-4o',
        credentialId: 'cred-123',
        messages: [{ role: 'user' as const, content: 'Hello' }],
        systemPrompt: 'You are a helpful assistant',
        maxTokens: 1000,
        temperature: 0.7,
      }

      await aiAPI.chatCompletion(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/chat', config)
    })
  })

  describe('summarize', () => {
    it('should execute summarization', async () => {
      const mockResponse = {
        summary: 'This is a summary.',
        wordCount: 5,
        usage: { promptTokens: 100, completionTokens: 20, totalTokens: 120 },
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'openai' as const,
        model: 'gpt-4o',
        credentialId: 'cred-123',
        text: 'Long text to summarize...',
      }

      const result = await aiAPI.summarize(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/summarize', config)
      expect(result).toEqual(mockResponse)
    })

    it('should include format and maxLength options', async () => {
      const mockResponse = { summary: '- Point 1\n- Point 2' }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'anthropic' as const,
        model: 'claude-3-sonnet',
        credentialId: 'cred-123',
        text: 'Text to summarize',
        format: 'bullets' as const,
        maxLength: 100,
      }

      await aiAPI.summarize(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/summarize', config)
    })
  })

  describe('classify', () => {
    it('should execute classification', async () => {
      const mockResponse = {
        category: 'positive',
        confidence: 0.95,
        reasoning: 'The text expresses positive sentiment.',
        usage: { promptTokens: 50, completionTokens: 30, totalTokens: 80 },
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'openai' as const,
        model: 'gpt-4o',
        credentialId: 'cred-123',
        text: 'I love this product!',
        categories: ['positive', 'negative', 'neutral'],
      }

      const result = await aiAPI.classify(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/classify', config)
      expect(result).toEqual(mockResponse)
    })

    it('should support multi-label classification', async () => {
      const mockResponse = {
        category: 'urgent',
        categories: ['urgent', 'customer-support'],
        confidence: 0.9,
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'openai' as const,
        model: 'gpt-4o',
        credentialId: 'cred-123',
        text: 'Need help ASAP!',
        categories: ['urgent', 'customer-support', 'billing', 'technical'],
        multiLabel: true,
      }

      await aiAPI.classify(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/classify', config)
    })
  })

  describe('extractEntities', () => {
    it('should execute entity extraction', async () => {
      const mockResponse = {
        entities: [
          { type: 'person', value: 'John Smith', confidence: 0.95 },
          { type: 'organization', value: 'Acme Corp', confidence: 0.9 },
        ],
        entityCount: 2,
        usage: { promptTokens: 60, completionTokens: 40, totalTokens: 100 },
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'openai' as const,
        model: 'gpt-4o',
        credentialId: 'cred-123',
        text: 'John Smith works at Acme Corp.',
        entityTypes: ['person', 'organization'],
      }

      const result = await aiAPI.extractEntities(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/extract-entities', config)
      expect(result).toEqual(mockResponse)
    })

    it('should support custom entity definitions', async () => {
      const mockResponse = {
        entities: [{ type: 'order_number', value: '#12345' }],
        entityCount: 1,
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'openai' as const,
        model: 'gpt-4o',
        credentialId: 'cred-123',
        text: 'Order #12345 has shipped.',
        entityTypes: [],
        customEntities: {
          order_number: 'Order ID starting with #',
        },
      }

      await aiAPI.extractEntities(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/extract-entities', config)
    })
  })

  describe('generateEmbeddings', () => {
    it('should generate embeddings for texts', async () => {
      const mockResponse = {
        embeddings: [
          [0.1, 0.2, 0.3],
          [0.4, 0.5, 0.6],
        ],
        dimensions: 3,
        count: 2,
        usage: { promptTokens: 10, completionTokens: 0, totalTokens: 10 },
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const config = {
        provider: 'openai' as const,
        model: 'text-embedding-3-small',
        credentialId: 'cred-123',
        texts: ['Hello world', 'Goodbye world'],
      }

      const result = await aiAPI.generateEmbeddings(config)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/embeddings', config)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('getUsage', () => {
    it('should get usage summary for date range', async () => {
      const mockResponse = {
        totalRequests: 100,
        totalPromptTokens: 10000,
        totalCompletionTokens: 5000,
        totalTokens: 15000,
        totalCostCents: 250,
        byModel: [
          {
            provider: 'openai',
            model: 'gpt-4o',
            requestCount: 50,
            totalTokens: 10000,
            totalCostCents: 200,
          },
        ],
        from: '2024-01-01T00:00:00Z',
        to: '2024-01-31T23:59:59Z',
      }
      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const from = new Date('2024-01-01')
      const to = new Date('2024-01-31')

      const result = await aiAPI.getUsage(from, to)

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/ai/usage', {
        params: {
          from: '2024-01-01T00:00:00.000Z',
          to: '2024-01-31T00:00:00.000Z',
        },
      })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('getUsageLogs', () => {
    it('should get paginated usage logs', async () => {
      const mockResponse = {
        logs: [
          {
            id: 'log-1',
            provider: 'openai',
            model: 'gpt-4o',
            actionType: 'chat_completion',
            usage: { promptTokens: 100, completionTokens: 50, totalTokens: 150 },
            success: true,
            createdAt: '2024-01-15T10:00:00Z',
          },
        ],
        total: 100,
        page: 1,
        limit: 20,
      }
      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const result = await aiAPI.getUsageLogs({ page: 1, limit: 20 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/ai/usage/logs', {
        params: { page: 1, limit: 20 },
      })
      expect(result).toEqual(mockResponse)
    })

    it('should filter logs by provider and model', async () => {
      const mockResponse = { logs: [], total: 0 }
      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      await aiAPI.getUsageLogs({
        provider: 'openai',
        model: 'gpt-4o',
        actionType: 'chat_completion',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/ai/usage/logs', {
        params: {
          provider: 'openai',
          model: 'gpt-4o',
          actionType: 'chat_completion',
        },
      })
    })
  })

  describe('testCredential', () => {
    it('should test AI credential connectivity', async () => {
      const mockResponse = {
        success: true,
        provider: 'openai',
        message: 'Successfully connected to OpenAI API',
        models: ['gpt-4o', 'gpt-4o-mini'],
        testedAt: '2024-01-15T10:00:00Z',
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await aiAPI.testCredential('cred-123', 'openai')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/credentials/test', {
        credentialId: 'cred-123',
        provider: 'openai',
      })
      expect(result).toEqual(mockResponse)
    })

    it('should return error message on failure', async () => {
      const mockResponse = {
        success: false,
        provider: 'openai',
        message: 'Invalid API key',
        testedAt: '2024-01-15T10:00:00Z',
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await aiAPI.testCredential('cred-123', 'openai')

      expect(result.success).toBe(false)
      expect(result.message).toBe('Invalid API key')
    })
  })

  describe('estimateCost', () => {
    it('should estimate cost for given tokens', async () => {
      const mockResponse = {
        provider: 'openai',
        model: 'gpt-4o',
        promptTokens: 1000,
        completionTokens: 500,
        totalTokens: 1500,
        estimatedCostCents: 125,
      }
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await aiAPI.estimateCost('openai', 'gpt-4o', 1000, 500)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/ai/cost/estimate', {
        provider: 'openai',
        model: 'gpt-4o',
        promptTokens: 1000,
        completionTokens: 500,
      })
      expect(result).toEqual(mockResponse)
    })
  })
})
