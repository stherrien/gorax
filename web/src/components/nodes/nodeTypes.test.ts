import { describe, it, expect } from 'vitest'
import { nodeTypes } from './nodeTypes'

describe('nodeTypes', () => {
  describe('Node type registration', () => {
    it('should export nodeTypes object', () => {
      expect(nodeTypes).toBeDefined()
      expect(typeof nodeTypes).toBe('object')
    })

    it('should register trigger node type', () => {
      expect(nodeTypes.trigger).toBeDefined()
      expect(typeof nodeTypes.trigger).toBe('function')
    })

    it('should register action node type', () => {
      expect(nodeTypes.action).toBeDefined()
      expect(typeof nodeTypes.action).toBe('function')
    })

    it('should register conditional node type', () => {
      expect(nodeTypes.conditional).toBeDefined()
      expect(typeof nodeTypes.conditional).toBe('function')
    })
  })

  describe('Node types with execution status', () => {
    it('should wrap trigger node with ExecutionStatusNode', () => {
      const TriggerNode = nodeTypes.trigger
      expect(TriggerNode.name).toBe('WrappedNode')
    })

    it('should wrap action node with ExecutionStatusNode', () => {
      const ActionNode = nodeTypes.action
      expect(ActionNode.name).toBe('WrappedNode')
    })

    it('should wrap conditional node with ExecutionStatusNode', () => {
      const ConditionalNode = nodeTypes.conditional
      expect(ConditionalNode.name).toBe('WrappedNode')
    })
  })

  describe('Backward compatibility', () => {
    it('should have all required node types', () => {
      const requiredTypes = ['trigger', 'action', 'conditional']
      requiredTypes.forEach((type) => {
        expect(nodeTypes).toHaveProperty(type)
        expect(nodeTypes[type]).toBeDefined()
      })
    })

    it('should export node types that can be used in ReactFlow', () => {
      // All node types should be React components (functions)
      Object.values(nodeTypes).forEach((NodeComponent) => {
        expect(typeof NodeComponent).toBe('function')
      })
    })
  })
})
