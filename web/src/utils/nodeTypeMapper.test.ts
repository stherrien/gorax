import { describe, it, expect } from 'vitest'
import {
  toBackendNodeType,
  toFrontendNodeType,
  serializeNodeForBackend,
  deserializeNodeFromBackend,
  serializeWorkflowForBackend,
  deserializeWorkflowFromBackend,
  isValidBackendNodeType,
  isTriggerType,
  BACKEND_NODE_TYPES,
} from './nodeTypeMapper'

describe('nodeTypeMapper', () => {
  describe('toBackendNodeType', () => {
    it('converts trigger:webhook correctly', () => {
      expect(toBackendNodeType('trigger', 'webhook')).toBe('trigger:webhook')
    })

    it('converts trigger:schedule correctly', () => {
      expect(toBackendNodeType('trigger', 'schedule')).toBe('trigger:schedule')
    })

    it('converts trigger:manual to webhook (manual uses webhook infrastructure)', () => {
      expect(toBackendNodeType('trigger', 'manual')).toBe('trigger:webhook')
    })

    it('converts action:http correctly', () => {
      expect(toBackendNodeType('action', 'http')).toBe('action:http')
    })

    it('converts action:script to action:code', () => {
      expect(toBackendNodeType('action', 'script')).toBe('action:code')
    })

    it('converts slack actions with slack: prefix', () => {
      expect(toBackendNodeType('action', 'slack_send_message')).toBe('slack:send_message')
      expect(toBackendNodeType('action', 'slack_send_dm')).toBe('slack:send_dm')
    })

    it('converts control:conditional to control:if', () => {
      expect(toBackendNodeType('control', 'conditional')).toBe('control:if')
    })

    it('converts control:loop correctly', () => {
      expect(toBackendNodeType('control', 'loop')).toBe('control:loop')
    })

    it('converts control:parallel correctly', () => {
      expect(toBackendNodeType('control', 'parallel')).toBe('control:parallel')
    })
  })

  describe('toFrontendNodeType', () => {
    it('converts trigger:webhook correctly', () => {
      const result = toFrontendNodeType('trigger:webhook')
      expect(result).toEqual({ category: 'trigger', nodeType: 'webhook' })
    })

    it('converts trigger:schedule correctly', () => {
      const result = toFrontendNodeType('trigger:schedule')
      expect(result).toEqual({ category: 'trigger', nodeType: 'schedule' })
    })

    it('converts action:code to script', () => {
      const result = toFrontendNodeType('action:code')
      expect(result).toEqual({ category: 'action', nodeType: 'script' })
    })

    it('converts slack:send_message correctly', () => {
      const result = toFrontendNodeType('slack:send_message')
      expect(result).toEqual({ category: 'action', nodeType: 'slack_send_message' })
    })

    it('converts control:if to conditional', () => {
      const result = toFrontendNodeType('control:if')
      expect(result).toEqual({ category: 'control', nodeType: 'conditional' })
    })
  })

  describe('serializeNodeForBackend', () => {
    it('serializes a trigger node correctly', () => {
      const frontendNode = {
        id: 'node-123',
        type: 'trigger',
        position: { x: 100, y: 200 },
        data: {
          nodeType: 'webhook',
          label: 'My Webhook',
          path: '/api/webhook',
        },
      }

      const result = serializeNodeForBackend(frontendNode)

      expect(result).toEqual({
        id: 'node-123',
        type: 'trigger:webhook',
        position: { x: 100, y: 200 },
        data: {
          name: 'My Webhook',
          config: { path: '/api/webhook' },
        },
      })
    })

    it('serializes an action node correctly', () => {
      const frontendNode = {
        id: 'node-456',
        type: 'action',
        position: { x: 300, y: 400 },
        data: {
          nodeType: 'http',
          label: 'API Call',
          url: 'https://api.example.com',
          method: 'POST',
        },
      }

      const result = serializeNodeForBackend(frontendNode)

      expect(result).toEqual({
        id: 'node-456',
        type: 'action:http',
        position: { x: 300, y: 400 },
        data: {
          name: 'API Call',
          config: { url: 'https://api.example.com', method: 'POST' },
        },
      })
    })

    it('serializes a control node correctly', () => {
      const frontendNode = {
        id: 'node-789',
        type: 'control',
        position: { x: 500, y: 600 },
        data: {
          nodeType: 'conditional',
          label: 'Check Status',
          condition: 'status === 200',
        },
      }

      const result = serializeNodeForBackend(frontendNode)

      expect(result).toEqual({
        id: 'node-789',
        type: 'control:if',
        position: { x: 500, y: 600 },
        data: {
          name: 'Check Status',
          config: { condition: 'status === 200' },
        },
      })
    })
  })

  describe('deserializeNodeFromBackend', () => {
    it('deserializes a trigger node correctly', () => {
      const backendNode = {
        id: 'node-123',
        type: 'trigger:webhook',
        position: { x: 100, y: 200 },
        data: {
          name: 'My Webhook',
          config: { path: '/api/webhook' },
        },
      }

      const result = deserializeNodeFromBackend(backendNode)

      expect(result).toEqual({
        id: 'node-123',
        type: 'trigger',
        position: { x: 100, y: 200 },
        data: {
          nodeType: 'webhook',
          label: 'My Webhook',
          path: '/api/webhook',
        },
      })
    })

    it('deserializes control:if to conditional', () => {
      const backendNode = {
        id: 'node-789',
        type: 'control:if',
        position: { x: 500, y: 600 },
        data: {
          name: 'Check Status',
          config: { condition: 'status === 200' },
        },
      }

      const result = deserializeNodeFromBackend(backendNode)

      expect(result.type).toBe('control')
      expect(result.data.nodeType).toBe('conditional')
    })
  })

  describe('serializeWorkflowForBackend', () => {
    it('serializes a complete workflow', () => {
      const frontendDefinition = {
        nodes: [
          {
            id: 'trigger-1',
            type: 'trigger',
            position: { x: 0, y: 0 },
            data: { nodeType: 'webhook', label: 'Start' },
          },
          {
            id: 'action-1',
            type: 'action',
            position: { x: 200, y: 0 },
            data: { nodeType: 'http', label: 'Call API' },
          },
        ],
        edges: [
          { id: 'e1', source: 'trigger-1', target: 'action-1' },
        ],
      }

      const result = serializeWorkflowForBackend(frontendDefinition)

      expect(result.nodes[0].type).toBe('trigger:webhook')
      expect(result.nodes[1].type).toBe('action:http')
      expect(result.edges).toHaveLength(1)
    })
  })

  describe('deserializeWorkflowFromBackend', () => {
    it('deserializes a complete workflow', () => {
      const backendDefinition = {
        nodes: [
          {
            id: 'trigger-1',
            type: 'trigger:webhook',
            position: { x: 0, y: 0 },
            data: { name: 'Start', config: {} },
          },
          {
            id: 'action-1',
            type: 'action:http',
            position: { x: 200, y: 0 },
            data: { name: 'Call API', config: {} },
          },
        ],
        edges: [
          { id: 'e1', source: 'trigger-1', target: 'action-1' },
        ],
      }

      const result = deserializeWorkflowFromBackend(backendDefinition)

      expect(result.nodes[0].type).toBe('trigger')
      expect(result.nodes[0].data.nodeType).toBe('webhook')
      expect(result.nodes[1].type).toBe('action')
      expect(result.nodes[1].data.nodeType).toBe('http')
    })
  })

  describe('isValidBackendNodeType', () => {
    it('returns true for valid backend types', () => {
      expect(isValidBackendNodeType('trigger:webhook')).toBe(true)
      expect(isValidBackendNodeType('trigger:schedule')).toBe(true)
      expect(isValidBackendNodeType('action:http')).toBe(true)
      expect(isValidBackendNodeType('control:if')).toBe(true)
    })

    it('returns false for invalid backend types', () => {
      expect(isValidBackendNodeType('trigger')).toBe(false)
      expect(isValidBackendNodeType('action')).toBe(false)
      expect(isValidBackendNodeType('invalid:type')).toBe(false)
    })
  })

  describe('isTriggerType', () => {
    it('returns true for trigger types', () => {
      expect(isTriggerType('trigger:webhook')).toBe(true)
      expect(isTriggerType('trigger:schedule')).toBe(true)
    })

    it('returns false for non-trigger types', () => {
      expect(isTriggerType('action:http')).toBe(false)
      expect(isTriggerType('control:if')).toBe(false)
    })
  })

  describe('BACKEND_NODE_TYPES constants', () => {
    it('has correct trigger type values', () => {
      expect(BACKEND_NODE_TYPES.TRIGGER_WEBHOOK).toBe('trigger:webhook')
      expect(BACKEND_NODE_TYPES.TRIGGER_SCHEDULE).toBe('trigger:schedule')
    })

    it('has correct action type values', () => {
      expect(BACKEND_NODE_TYPES.ACTION_HTTP).toBe('action:http')
      expect(BACKEND_NODE_TYPES.ACTION_TRANSFORM).toBe('action:transform')
      expect(BACKEND_NODE_TYPES.ACTION_CODE).toBe('action:code')
    })

    it('has correct control type values', () => {
      expect(BACKEND_NODE_TYPES.CONTROL_IF).toBe('control:if')
      expect(BACKEND_NODE_TYPES.CONTROL_LOOP).toBe('control:loop')
      expect(BACKEND_NODE_TYPES.CONTROL_PARALLEL).toBe('control:parallel')
    })
  })
})
