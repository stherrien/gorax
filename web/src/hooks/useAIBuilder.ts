import { useState, useCallback } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { aibuilderApi } from '../api/aibuilder'
import type {
  BuildRequest,
  BuildResult,
  RefineRequest,
  ConversationMessage,
} from '../types/aibuilder'

// Query keys
export const aibuilderKeys = {
  all: ['aibuilder'] as const,
  conversations: () => [...aibuilderKeys.all, 'conversations'] as const,
  conversation: (id: string) =>
    [...aibuilderKeys.all, 'conversation', id] as const,
}

/**
 * Hook to list all conversations
 */
export function useConversations() {
  return useQuery({
    queryKey: aibuilderKeys.conversations(),
    queryFn: () => aibuilderApi.listConversations(),
  })
}

/**
 * Hook to get a single conversation
 */
export function useConversation(conversationId: string | undefined) {
  return useQuery({
    queryKey: conversationId
      ? aibuilderKeys.conversation(conversationId)
      : ['disabled'],
    queryFn: () => aibuilderApi.getConversation(conversationId!),
    enabled: !!conversationId,
  })
}

/**
 * Hook to generate a new workflow
 */
export function useGenerateWorkflow() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: BuildRequest) => aibuilderApi.generate(request),
    onSuccess: () => {
      // Invalidate conversations list
      queryClient.invalidateQueries({
        queryKey: aibuilderKeys.conversations(),
      })
    },
  })
}

/**
 * Hook to refine an existing workflow
 */
export function useRefineWorkflow() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: RefineRequest) => aibuilderApi.refine(request),
    onSuccess: (_, variables) => {
      // Invalidate the specific conversation
      queryClient.invalidateQueries({
        queryKey: aibuilderKeys.conversation(variables.conversation_id),
      })
    },
  })
}

/**
 * Hook to apply a generated workflow
 */
export function useApplyWorkflow() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      conversationId,
      workflowName,
    }: {
      conversationId: string
      workflowName?: string
    }) => aibuilderApi.apply(conversationId, workflowName),
    onSuccess: () => {
      // Invalidate conversations list
      queryClient.invalidateQueries({
        queryKey: aibuilderKeys.conversations(),
      })
    },
  })
}

/**
 * Hook to abandon a conversation
 */
export function useAbandonConversation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (conversationId: string) => aibuilderApi.abandon(conversationId),
    onSuccess: () => {
      // Invalidate conversations list
      queryClient.invalidateQueries({
        queryKey: aibuilderKeys.conversations(),
      })
    },
  })
}

/**
 * Combined hook for AI builder actions
 */
export function useAIBuilderActions() {
  const generateMutation = useGenerateWorkflow()
  const refineMutation = useRefineWorkflow()
  const applyMutation = useApplyWorkflow()
  const abandonMutation = useAbandonConversation()

  const generate = useCallback(
    async (request: BuildRequest) => {
      return await generateMutation.mutateAsync(request)
    },
    [generateMutation]
  )

  const refine = useCallback(
    async (request: RefineRequest) => {
      return await refineMutation.mutateAsync(request)
    },
    [refineMutation]
  )

  const apply = useCallback(
    async (conversationId: string, workflowName?: string) => {
      return await applyMutation.mutateAsync({ conversationId, workflowName })
    },
    [applyMutation]
  )

  const abandon = useCallback(
    async (conversationId: string) => {
      await abandonMutation.mutateAsync(conversationId)
    },
    [abandonMutation]
  )

  return {
    generate,
    refine,
    apply,
    abandon,
    isGenerating: generateMutation.isPending,
    isRefining: refineMutation.isPending,
    isApplying: applyMutation.isPending,
    isAbandoning: abandonMutation.isPending,
    isLoading:
      generateMutation.isPending ||
      refineMutation.isPending ||
      applyMutation.isPending ||
      abandonMutation.isPending,
    generateError: generateMutation.error,
    refineError: refineMutation.error,
    applyError: applyMutation.error,
    abandonError: abandonMutation.error,
  }
}

/**
 * Hook to manage AI builder chat state
 */
export function useAIBuilderChat() {
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [messages, setMessages] = useState<ConversationMessage[]>([])
  const [currentWorkflow, setCurrentWorkflow] = useState<
    BuildResult['workflow'] | null
  >(null)

  const { generate, refine, apply, abandon, isGenerating, isRefining } =
    useAIBuilderActions()

  // Start a new conversation with a description
  const startConversation = useCallback(
    async (description: string) => {
      const result = await generate({ description })
      setConversationId(result.conversation_id)
      setCurrentWorkflow(result.workflow)

      // Add messages to local state
      const userMessage: ConversationMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content: description,
        created_at: new Date().toISOString(),
      }
      const assistantMessage: ConversationMessage = {
        id: `assistant-${Date.now()}`,
        role: 'assistant',
        content: result.explanation,
        workflow: result.workflow,
        created_at: new Date().toISOString(),
      }
      setMessages([userMessage, assistantMessage])

      return result
    },
    [generate]
  )

  // Send a message to refine the workflow
  const sendMessage = useCallback(
    async (message: string) => {
      if (!conversationId) {
        throw new Error('No active conversation')
      }

      // Add user message optimistically
      const userMessage: ConversationMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content: message,
        created_at: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, userMessage])

      const result = await refine({
        conversation_id: conversationId,
        message,
      })

      setCurrentWorkflow(result.workflow)

      // Add assistant response
      const assistantMessage: ConversationMessage = {
        id: `assistant-${Date.now()}`,
        role: 'assistant',
        content: result.explanation,
        workflow: result.workflow,
        created_at: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, assistantMessage])

      return result
    },
    [conversationId, refine]
  )

  // Apply the current workflow
  const applyWorkflow = useCallback(
    async (workflowName?: string) => {
      if (!conversationId) {
        throw new Error('No active conversation')
      }
      return await apply(conversationId, workflowName)
    },
    [conversationId, apply]
  )

  // Abandon the current conversation
  const abandonConversation = useCallback(async () => {
    if (!conversationId) return
    await abandon(conversationId)
    reset()
  }, [conversationId, abandon])

  // Reset the chat state
  const reset = useCallback(() => {
    setConversationId(null)
    setMessages([])
    setCurrentWorkflow(null)
  }, [])

  return {
    conversationId,
    messages,
    currentWorkflow,
    isGenerating,
    isRefining,
    isLoading: isGenerating || isRefining,
    startConversation,
    sendMessage,
    applyWorkflow,
    abandonConversation,
    reset,
    hasConversation: !!conversationId,
    hasWorkflow: !!currentWorkflow,
  }
}

/**
 * Hook to load an existing conversation
 */
export function useLoadConversation(conversationId: string | undefined) {
  const { data: conversation, isLoading, error } = useConversation(conversationId)
  const { refine, apply, abandon, isRefining, isApplying, isAbandoning } =
    useAIBuilderActions()

  const sendMessage = useCallback(
    async (message: string) => {
      if (!conversationId) {
        throw new Error('No conversation ID')
      }
      return await refine({
        conversation_id: conversationId,
        message,
      })
    },
    [conversationId, refine]
  )

  const applyWorkflow = useCallback(
    async (workflowName?: string) => {
      if (!conversationId) {
        throw new Error('No conversation ID')
      }
      return await apply(conversationId, workflowName)
    },
    [conversationId, apply]
  )

  const abandonConversation = useCallback(async () => {
    if (!conversationId) return
    await abandon(conversationId)
  }, [conversationId, abandon])

  return {
    conversation,
    isLoading,
    error,
    sendMessage,
    applyWorkflow,
    abandonConversation,
    isRefining,
    isApplying,
    isAbandoning,
    isActive: conversation?.status === 'active',
    currentWorkflow: conversation?.current_workflow,
    messages: conversation?.messages ?? [],
  }
}
