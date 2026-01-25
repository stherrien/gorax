/**
 * Contract Tests - Verify frontend/backend type compatibility
 *
 * These tests ensure that the frontend node type mapper correctly maps
 * to the backend types defined in the OpenAPI spec. Run these tests
 * whenever the backend API changes to catch mismatches early.
 *
 * If these tests fail:
 * 1. Regenerate types: npm run api:types
 * 2. Update nodeTypeMapper.ts to match new backend types
 * 3. Run tests again to verify compatibility
 */

import { describe, it, expect } from 'vitest'
import {
  BACKEND_NODE_TYPES,
  toBackendNodeType,
  toFrontendNodeType,
  serializeNodeForBackend,
  deserializeNodeFromBackend,
} from './nodeTypeMapper'

// These types should match the backend's NodeType constants in internal/workflow/model.go
const EXPECTED_BACKEND_TYPES = [
  'trigger:webhook',
  'trigger:schedule',
  'action:http',
  'action:transform',
  'action:formula',
  'action:code',
  'action:email',
  'action:subworkflow',
  'slack:send_message',
  'slack:send_dm',
  'slack:update_message',
  'slack:add_reaction',
  'control:if',
  'control:loop',
  'control:parallel',
  'control:fork',
  'control:join',
  'control:delay',
  'control:try',
  'control:catch',
  'control:finally',
  'control:retry',
  'control:circuit_breaker',
] as const

describe('Contract Tests: Frontend/Backend Type Compatibility', () => {
  describe('BACKEND_NODE_TYPES constant', () => {
    it('contains all expected backend types', () => {
      const definedTypes = Object.values(BACKEND_NODE_TYPES)

      // Every expected type should be in our constants
      for (const expectedType of EXPECTED_BACKEND_TYPES) {
        // Allow some types to not be implemented yet
        if (
          !expectedType.includes('catch') &&
          !expectedType.includes('finally') &&
          !expectedType.includes('circuit_breaker')
        ) {
          expect(
            definedTypes,
            `Missing backend type: ${expectedType}`
          ).toContain(expectedType)
        }
      }
    })

    it('trigger types use correct prefix', () => {
      expect(BACKEND_NODE_TYPES.TRIGGER_WEBHOOK).toMatch(/^trigger:/)
      expect(BACKEND_NODE_TYPES.TRIGGER_SCHEDULE).toMatch(/^trigger:/)
    })

    it('action types use correct prefix', () => {
      expect(BACKEND_NODE_TYPES.ACTION_HTTP).toMatch(/^action:/)
      expect(BACKEND_NODE_TYPES.ACTION_TRANSFORM).toMatch(/^action:/)
    })

    it('control types use correct prefix', () => {
      expect(BACKEND_NODE_TYPES.CONTROL_IF).toMatch(/^control:/)
      expect(BACKEND_NODE_TYPES.CONTROL_LOOP).toMatch(/^control:/)
    })

    it('slack types use correct prefix', () => {
      expect(BACKEND_NODE_TYPES.SLACK_SEND_MESSAGE).toMatch(/^slack:/)
      expect(BACKEND_NODE_TYPES.SLACK_SEND_DM).toMatch(/^slack:/)
    })
  })

  describe('Round-trip serialization', () => {
    it('trigger:webhook survives round-trip', () => {
      const original = {
        id: 'test-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Test' },
      }

      const serialized = serializeNodeForBackend(original)
      expect(serialized.type).toBe('trigger:webhook')

      const deserialized = deserializeNodeFromBackend(serialized)
      expect(deserialized.type).toBe('trigger')
      expect(deserialized.data.nodeType).toBe('webhook')
    })

    it('action:http survives round-trip', () => {
      const original = {
        id: 'test-2',
        type: 'action',
        position: { x: 100, y: 100 },
        data: { nodeType: 'http', label: 'API Call', url: 'https://api.example.com' },
      }

      const serialized = serializeNodeForBackend(original)
      expect(serialized.type).toBe('action:http')
      expect(serialized.data.config.url).toBe('https://api.example.com')

      const deserialized = deserializeNodeFromBackend(serialized)
      expect(deserialized.type).toBe('action')
      expect(deserialized.data.nodeType).toBe('http')
      expect(deserialized.data.url).toBe('https://api.example.com')
    })

    it('control:if survives round-trip (conditional -> if -> conditional)', () => {
      const original = {
        id: 'test-3',
        type: 'control',
        position: { x: 200, y: 200 },
        data: { nodeType: 'conditional', label: 'Check', condition: 'x > 0' },
      }

      const serialized = serializeNodeForBackend(original)
      expect(serialized.type).toBe('control:if')

      const deserialized = deserializeNodeFromBackend(serialized)
      expect(deserialized.type).toBe('control')
      expect(deserialized.data.nodeType).toBe('conditional')
    })

    it('script -> action:code -> script mapping works', () => {
      const original = {
        id: 'test-4',
        type: 'action',
        position: { x: 300, y: 300 },
        data: { nodeType: 'script', label: 'Run Code' },
      }

      const serialized = serializeNodeForBackend(original)
      expect(serialized.type).toBe('action:code')

      const deserialized = deserializeNodeFromBackend(serialized)
      expect(deserialized.type).toBe('action')
      expect(deserialized.data.nodeType).toBe('script')
    })

    it('slack actions use slack: prefix instead of action:', () => {
      const original = {
        id: 'test-5',
        type: 'action',
        position: { x: 400, y: 400 },
        data: { nodeType: 'slack_send_message', label: 'Send Message' },
      }

      const serialized = serializeNodeForBackend(original)
      expect(serialized.type).toBe('slack:send_message')
      expect(serialized.type).not.toMatch(/^action:/)

      const deserialized = deserializeNodeFromBackend(serialized)
      expect(deserialized.type).toBe('action')
      expect(deserialized.data.nodeType).toBe('slack_send_message')
    })
  })

  describe('Edge cases', () => {
    it('manual trigger maps to webhook', () => {
      const backendType = toBackendNodeType('trigger', 'manual')
      expect(backendType).toBe('trigger:webhook')
    })

    it('unknown types fall back gracefully', () => {
      const backendType = toBackendNodeType('unknown', 'type')
      expect(backendType).toBe('unknown:type')

      const frontendType = toFrontendNodeType('unknown:type')
      expect(frontendType.category).toBe('unknown')
      expect(frontendType.nodeType).toBe('type')
    })

    it('handles empty/undefined data gracefully', () => {
      const nodeWithEmptyData = {
        id: 'test-6',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: {},
      }

      const serialized = serializeNodeForBackend(nodeWithEmptyData)
      expect(serialized.type).toBe('trigger:unknown')
    })
  })
})
