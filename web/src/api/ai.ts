import { apiClient } from './client'
import type {
  AIProvider,
  AIAction,
  AIModel,
  ChatMessage,
  TokenUsage,
  ChatCompletionResult,
  SummarizationResult,
  ClassificationResult,
  EntityExtractionResult,
  EmbeddingResult,
  AIUsageLog,
  AIUsageSummary,
} from '../types/ai'

// Request types for API calls
export interface ChatCompletionRequest {
  provider: AIProvider
  model: string
  credentialId: string
  messages: ChatMessage[]
  systemPrompt?: string
  maxTokens?: number
  temperature?: number
  topP?: number
  stop?: string[]
  presencePenalty?: number
  frequencyPenalty?: number
  user?: string
}

export interface SummarizationRequest {
  provider: AIProvider
  model: string
  credentialId: string
  text: string
  maxLength?: number
  format?: 'paragraph' | 'bullets'
  focus?: string
  maxTokens?: number
  temperature?: number
}

export interface ClassificationRequest {
  provider: AIProvider
  model: string
  credentialId: string
  text: string
  categories: string[]
  multiLabel?: boolean
  description?: string
  maxTokens?: number
  temperature?: number
}

export interface EntityExtractionRequest {
  provider: AIProvider
  model: string
  credentialId: string
  text: string
  entityTypes: string[]
  customEntities?: Record<string, string>
  maxTokens?: number
  temperature?: number
}

export interface EmbeddingRequest {
  provider: AIProvider
  model: string
  credentialId: string
  texts: string[]
  user?: string
}

// Response types
export interface ModelsListResponse {
  models: AIModel[]
}

export interface ModelPricing {
  provider: string
  model: string
  inputCostPer1M: number
  outputCostPer1M: number
  contextWindow: number
  maxOutputTokens?: number
  supportsVision?: boolean
  supportsFunctionCalling?: boolean
  supportsJsonMode?: boolean
}

export interface TokenEstimateResponse {
  tokens: number
  model: string
}

export interface UsageLogsResponse {
  logs: AIUsageLog[]
  total: number
  page?: number
  limit?: number
}

export interface UsageLogsParams {
  page?: number
  limit?: number
  provider?: AIProvider
  model?: string
  actionType?: AIAction
  from?: string
  to?: string
  success?: boolean
}

export interface CredentialTestResult {
  success: boolean
  provider: string
  message: string
  models?: string[]
  testedAt: string
}

export interface CostEstimateResponse {
  provider: string
  model: string
  promptTokens: number
  completionTokens: number
  totalTokens: number
  estimatedCostCents: number
}

class AIAPI {
  /**
   * List available AI models
   * @param provider - Optional filter by provider
   */
  async listModels(provider?: AIProvider): Promise<ModelsListResponse> {
    const options = provider ? { params: { provider } } : undefined
    return await apiClient.get('/api/v1/ai/models', options)
  }

  /**
   * Get pricing information for a specific model
   */
  async getModelPricing(provider: string, model: string): Promise<ModelPricing> {
    return await apiClient.get(`/api/v1/ai/models/${provider}/${model}/pricing`)
  }

  /**
   * Estimate token count for given text
   */
  async estimateTokens(text: string, model: string): Promise<TokenEstimateResponse> {
    return await apiClient.post('/api/v1/ai/tokens/estimate', { text, model })
  }

  /**
   * Execute a chat completion request
   */
  async chatCompletion(config: ChatCompletionRequest): Promise<ChatCompletionResult> {
    return await apiClient.post('/api/v1/ai/chat', config)
  }

  /**
   * Execute a summarization request
   */
  async summarize(config: SummarizationRequest): Promise<SummarizationResult> {
    return await apiClient.post('/api/v1/ai/summarize', config)
  }

  /**
   * Execute a classification request
   */
  async classify(config: ClassificationRequest): Promise<ClassificationResult> {
    return await apiClient.post('/api/v1/ai/classify', config)
  }

  /**
   * Execute an entity extraction request
   */
  async extractEntities(config: EntityExtractionRequest): Promise<EntityExtractionResult> {
    return await apiClient.post('/api/v1/ai/extract-entities', config)
  }

  /**
   * Generate embeddings for texts
   */
  async generateEmbeddings(config: EmbeddingRequest): Promise<EmbeddingResult> {
    return await apiClient.post('/api/v1/ai/embeddings', config)
  }

  /**
   * Get AI usage summary for a date range
   */
  async getUsage(from: Date, to: Date): Promise<AIUsageSummary> {
    return await apiClient.get('/api/v1/ai/usage', {
      params: {
        from: from.toISOString(),
        to: to.toISOString(),
      },
    })
  }

  /**
   * Get paginated AI usage logs
   */
  async getUsageLogs(params: UsageLogsParams): Promise<UsageLogsResponse> {
    return await apiClient.get('/api/v1/ai/usage/logs', { params })
  }

  /**
   * Test an AI credential's connectivity
   */
  async testCredential(credentialId: string, provider: AIProvider): Promise<CredentialTestResult> {
    return await apiClient.post('/api/v1/ai/credentials/test', {
      credentialId,
      provider,
    })
  }

  /**
   * Estimate cost for given token counts
   */
  async estimateCost(
    provider: string,
    model: string,
    promptTokens: number,
    completionTokens: number
  ): Promise<CostEstimateResponse> {
    return await apiClient.post('/api/v1/ai/cost/estimate', {
      provider,
      model,
      promptTokens,
      completionTokens,
    })
  }
}

export const aiAPI = new AIAPI()
