import { useState, useCallback } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { suggestionsApi } from '../api/suggestions'
import type { Suggestion, AnalyzeRequest } from '../types/suggestions'

// Query keys
export const suggestionKeys = {
  all: ['suggestions'] as const,
  list: (executionId: string) =>
    [...suggestionKeys.all, 'list', executionId] as const,
  detail: (suggestionId: string) =>
    [...suggestionKeys.all, 'detail', suggestionId] as const,
}

/**
 * Hook to fetch suggestions for an execution
 */
export function useSuggestions(executionId: string | undefined) {
  return useQuery({
    queryKey: executionId ? suggestionKeys.list(executionId) : ['disabled'],
    queryFn: () => suggestionsApi.list(executionId!),
    enabled: !!executionId,
  })
}

/**
 * Hook to fetch a single suggestion
 */
export function useSuggestion(suggestionId: string | undefined) {
  return useQuery({
    queryKey: suggestionId ? suggestionKeys.detail(suggestionId) : ['disabled'],
    queryFn: () => suggestionsApi.get(suggestionId!),
    enabled: !!suggestionId,
  })
}

/**
 * Hook to analyze an execution error
 */
export function useAnalyzeError(executionId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: AnalyzeRequest) =>
      suggestionsApi.analyze(executionId, request),
    onSuccess: () => {
      // Invalidate suggestions list to refresh
      queryClient.invalidateQueries({
        queryKey: suggestionKeys.list(executionId),
      })
    },
  })
}

/**
 * Hook to apply a suggestion
 */
export function useApplySuggestion() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (suggestionId: string) => suggestionsApi.apply(suggestionId),
    onSuccess: () => {
      // Invalidate all suggestion queries
      queryClient.invalidateQueries({ queryKey: suggestionKeys.all })
    },
  })
}

/**
 * Hook to dismiss a suggestion
 */
export function useDismissSuggestion() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (suggestionId: string) => suggestionsApi.dismiss(suggestionId),
    onSuccess: () => {
      // Invalidate all suggestion queries
      queryClient.invalidateQueries({ queryKey: suggestionKeys.all })
    },
  })
}

/**
 * Combined hook for suggestion actions
 */
export function useSuggestionActions() {
  const applyMutation = useApplySuggestion()
  const dismissMutation = useDismissSuggestion()

  const apply = useCallback(
    async (suggestionId: string) => {
      await applyMutation.mutateAsync(suggestionId)
    },
    [applyMutation]
  )

  const dismiss = useCallback(
    async (suggestionId: string) => {
      await dismissMutation.mutateAsync(suggestionId)
    },
    [dismissMutation]
  )

  return {
    apply,
    dismiss,
    isApplying: applyMutation.isPending,
    isDismissing: dismissMutation.isPending,
    isLoading: applyMutation.isPending || dismissMutation.isPending,
    applyError: applyMutation.error,
    dismissError: dismissMutation.error,
  }
}

/**
 * Hook to manage suggestions panel state
 */
export function useSuggestionsPanel(executionId: string | undefined) {
  const [selectedSuggestionId, setSelectedSuggestionId] = useState<
    string | null
  >(null)

  const {
    data: suggestions,
    isLoading,
    error,
    refetch,
  } = useSuggestions(executionId)

  const { apply, dismiss, isApplying, isDismissing } = useSuggestionActions()

  const selectedSuggestion = suggestions?.find(
    (s) => s.id === selectedSuggestionId
  )

  const pendingSuggestions =
    suggestions?.filter((s) => s.status === 'pending') ?? []
  const appliedSuggestions =
    suggestions?.filter((s) => s.status === 'applied') ?? []
  const dismissedSuggestions =
    suggestions?.filter((s) => s.status === 'dismissed') ?? []

  const handleApply = useCallback(
    async (suggestionId: string) => {
      await apply(suggestionId)
      if (selectedSuggestionId === suggestionId) {
        setSelectedSuggestionId(null)
      }
    },
    [apply, selectedSuggestionId]
  )

  const handleDismiss = useCallback(
    async (suggestionId: string) => {
      await dismiss(suggestionId)
      if (selectedSuggestionId === suggestionId) {
        setSelectedSuggestionId(null)
      }
    },
    [dismiss, selectedSuggestionId]
  )

  return {
    suggestions: suggestions ?? [],
    pendingSuggestions,
    appliedSuggestions,
    dismissedSuggestions,
    selectedSuggestion,
    selectedSuggestionId,
    setSelectedSuggestionId,
    isLoading,
    error,
    refetch,
    apply: handleApply,
    dismiss: handleDismiss,
    isApplying,
    isDismissing,
  }
}

/**
 * Utility to group suggestions by category
 */
export function groupSuggestionsByCategory(
  suggestions: Suggestion[]
): Record<string, Suggestion[]> {
  return suggestions.reduce(
    (acc, suggestion) => {
      const category = suggestion.category
      if (!acc[category]) {
        acc[category] = []
      }
      acc[category].push(suggestion)
      return acc
    },
    {} as Record<string, Suggestion[]>
  )
}

/**
 * Utility to sort suggestions by confidence
 */
export function sortSuggestionsByConfidence(
  suggestions: Suggestion[]
): Suggestion[] {
  const confidenceOrder = { high: 0, medium: 1, low: 2 }
  return [...suggestions].sort(
    (a, b) => confidenceOrder[a.confidence] - confidenceOrder[b.confidence]
  )
}
