import { apiClient } from './client'
import type {
  BuildRequest,
  BuildResult,
  RefineRequest,
  Conversation,
  ConversationsListResponse,
  ApplyResponse,
} from '../types/aibuilder'

/**
 * AI Builder API client
 * Provides methods for AI-powered workflow generation
 */
export const aibuilderApi = {
  /**
   * Generate a workflow from a description
   * @param request - Build request with description and optional context/constraints
   * @returns Build result with generated workflow
   */
  async generate(request: BuildRequest): Promise<BuildResult> {
    const response: BuildResult = await apiClient.post(
      '/ai/workflows/generate',
      request
    )
    return response
  },

  /**
   * Refine an existing workflow based on feedback
   * @param request - Refine request with conversation ID and message
   * @returns Build result with refined workflow
   */
  async refine(request: RefineRequest): Promise<BuildResult> {
    const response: BuildResult = await apiClient.post(
      '/ai/workflows/refine',
      request
    )
    return response
  },

  /**
   * Get a conversation by ID
   * @param conversationId - The conversation ID
   * @returns The conversation with messages
   */
  async getConversation(conversationId: string): Promise<Conversation> {
    const response: Conversation = await apiClient.get(
      `/ai/workflows/conversations/${conversationId}`
    )
    return response
  },

  /**
   * List all conversations for the current user
   * @returns List of conversations
   */
  async listConversations(): Promise<Conversation[]> {
    const response: ConversationsListResponse = await apiClient.get(
      '/ai/workflows/conversations'
    )
    return response.data
  },

  /**
   * Apply a generated workflow (create actual workflow)
   * @param conversationId - The conversation ID
   * @param workflowName - Optional workflow name override
   * @returns The created workflow ID
   */
  async apply(
    conversationId: string,
    workflowName?: string
  ): Promise<string> {
    const response: ApplyResponse = await apiClient.post(
      `/ai/workflows/conversations/${conversationId}/apply`,
      workflowName ? { workflow_name: workflowName } : {}
    )
    return response.workflow_id
  },

  /**
   * Abandon a conversation
   * @param conversationId - The conversation ID
   */
  async abandon(conversationId: string): Promise<void> {
    await apiClient.post(
      `/ai/workflows/conversations/${conversationId}/abandon`,
      {}
    )
  },
}

export default aibuilderApi
