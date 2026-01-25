import { describe, it, expect, beforeEach, vi } from 'vitest'
import { aibuilderApi } from './aibuilder'
import type {
  BuildRequest,
  BuildResult,
  RefineRequest,
  Conversation,
  ConversationsListResponse,
} from '../types/aibuilder'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('AI Builder API', () => {
  const mockBuildResult: BuildResult = {
    conversation_id: 'conv-123',
    workflow: {
      name: 'Data Sync Workflow',
      description: 'Syncs data from API to database',
      definition: {
        nodes: [
          {
            id: 'node-1',
            type: 'trigger:webhook',
            name: 'Webhook Trigger',
            config: { path: '/webhook' },
            position: { x: 0, y: 0 },
          },
          {
            id: 'node-2',
            type: 'action:http',
            name: 'HTTP Request',
            config: { url: 'https://api.example.com/data' },
            position: { x: 200, y: 0 },
          },
        ],
        edges: [
          {
            id: 'edge-1',
            source: 'node-1',
            target: 'node-2',
          },
        ],
      },
    },
    explanation: 'Created a workflow that triggers on webhook and fetches data from API',
    warnings: ['Consider adding error handling'],
    suggestions: ['You might want to add a retry mechanism'],
    prompt_tokens: 100,
    completion_tokens: 250,
  }

  const mockConversation: Conversation = {
    id: 'conv-123',
    tenant_id: 'tenant-1',
    user_id: 'user-1',
    status: 'active',
    current_workflow: mockBuildResult.workflow,
    messages: [
      {
        id: 'msg-1',
        role: 'user',
        content: 'Create a workflow that syncs data',
        created_at: '2024-01-15T10:00:00Z',
      },
      {
        id: 'msg-2',
        role: 'assistant',
        content: 'I created a workflow for you',
        workflow: mockBuildResult.workflow,
        created_at: '2024-01-15T10:00:01Z',
      },
    ],
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-15T10:00:01Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('generate', () => {
    it('should generate workflow from description', async () => {
      const request: BuildRequest = {
        description: 'Create a workflow that syncs data from an API to a database',
      }
      ;(apiClient.post as any).mockResolvedValueOnce(mockBuildResult)

      const result = await aibuilderApi.generate(request)

      expect(apiClient.post).toHaveBeenCalledWith('/ai/workflows/generate', request)
      expect(result).toEqual(mockBuildResult)
      expect(result.workflow).toBeDefined()
    })

    it('should generate workflow with context', async () => {
      const request: BuildRequest = {
        description: 'Create a workflow that uses Slack',
        context: {
          available_credentials: ['slack-api-key'],
          available_integrations: ['slack', 'email'],
        },
      }
      ;(apiClient.post as any).mockResolvedValueOnce(mockBuildResult)

      await aibuilderApi.generate(request)

      expect(apiClient.post).toHaveBeenCalledWith('/ai/workflows/generate', request)
    })

    it('should generate workflow with constraints', async () => {
      const request: BuildRequest = {
        description: 'Create a simple workflow',
        constraints: {
          max_nodes: 5,
          allowed_types: ['trigger:webhook', 'action:http'],
          require_trigger: true,
        },
      }
      ;(apiClient.post as any).mockResolvedValueOnce(mockBuildResult)

      await aibuilderApi.generate(request)

      expect(apiClient.post).toHaveBeenCalledWith('/ai/workflows/generate', request)
    })

    it('should handle generation error', async () => {
      const error = new Error('Failed to generate workflow')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(
        aibuilderApi.generate({ description: 'Invalid request' })
      ).rejects.toThrow('Failed to generate workflow')
    })

    it('should return warnings and suggestions', async () => {
      (apiClient.post as any).mockResolvedValueOnce(mockBuildResult)

      const result = await aibuilderApi.generate({ description: 'Create workflow' })

      expect(result.warnings).toContain('Consider adding error handling')
      expect(result.suggestions).toContain('You might want to add a retry mechanism')
    })
  })

  describe('refine', () => {
    it('should refine existing workflow', async () => {
      const request: RefineRequest = {
        conversation_id: 'conv-123',
        message: 'Add error handling to the workflow',
      }
      const refinedResult: BuildResult = {
        ...mockBuildResult,
        explanation: 'Added error handling to your workflow',
      }
      ;(apiClient.post as any).mockResolvedValueOnce(refinedResult)

      const result = await aibuilderApi.refine(request)

      expect(apiClient.post).toHaveBeenCalledWith('/ai/workflows/refine', request)
      expect(result.explanation).toBe('Added error handling to your workflow')
    })

    it('should handle refine error', async () => {
      const error = new Error('Conversation not found')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(
        aibuilderApi.refine({ conversation_id: 'invalid', message: 'test' })
      ).rejects.toThrow('Conversation not found')
    })
  })

  describe('getConversation', () => {
    it('should fetch conversation by ID', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockConversation)

      const result = await aibuilderApi.getConversation('conv-123')

      expect(apiClient.get).toHaveBeenCalledWith('/ai/workflows/conversations/conv-123')
      expect(result).toEqual(mockConversation)
    })

    it('should return conversation with messages', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockConversation)

      const result = await aibuilderApi.getConversation('conv-123')

      expect(result.messages).toHaveLength(2)
      expect(result.messages[0].role).toBe('user')
      expect(result.messages[1].role).toBe('assistant')
    })

    it('should handle not found error', async () => {
      const error = new Error('Conversation not found')
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(aibuilderApi.getConversation('invalid')).rejects.toThrow(
        'Conversation not found'
      )
    })
  })

  describe('listConversations', () => {
    it('should fetch all conversations', async () => {
      const response: ConversationsListResponse = {
        data: [mockConversation],
      }
      ;(apiClient.get as any).mockResolvedValueOnce(response)

      const result = await aibuilderApi.listConversations()

      expect(apiClient.get).toHaveBeenCalledWith('/ai/workflows/conversations')
      expect(result).toEqual([mockConversation])
    })

    it('should handle empty conversations list', async () => {
      const response: ConversationsListResponse = {
        data: [],
      }
      ;(apiClient.get as any).mockResolvedValueOnce(response)

      const result = await aibuilderApi.listConversations()

      expect(result).toEqual([])
    })

    it('should return conversations with different statuses', async () => {
      const conversations: Conversation[] = [
        { ...mockConversation, id: 'conv-1', status: 'active' },
        { ...mockConversation, id: 'conv-2', status: 'completed' },
        { ...mockConversation, id: 'conv-3', status: 'abandoned' },
      ]
      ;(apiClient.get as any).mockResolvedValueOnce({ data: conversations })

      const result = await aibuilderApi.listConversations()

      expect(result).toHaveLength(3)
      expect(result.map((c) => c.status)).toEqual(['active', 'completed', 'abandoned'])
    })
  })

  describe('apply', () => {
    it('should apply generated workflow', async () => {
      const workflowId = 'wf-123'
      ;(apiClient.post as any).mockResolvedValueOnce({ workflow_id: workflowId })

      const result = await aibuilderApi.apply('conv-123')

      expect(apiClient.post).toHaveBeenCalledWith(
        '/ai/workflows/conversations/conv-123/apply',
        {}
      )
      expect(result).toBe(workflowId)
    })

    it('should apply workflow with custom name', async () => {
      const workflowId = 'wf-123'
      ;(apiClient.post as any).mockResolvedValueOnce({ workflow_id: workflowId })

      const result = await aibuilderApi.apply('conv-123', 'My Custom Workflow')

      expect(apiClient.post).toHaveBeenCalledWith(
        '/ai/workflows/conversations/conv-123/apply',
        { workflow_name: 'My Custom Workflow' }
      )
      expect(result).toBe(workflowId)
    })

    it('should handle apply error', async () => {
      const error = new Error('Failed to create workflow')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(aibuilderApi.apply('conv-123')).rejects.toThrow('Failed to create workflow')
    })
  })

  describe('abandon', () => {
    it('should abandon conversation', async () => {
      (apiClient.post as any).mockResolvedValueOnce({})

      await aibuilderApi.abandon('conv-123')

      expect(apiClient.post).toHaveBeenCalledWith(
        '/ai/workflows/conversations/conv-123/abandon',
        {}
      )
    })

    it('should handle abandon error', async () => {
      const error = new Error('Conversation not found')
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(aibuilderApi.abandon('invalid')).rejects.toThrow('Conversation not found')
    })
  })
})
