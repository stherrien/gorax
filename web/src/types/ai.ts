// AI Provider Types
export type AIProvider = 'openai' | 'anthropic' | 'bedrock';

// AI Action Types
export type AIAction =
  | 'chat_completion'
  | 'summarization'
  | 'classification'
  | 'entity_extraction'
  | 'embedding';

// Token Usage
export interface TokenUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}

// Chat Message
export interface ChatMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
  name?: string;
}

// AI Model
export interface AIModel {
  id: string;
  name: string;
  provider: AIProvider;
  maxTokens: number;
  contextWindow: number;
  inputCostPer1M: number;
  outputCostPer1M: number;
  capabilities: string[];
}

// Chat Completion Types
export interface ChatCompletionConfig {
  provider: AIProvider;
  model: string;
  systemPrompt?: string;
  messages: ChatMessage[];
  maxTokens?: number;
  temperature?: number;
  topP?: number;
  stop?: string[];
  presencePenalty?: number;
  frequencyPenalty?: number;
  user?: string;
}

export interface ChatCompletionResult {
  id: string;
  model: string;
  role: string;
  content: string;
  finishReason: string;
  usage: TokenUsage;
}

// Summarization Types
export interface SummarizationConfig {
  provider: AIProvider;
  model: string;
  text: string;
  maxLength?: number;
  format?: 'paragraph' | 'bullets';
  focus?: string;
  maxTokens?: number;
  temperature?: number;
}

export interface SummarizationResult {
  summary: string;
  wordCount: number;
  usage: TokenUsage;
}

// Classification Types
export interface ClassificationConfig {
  provider: AIProvider;
  model: string;
  text: string;
  categories: string[];
  multiLabel?: boolean;
  description?: string;
  maxTokens?: number;
  temperature?: number;
}

export interface ClassificationResult {
  category: string;
  categories?: string[];
  confidence: number;
  reasoning: string;
  usage: TokenUsage;
}

// Entity Extraction Types
export interface EntityExtractionConfig {
  provider: AIProvider;
  model: string;
  text: string;
  entityTypes: string[];
  customEntities?: Record<string, string>;
  maxTokens?: number;
  temperature?: number;
}

export interface ExtractedEntity {
  type: string;
  value: string;
  normalizedValue?: string;
  confidence?: number;
  context?: string;
}

export interface EntityExtractionResult {
  entities: ExtractedEntity[];
  entityCount: number;
  usage: TokenUsage;
}

// Embedding Types
export interface EmbeddingConfig {
  provider: AIProvider;
  model: string;
  texts: string[];
  user?: string;
}

export interface EmbeddingResult {
  embeddings: number[][];
  dimensions: number;
  count: number;
  usage: TokenUsage;
}

// AI Usage Types
export interface AIUsageLog {
  id: string;
  tenantId: string;
  credentialId?: string;
  provider: AIProvider;
  model: string;
  actionType: AIAction;
  executionId?: string;
  workflowId?: string;
  usage: TokenUsage;
  estimatedCostCents: number;
  success: boolean;
  errorCode?: string;
  errorMessage?: string;
  latencyMs: number;
  createdAt: string;
}

export interface ModelUsage {
  provider: AIProvider;
  model: string;
  requestCount: number;
  totalPromptTokens: number;
  totalCompletionTokens: number;
  totalTokens: number;
  totalCostCents: number;
}

export interface AIUsageSummary {
  totalRequests: number;
  totalPromptTokens: number;
  totalCompletionTokens: number;
  totalTokens: number;
  totalCostCents: number;
  byModel: ModelUsage[];
  from: string;
  to: string;
}

// AI Node Config (for workflow builder)
export interface AINodeConfig {
  action: AIAction;
  credentialId: string;
  config:
    | ChatCompletionConfig
    | SummarizationConfig
    | ClassificationConfig
    | EntityExtractionConfig
    | EmbeddingConfig;
}

// Available models by provider
export const AI_MODELS: AIModel[] = [
  // OpenAI Models
  {
    id: 'gpt-4o',
    name: 'GPT-4o',
    provider: 'openai',
    maxTokens: 4096,
    contextWindow: 128000,
    inputCostPer1M: 500,
    outputCostPer1M: 1500,
    capabilities: ['chat', 'function_calling', 'json_mode', 'vision'],
  },
  {
    id: 'gpt-4o-mini',
    name: 'GPT-4o Mini',
    provider: 'openai',
    maxTokens: 16384,
    contextWindow: 128000,
    inputCostPer1M: 15,
    outputCostPer1M: 60,
    capabilities: ['chat', 'function_calling', 'json_mode', 'vision'],
  },
  {
    id: 'gpt-4-turbo',
    name: 'GPT-4 Turbo',
    provider: 'openai',
    maxTokens: 4096,
    contextWindow: 128000,
    inputCostPer1M: 1000,
    outputCostPer1M: 3000,
    capabilities: ['chat', 'function_calling', 'json_mode', 'vision'],
  },
  {
    id: 'gpt-3.5-turbo',
    name: 'GPT-3.5 Turbo',
    provider: 'openai',
    maxTokens: 4096,
    contextWindow: 16385,
    inputCostPer1M: 50,
    outputCostPer1M: 150,
    capabilities: ['chat', 'function_calling', 'json_mode'],
  },
  {
    id: 'text-embedding-3-small',
    name: 'Text Embedding 3 Small',
    provider: 'openai',
    maxTokens: 0,
    contextWindow: 8191,
    inputCostPer1M: 2,
    outputCostPer1M: 0,
    capabilities: ['embedding'],
  },
  {
    id: 'text-embedding-3-large',
    name: 'Text Embedding 3 Large',
    provider: 'openai',
    maxTokens: 0,
    contextWindow: 8191,
    inputCostPer1M: 13,
    outputCostPer1M: 0,
    capabilities: ['embedding'],
  },
  // Anthropic Models
  {
    id: 'claude-3-opus-20240229',
    name: 'Claude 3 Opus',
    provider: 'anthropic',
    maxTokens: 4096,
    contextWindow: 200000,
    inputCostPer1M: 1500,
    outputCostPer1M: 7500,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  {
    id: 'claude-3-sonnet-20240229',
    name: 'Claude 3 Sonnet',
    provider: 'anthropic',
    maxTokens: 4096,
    contextWindow: 200000,
    inputCostPer1M: 300,
    outputCostPer1M: 1500,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  {
    id: 'claude-3-haiku-20240307',
    name: 'Claude 3 Haiku',
    provider: 'anthropic',
    maxTokens: 4096,
    contextWindow: 200000,
    inputCostPer1M: 25,
    outputCostPer1M: 125,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  {
    id: 'claude-3-5-sonnet-20241022',
    name: 'Claude 3.5 Sonnet',
    provider: 'anthropic',
    maxTokens: 8192,
    contextWindow: 200000,
    inputCostPer1M: 300,
    outputCostPer1M: 1500,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  // AWS Bedrock Models (Claude)
  {
    id: 'anthropic.claude-3-opus-20240229-v1:0',
    name: 'Claude 3 Opus (Bedrock)',
    provider: 'bedrock',
    maxTokens: 4096,
    contextWindow: 200000,
    inputCostPer1M: 1500,
    outputCostPer1M: 7500,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  {
    id: 'anthropic.claude-3-sonnet-20240229-v1:0',
    name: 'Claude 3 Sonnet (Bedrock)',
    provider: 'bedrock',
    maxTokens: 4096,
    contextWindow: 200000,
    inputCostPer1M: 300,
    outputCostPer1M: 1500,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  {
    id: 'anthropic.claude-3-haiku-20240307-v1:0',
    name: 'Claude 3 Haiku (Bedrock)',
    provider: 'bedrock',
    maxTokens: 4096,
    contextWindow: 200000,
    inputCostPer1M: 25,
    outputCostPer1M: 125,
    capabilities: ['chat', 'function_calling', 'vision'],
  },
  {
    id: 'amazon.titan-text-express-v1',
    name: 'Titan Text Express',
    provider: 'bedrock',
    maxTokens: 8000,
    contextWindow: 8000,
    inputCostPer1M: 80,
    outputCostPer1M: 240,
    capabilities: ['chat'],
  },
  {
    id: 'amazon.titan-embed-text-v2:0',
    name: 'Titan Embeddings V2',
    provider: 'bedrock',
    maxTokens: 0,
    contextWindow: 8192,
    inputCostPer1M: 2,
    outputCostPer1M: 0,
    capabilities: ['embedding'],
  },
];

// Helper functions
export function getModelsByProvider(provider: AIProvider): AIModel[] {
  return AI_MODELS.filter((model) => model.provider === provider);
}

export function getModelsWithCapability(capability: string): AIModel[] {
  return AI_MODELS.filter((model) => model.capabilities.includes(capability));
}

export function getChatModels(): AIModel[] {
  return getModelsWithCapability('chat');
}

export function getEmbeddingModels(): AIModel[] {
  return getModelsWithCapability('embedding');
}

export function estimateCost(
  model: AIModel,
  promptTokens: number,
  completionTokens: number
): number {
  const inputCost = (promptTokens / 1000000) * model.inputCostPer1M;
  const outputCost = (completionTokens / 1000000) * model.outputCostPer1M;
  return Math.round((inputCost + outputCost) * 100) / 100; // Round to cents
}

export function formatCost(cents: number): string {
  if (cents < 1) {
    return '<$0.01';
  }
  return `$${(cents / 100).toFixed(2)}`;
}

export function formatTokens(tokens: number): string {
  if (tokens >= 1000000) {
    return `${(tokens / 1000000).toFixed(1)}M`;
  }
  if (tokens >= 1000) {
    return `${(tokens / 1000).toFixed(1)}K`;
  }
  return tokens.toString();
}
