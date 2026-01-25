import { describe, it, expect } from 'vitest'
import {
  parseWorkflowContent,
  parseWorkflowFile,
  validateWorkflowStructure,
} from './workflowParser'

describe('workflowParser', () => {
  describe('parseWorkflowContent - JSON', () => {
    it('should parse valid JSON workflow', () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Webhook' } },
          { id: 'node-2', type: 'action', position: { x: 100, y: 100 }, data: { label: 'HTTP Request' } },
        ],
        edges: [
          { id: 'edge-1', source: 'node-1', target: 'node-2' },
        ],
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(true)
      expect(result.nodes).toHaveLength(2)
      expect(result.edges).toHaveLength(1)
    })

    it('should parse workflow with definition wrapper', () => {
      const content = JSON.stringify({
        name: 'My Workflow',
        description: 'Test workflow',
        definition: {
          nodes: [
            { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
          ],
          edges: [],
        },
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(true)
      expect(result.name).toBe('My Workflow')
      expect(result.description).toBe('Test workflow')
      expect(result.nodes).toHaveLength(1)
    })

    it('should reject invalid JSON', () => {
      const content = '{ invalid json }'

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
      expect(result.error).toContain('Invalid JSON format')
    })

    it('should reject workflow with no nodes', () => {
      const content = JSON.stringify({ nodes: [], edges: [] })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
      expect(result.error).toContain('no nodes')
    })

    it('should reject nodes without id', () => {
      const content = JSON.stringify({
        nodes: [
          { type: 'trigger', position: { x: 0, y: 0 } },
        ],
        edges: [],
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
      expect(result.errorDetails).toBeDefined()
      expect(result.errorDetails?.some(e => e.includes("Missing required field 'id'"))).toBe(true)
    })

    it('should detect duplicate node IDs', () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 } },
          { id: 'node-1', type: 'action', position: { x: 100, y: 100 } },
        ],
        edges: [],
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
      expect(result.errorDetails?.some(e => e.includes('Duplicate node ID'))).toBe(true)
    })

    it('should detect edges referencing non-existent nodes', () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 } },
        ],
        edges: [
          { id: 'edge-1', source: 'node-1', target: 'node-missing' },
        ],
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
      expect(result.errorDetails?.some(e => e.includes('does not exist'))).toBe(true)
    })

    it('should detect self-loops', () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 } },
        ],
        edges: [
          { id: 'edge-1', source: 'node-1', target: 'node-1' },
        ],
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
      expect(result.errorDetails?.some(e => e.includes('Self-loop'))).toBe(true)
    })

    it('should convert backend node format to frontend format', () => {
      const content = JSON.stringify({
        nodes: [
          {
            id: 'node-1',
            type: 'trigger:webhook',
            position: { x: 0, y: 0 },
            data: { name: 'My Webhook', config: { path: '/api/hook' } },
          },
        ],
        edges: [],
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(true)
      expect(result.nodes?.[0].type).toBe('trigger')
      expect(result.nodes?.[0].data?.nodeType).toBe('webhook')
    })

    it('should auto-position nodes without positions and add warning', () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger' },
        ],
        edges: [],
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(true)
      expect(result.nodes?.[0].position).toBeDefined()
      expect(result.warnings?.some(w => w.includes('auto-positioned'))).toBe(true)
    })
  })

  describe('parseWorkflowContent - YAML', () => {
    it('should parse simple YAML (JSON subset)', () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
        ],
        edges: [],
      })

      const result = parseWorkflowContent(content, '.yaml')

      expect(result.success).toBe(true)
      expect(result.nodes).toHaveLength(1)
    })
  })

  describe('parseWorkflowFile', () => {
    it('should parse a valid workflow file', async () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
        ],
        edges: [],
      })
      const file = new File([content], 'workflow.json', { type: 'application/json' })

      const result = await parseWorkflowFile(file)

      expect(result.success).toBe(true)
      expect(result.nodes).toHaveLength(1)
    })

    it('should call progress callback', async () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
        ],
        edges: [],
      })
      const file = new File([content], 'workflow.json', { type: 'application/json' })
      const progressValues: number[] = []

      await parseWorkflowFile(file, (progress) => {
        progressValues.push(progress)
      })

      expect(progressValues.length).toBeGreaterThan(0)
      expect(progressValues[progressValues.length - 1]).toBe(100)
    })
  })

  describe('validateWorkflowStructure', () => {
    it('should warn when workflow has no trigger', () => {
      const nodes = [
        { id: 'node-1', type: 'action', position: { x: 0, y: 0 }, data: { label: 'Action' } },
      ]
      const edges: never[] = []

      const result = validateWorkflowStructure(nodes, edges)

      expect(result.valid).toBe(true)
      expect(result.warnings.some(w => w.includes('no trigger node'))).toBe(true)
    })

    it('should warn when workflow has multiple triggers', () => {
      const nodes = [
        { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Trigger 1' } },
        { id: 'node-2', type: 'trigger', position: { x: 100, y: 0 }, data: { label: 'Trigger 2' } },
      ]
      const edges: never[] = []

      const result = validateWorkflowStructure(nodes, edges)

      expect(result.warnings.some(w => w.includes('2 trigger nodes'))).toBe(true)
    })

    it('should warn about disconnected nodes', () => {
      const nodes = [
        { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Trigger' } },
        { id: 'node-2', type: 'action', position: { x: 100, y: 0 }, data: { label: 'Disconnected' } },
      ]
      const edges: never[] = []

      const result = validateWorkflowStructure(nodes, edges)

      expect(result.warnings.some(w => w.includes('not connected'))).toBe(true)
    })

    it('should not warn about connected nodes', () => {
      const nodes = [
        { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Trigger' } },
        { id: 'node-2', type: 'action', position: { x: 100, y: 0 }, data: { label: 'Action' } },
      ]
      const edges = [
        { id: 'edge-1', source: 'node-1', target: 'node-2' },
      ]

      const result = validateWorkflowStructure(nodes, edges)

      expect(result.warnings.filter(w => w.includes('not connected'))).toHaveLength(0)
    })
  })

  describe('Edge cases', () => {
    it('should handle empty object', () => {
      const content = JSON.stringify({})

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
    })

    it('should handle null content', () => {
      const content = 'null'

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
      expect(result.error).toContain('Invalid workflow structure')
    })

    it('should handle array instead of object', () => {
      const content = JSON.stringify([
        { id: 'node-1', type: 'trigger' },
      ])

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(false)
    })

    it('should add labels to nodes without labels', () => {
      const content = JSON.stringify({
        nodes: [
          { id: 'my-node-id', type: 'trigger', position: { x: 0, y: 0 }, data: {} },
        ],
        edges: [],
      })

      const result = parseWorkflowContent(content, '.json')

      expect(result.success).toBe(true)
      expect(result.nodes?.[0].data?.label).toBe('my-node-id')
    })
  })
})
