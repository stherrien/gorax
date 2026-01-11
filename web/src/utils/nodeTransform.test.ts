import { describe, it, expect } from 'vitest'
import {
  parseFullType,
  buildFullType,
  getReactFlowCategory,
  transformNodeToBackend,
  transformNodeToFrontend,
  transformEdgeToBackend,
  transformEdgeToFrontend,
  transformWorkflowToBackend,
  transformWorkflowToFrontend,
  type FrontendNode,
  type BackendNode,
  type FrontendEdge,
  type BackendEdge,
} from './nodeTransform'

describe('nodeTransform', () => {
  describe('parseFullType', () => {
    it('should parse standard full type', () => {
      expect(parseFullType('trigger:webhook')).toEqual(['trigger', 'webhook'])
    })

    it('should parse action type', () => {
      expect(parseFullType('action:http')).toEqual(['action', 'http'])
    })

    it('should parse control type', () => {
      expect(parseFullType('control:loop')).toEqual(['control', 'loop'])
    })

    it('should handle single part type', () => {
      expect(parseFullType('trigger')).toEqual(['trigger', 'trigger'])
    })

    it('should handle empty string', () => {
      expect(parseFullType('')).toEqual(['action', 'unknown'])
    })

    it('should handle type with multiple colons', () => {
      expect(parseFullType('action:slack:send_message')).toEqual(['action', 'slack:send_message'])
    })
  })

  describe('buildFullType', () => {
    it('should build full type from category and nodeType', () => {
      expect(buildFullType('trigger', 'webhook')).toBe('trigger:webhook')
    })

    it('should handle already full type in nodeType', () => {
      expect(buildFullType('trigger', 'trigger:webhook')).toBe('trigger:webhook')
    })

    it('should handle empty category', () => {
      expect(buildFullType('', 'webhook')).toBe('webhook')
    })

    it('should handle empty nodeType', () => {
      expect(buildFullType('trigger', '')).toBe('trigger')
    })

    it('should handle both empty', () => {
      expect(buildFullType('', '')).toBe('action:unknown')
    })
  })

  describe('getReactFlowCategory', () => {
    it('should return trigger for trigger types', () => {
      expect(getReactFlowCategory('trigger:webhook')).toBe('trigger')
    })

    it('should return action for action types', () => {
      expect(getReactFlowCategory('action:http')).toBe('action')
    })

    it('should return ai for ai types', () => {
      expect(getReactFlowCategory('ai:chat')).toBe('ai')
    })

    it('should return control for control types', () => {
      expect(getReactFlowCategory('control:loop')).toBe('control')
    })

    it('should map conditional to control', () => {
      expect(getReactFlowCategory('conditional:if')).toBe('control')
    })
  })

  describe('transformNodeToBackend', () => {
    it('should transform frontend node to backend format', () => {
      const frontendNode: FrontendNode = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 100, y: 200 },
        data: {
          label: 'My Webhook',
          type: 'trigger:webhook',
          nodeType: 'webhook',
          path: '/api/webhook',
          authType: 'signature',
        },
      }

      const result = transformNodeToBackend(frontendNode)

      expect(result).toEqual({
        id: 'node-1',
        type: 'trigger:webhook',
        position: { x: 100, y: 200 },
        data: {
          name: 'My Webhook',
          config: {
            path: '/api/webhook',
            authType: 'signature',
          },
        },
      })
    })

    it('should use data.type for full type', () => {
      const frontendNode: FrontendNode = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: {
          label: 'HTTP Request',
          type: 'action:http',
          nodeType: 'http',
          url: 'https://api.example.com',
        },
      }

      const result = transformNodeToBackend(frontendNode)

      expect(result.type).toBe('action:http')
    })

    it('should construct type from node.type and nodeType if data.type is missing', () => {
      const frontendNode: FrontendNode = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: {
          label: 'Transform',
          type: '', // Empty
          nodeType: 'transform',
        },
      }

      const result = transformNodeToBackend(frontendNode)

      expect(result.type).toBe('action:transform')
    })

    it('should default to Unnamed Node if label is missing', () => {
      const frontendNode: FrontendNode = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: {
          label: '',
          type: 'action:http',
          nodeType: 'http',
        },
      }

      const result = transformNodeToBackend(frontendNode)

      expect(result.data.name).toBe('Unnamed Node')
    })

    it('should handle undefined position', () => {
      const frontendNode = {
        id: 'node-1',
        type: 'action',
        data: {
          label: 'Test',
          type: 'action:http',
          nodeType: 'http',
        },
      } as unknown as FrontendNode

      const result = transformNodeToBackend(frontendNode)

      expect(result.position).toEqual({ x: 0, y: 0 })
    })

    it('should exclude label, type, nodeType from config', () => {
      const frontendNode: FrontendNode = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: {
          label: 'Test',
          type: 'action:http',
          nodeType: 'http',
          url: 'https://example.com',
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
        },
      }

      const result = transformNodeToBackend(frontendNode)

      expect(result.data.config).toEqual({
        url: 'https://example.com',
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
      expect(result.data.config).not.toHaveProperty('label')
      expect(result.data.config).not.toHaveProperty('type')
      expect(result.data.config).not.toHaveProperty('nodeType')
    })
  })

  describe('transformNodeToFrontend', () => {
    it('should transform backend node to frontend format', () => {
      const backendNode: BackendNode = {
        id: 'node-1',
        type: 'trigger:webhook',
        position: { x: 100, y: 200 },
        data: {
          name: 'My Webhook',
          config: {
            path: '/api/webhook',
            authType: 'signature',
          },
        },
      }

      const result = transformNodeToFrontend(backendNode)

      expect(result).toEqual({
        id: 'node-1',
        type: 'trigger', // ReactFlow category
        position: { x: 100, y: 200 },
        data: {
          label: 'My Webhook',
          type: 'trigger:webhook', // Full type preserved
          nodeType: 'webhook',
          path: '/api/webhook',
          authType: 'signature',
        },
      })
    })

    it('should spread config to top level of data', () => {
      const backendNode: BackendNode = {
        id: 'node-1',
        type: 'action:http',
        position: { x: 0, y: 0 },
        data: {
          name: 'HTTP Request',
          config: {
            url: 'https://api.example.com',
            method: 'GET',
            timeout: 30000,
          },
        },
      }

      const result = transformNodeToFrontend(backendNode)

      expect(result.data.url).toBe('https://api.example.com')
      expect(result.data.method).toBe('GET')
      expect(result.data.timeout).toBe(30000)
    })

    it('should default to Unnamed Node if name is missing', () => {
      const backendNode: BackendNode = {
        id: 'node-1',
        type: 'action:http',
        position: { x: 0, y: 0 },
        data: {
          name: '',
          config: {},
        },
      }

      const result = transformNodeToFrontend(backendNode)

      expect(result.data.label).toBe('Unnamed Node')
    })

    it('should handle missing config', () => {
      const backendNode = {
        id: 'node-1',
        type: 'action:http',
        position: { x: 0, y: 0 },
        data: {
          name: 'Test',
        },
      } as unknown as BackendNode

      const result = transformNodeToFrontend(backendNode)

      expect(result.data).toEqual({
        label: 'Test',
        type: 'action:http',
        nodeType: 'http',
      })
    })

    it('should handle missing data', () => {
      const backendNode = {
        id: 'node-1',
        type: 'action:http',
        position: { x: 0, y: 0 },
      } as unknown as BackendNode

      const result = transformNodeToFrontend(backendNode)

      expect(result.data.label).toBe('Unnamed Node')
    })
  })

  describe('round-trip consistency', () => {
    it('should preserve data through frontend -> backend -> frontend', () => {
      const original: FrontendNode = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 100, y: 200 },
        data: {
          label: 'My Webhook',
          type: 'trigger:webhook',
          nodeType: 'webhook',
          path: '/api/webhook',
          authType: 'signature',
          secret: 'my-secret',
        },
      }

      const backend = transformNodeToBackend(original)
      const restored = transformNodeToFrontend(backend)

      expect(restored.id).toBe(original.id)
      expect(restored.type).toBe(original.type)
      expect(restored.position).toEqual(original.position)
      expect(restored.data.label).toBe(original.data.label)
      expect(restored.data.type).toBe(original.data.type)
      expect(restored.data.nodeType).toBe(original.data.nodeType)
      expect(restored.data.path).toBe(original.data.path)
      expect(restored.data.authType).toBe(original.data.authType)
      expect(restored.data.secret).toBe(original.data.secret)
    })

    it('should preserve data through backend -> frontend -> backend', () => {
      const original: BackendNode = {
        id: 'node-1',
        type: 'action:http',
        position: { x: 50, y: 150 },
        data: {
          name: 'API Call',
          config: {
            url: 'https://api.example.com',
            method: 'POST',
            body: '{"key": "value"}',
          },
        },
      }

      const frontend = transformNodeToFrontend(original)
      const restored = transformNodeToBackend(frontend)

      expect(restored.id).toBe(original.id)
      expect(restored.type).toBe(original.type)
      expect(restored.position).toEqual(original.position)
      expect(restored.data.name).toBe(original.data.name)
      expect(restored.data.config).toEqual(original.data.config)
    })
  })

  describe('transformEdgeToBackend', () => {
    it('should transform edge correctly', () => {
      const frontendEdge: FrontendEdge = {
        id: 'e1-2',
        source: 'node-1',
        target: 'node-2',
        sourceHandle: 'output',
        targetHandle: 'input',
      }

      const result = transformEdgeToBackend(frontendEdge)

      expect(result).toEqual({
        id: 'e1-2',
        source: 'node-1',
        target: 'node-2',
        sourceHandle: 'output',
        targetHandle: 'input',
      })
    })

    it('should exclude undefined handles', () => {
      const frontendEdge: FrontendEdge = {
        id: 'e1-2',
        source: 'node-1',
        target: 'node-2',
      }

      const result = transformEdgeToBackend(frontendEdge)

      expect(result).toEqual({
        id: 'e1-2',
        source: 'node-1',
        target: 'node-2',
      })
      expect(Object.keys(result)).not.toContain('sourceHandle')
      expect(Object.keys(result)).not.toContain('targetHandle')
    })
  })

  describe('transformEdgeToFrontend', () => {
    it('should transform edge correctly', () => {
      const backendEdge: BackendEdge = {
        id: 'e1-2',
        source: 'node-1',
        target: 'node-2',
        sourceHandle: 'output',
        targetHandle: 'input',
      }

      const result = transformEdgeToFrontend(backendEdge)

      expect(result).toEqual({
        id: 'e1-2',
        source: 'node-1',
        target: 'node-2',
        sourceHandle: 'output',
        targetHandle: 'input',
      })
    })
  })

  describe('transformWorkflowToBackend', () => {
    it('should transform entire workflow definition', () => {
      const nodes: FrontendNode[] = [
        {
          id: 'node-1',
          type: 'trigger',
          position: { x: 0, y: 0 },
          data: {
            label: 'Webhook',
            type: 'trigger:webhook',
            nodeType: 'webhook',
          },
        },
        {
          id: 'node-2',
          type: 'action',
          position: { x: 200, y: 0 },
          data: {
            label: 'HTTP',
            type: 'action:http',
            nodeType: 'http',
            url: 'https://api.example.com',
          },
        },
      ]

      const edges: FrontendEdge[] = [
        {
          id: 'e1-2',
          source: 'node-1',
          target: 'node-2',
        },
      ]

      const result = transformWorkflowToBackend(nodes, edges)

      expect(result.nodes).toHaveLength(2)
      expect(result.nodes[0].type).toBe('trigger:webhook')
      expect(result.nodes[1].type).toBe('action:http')
      expect(result.nodes[1].data.config.url).toBe('https://api.example.com')
      expect(result.edges).toHaveLength(1)
    })
  })

  describe('transformWorkflowToFrontend', () => {
    it('should transform entire workflow definition', () => {
      const definition = {
        nodes: [
          {
            id: 'node-1',
            type: 'trigger:webhook',
            position: { x: 0, y: 0 },
            data: { name: 'Webhook', config: {} },
          },
          {
            id: 'node-2',
            type: 'action:http',
            position: { x: 200, y: 0 },
            data: { name: 'HTTP', config: { url: 'https://api.example.com' } },
          },
        ],
        edges: [
          {
            id: 'e1-2',
            source: 'node-1',
            target: 'node-2',
          },
        ],
      }

      const result = transformWorkflowToFrontend(definition)

      expect(result.nodes).toHaveLength(2)
      expect(result.nodes[0].type).toBe('trigger') // ReactFlow category
      expect(result.nodes[0].data.type).toBe('trigger:webhook') // Full type
      expect(result.nodes[1].data.url).toBe('https://api.example.com') // Config spread
      expect(result.edges).toHaveLength(1)
    })

    it('should handle null definition', () => {
      const result = transformWorkflowToFrontend(null)

      expect(result).toEqual({ nodes: [], edges: [] })
    })

    it('should handle undefined definition', () => {
      const result = transformWorkflowToFrontend(undefined)

      expect(result).toEqual({ nodes: [], edges: [] })
    })

    it('should handle empty definition', () => {
      const result = transformWorkflowToFrontend({ nodes: [], edges: [] })

      expect(result).toEqual({ nodes: [], edges: [] })
    })
  })
})
