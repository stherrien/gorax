import { apiClient } from './client'
import type {
  Suggestion,
  AnalyzeRequest,
  SuggestionsListResponse,
  SuggestionResponse,
  AnalyzeResponse,
} from '../types/suggestions'

/**
 * Suggestions API client
 * Provides methods for managing smart suggestions for workflow errors
 */
export const suggestionsApi = {
  /**
   * Get all suggestions for an execution
   * @param executionId - The execution ID
   * @returns List of suggestions
   */
  async list(executionId: string): Promise<Suggestion[]> {
    const response: SuggestionsListResponse = await apiClient.get(
      `/executions/${executionId}/suggestions`
    )
    return response.data
  },

  /**
   * Get a single suggestion by ID
   * @param suggestionId - The suggestion ID
   * @returns The suggestion
   */
  async get(suggestionId: string): Promise<Suggestion> {
    const response: SuggestionResponse = await apiClient.get(
      `/suggestions/${suggestionId}`
    )
    return response.data
  },

  /**
   * Analyze an execution error and generate suggestions
   * @param executionId - The execution ID
   * @param request - The error context for analysis
   * @returns Generated suggestions
   */
  async analyze(
    executionId: string,
    request: AnalyzeRequest
  ): Promise<Suggestion[]> {
    const response: AnalyzeResponse = await apiClient.post(
      `/executions/${executionId}/analyze`,
      request
    )
    return response.data
  },

  /**
   * Mark a suggestion as applied
   * @param suggestionId - The suggestion ID
   */
  async apply(suggestionId: string): Promise<void> {
    await apiClient.post(
      `/suggestions/${suggestionId}/apply`,
      {}
    )
  },

  /**
   * Mark a suggestion as dismissed
   * @param suggestionId - The suggestion ID
   */
  async dismiss(suggestionId: string): Promise<void> {
    await apiClient.post(
      `/suggestions/${suggestionId}/dismiss`,
      {}
    )
  },
}

export default suggestionsApi
