import { describe, it, expect } from 'vitest'
import { detectCycles, isValidDAG, getTopologicalOrder } from './dagValidation'
import { Node, Edge } from '@xyflow/react'

describe('dagValidation', () => {
  describe('detectCycles', () => {
    it('should return empty array for valid DAG with no cycles', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'B', target: 'C' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles).toEqual([])
    })

    it('should detect simple cycle (A -> B -> A)', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'B', target: 'A' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles.length).toBeGreaterThan(0)
      expect(cycles[0]).toContain('A')
      expect(cycles[0]).toContain('B')
    })

    it('should detect complex cycle (A -> B -> C -> A)', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'B', target: 'C' },
        { id: 'e3', source: 'C', target: 'A' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles.length).toBeGreaterThan(0)
      expect(cycles[0]).toContain('A')
      expect(cycles[0]).toContain('B')
      expect(cycles[0]).toContain('C')
    })

    it('should detect self-loop (A -> A)', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'A' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles.length).toBeGreaterThan(0)
      expect(cycles[0]).toEqual(['A', 'A'])
    })

    it('should handle disconnected components without cycles', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
        { id: 'D', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'C', target: 'D' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles).toEqual([])
    })

    it('should detect cycle in one component while other component is valid', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
        { id: 'D', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'C', target: 'D' },
        { id: 'e3', source: 'D', target: 'C' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles.length).toBeGreaterThan(0)
      expect(cycles[0]).toContain('C')
      expect(cycles[0]).toContain('D')
    })

    it('should return empty array for empty graph', () => {
      const nodes: Node[] = []
      const edges: Edge[] = []

      const cycles = detectCycles(nodes, edges)

      expect(cycles).toEqual([])
    })

    it('should return empty array for single node with no edges', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = []

      const cycles = detectCycles(nodes, edges)

      expect(cycles).toEqual([])
    })

    it('should handle parallel branches (fan-out/fan-in) without false positives', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
        { id: 'D', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'A', target: 'C' },
        { id: 'e3', source: 'B', target: 'D' },
        { id: 'e4', source: 'C', target: 'D' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles).toEqual([])
    })

    it('should handle conditional branches without false positives', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
        { id: 'D', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'A', target: 'C' },
        { id: 'e3', source: 'B', target: 'D' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles).toEqual([])
    })

    it('should detect multiple independent cycles', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
        { id: 'D', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'B', target: 'A' },
        { id: 'e3', source: 'C', target: 'D' },
        { id: 'e4', source: 'D', target: 'C' },
      ]

      const cycles = detectCycles(nodes, edges)

      expect(cycles.length).toBe(2)
    })
  })

  describe('isValidDAG', () => {
    it('should return true for valid DAG', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'B', target: 'C' },
      ]

      expect(isValidDAG(nodes, edges)).toBe(true)
    })

    it('should return false when cycles exist', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'B', target: 'A' },
      ]

      expect(isValidDAG(nodes, edges)).toBe(false)
    })

    it('should return true for empty graph', () => {
      const nodes: Node[] = []
      const edges: Edge[] = []

      expect(isValidDAG(nodes, edges)).toBe(true)
    })

    it('should return true for single node', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = []

      expect(isValidDAG(nodes, edges)).toBe(true)
    })

    it('should return false for self-loop', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'A' },
      ]

      expect(isValidDAG(nodes, edges)).toBe(false)
    })
  })

  describe('getTopologicalOrder', () => {
    it('should return topological order for valid DAG', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'B', target: 'C' },
      ]

      const result = getTopologicalOrder(nodes, edges)

      expect(result.success).toBe(true)
      expect(result.order).toEqual(['A', 'B', 'C'])
    })

    it('should return error when cycle exists', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'B', target: 'A' },
      ]

      const result = getTopologicalOrder(nodes, edges)

      expect(result.success).toBe(false)
      expect(result.error).toBeDefined()
      expect(result.order).toEqual([])
    })

    it('should handle disconnected components', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
        { id: 'D', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'C', target: 'D' },
      ]

      const result = getTopologicalOrder(nodes, edges)

      expect(result.success).toBe(true)
      expect(result.order).toHaveLength(4)
      // A comes before B
      expect(result.order.indexOf('A')).toBeLessThan(result.order.indexOf('B'))
      // C comes before D
      expect(result.order.indexOf('C')).toBeLessThan(result.order.indexOf('D'))
    })

    it('should handle empty graph', () => {
      const nodes: Node[] = []
      const edges: Edge[] = []

      const result = getTopologicalOrder(nodes, edges)

      expect(result.success).toBe(true)
      expect(result.order).toEqual([])
    })

    it('should handle single node', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = []

      const result = getTopologicalOrder(nodes, edges)

      expect(result.success).toBe(true)
      expect(result.order).toEqual(['A'])
    })

    it('should handle complex DAG with multiple valid orders', () => {
      const nodes: Node[] = [
        { id: 'A', position: { x: 0, y: 0 }, data: {} },
        { id: 'B', position: { x: 0, y: 0 }, data: {} },
        { id: 'C', position: { x: 0, y: 0 }, data: {} },
        { id: 'D', position: { x: 0, y: 0 }, data: {} },
      ]
      const edges: Edge[] = [
        { id: 'e1', source: 'A', target: 'B' },
        { id: 'e2', source: 'A', target: 'C' },
        { id: 'e3', source: 'B', target: 'D' },
        { id: 'e4', source: 'C', target: 'D' },
      ]

      const result = getTopologicalOrder(nodes, edges)

      expect(result.success).toBe(true)
      expect(result.order).toHaveLength(4)
      // A must come first
      expect(result.order[0]).toBe('A')
      // D must come last
      expect(result.order[3]).toBe('D')
      // B and C must come before D
      expect(result.order.indexOf('B')).toBeLessThan(result.order.indexOf('D'))
      expect(result.order.indexOf('C')).toBeLessThan(result.order.indexOf('D'))
    })
  })

  describe('Performance', () => {
    it('should handle large graphs efficiently (O(V+E))', () => {
      const nodeCount = 1000
      const nodes: Node[] = Array.from({ length: nodeCount }, (_, i) => ({
        id: `node-${i}`,
        position: { x: 0, y: 0 },
        data: {},
      }))

      const edges: Edge[] = Array.from({ length: nodeCount - 1 }, (_, i) => ({
        id: `edge-${i}`,
        source: `node-${i}`,
        target: `node-${i + 1}`,
      }))

      const startTime = performance.now()
      const result = isValidDAG(nodes, edges)
      const endTime = performance.now()

      expect(result).toBe(true)
      // Should complete in under 100ms for 1000 nodes
      expect(endTime - startTime).toBeLessThan(100)
    })
  })
})
