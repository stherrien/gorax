import { describe, it, expect, beforeEach } from 'vitest'
import { useExecutionTraceStore } from './executionTraceStore'
import type { StepInfo } from '../lib/websocket'

describe('executionTraceStore', () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    useExecutionTraceStore.getState().reset()
  })

  describe('Node status tracking', () => {
    it('should initialize with empty node statuses', () => {
      const { nodeStatuses } = useExecutionTraceStore.getState()
      expect(nodeStatuses).toEqual({})
    })

    it('should set node status to pending', () => {
      const { setNodeStatus } = useExecutionTraceStore.getState()

      setNodeStatus('node-1', 'pending')

      const { nodeStatuses } = useExecutionTraceStore.getState()
      expect(nodeStatuses['node-1']).toBe('pending')
    })

    it('should set node status to running', () => {
      const { setNodeStatus } = useExecutionTraceStore.getState()

      setNodeStatus('node-1', 'running')

      const { nodeStatuses } = useExecutionTraceStore.getState()
      expect(nodeStatuses['node-1']).toBe('running')
    })

    it('should set node status to completed', () => {
      const { setNodeStatus } = useExecutionTraceStore.getState()

      setNodeStatus('node-1', 'completed')

      const { nodeStatuses } = useExecutionTraceStore.getState()
      expect(nodeStatuses['node-1']).toBe('completed')
    })

    it('should set node status to failed', () => {
      const { setNodeStatus } = useExecutionTraceStore.getState()

      setNodeStatus('node-1', 'failed')

      const { nodeStatuses } = useExecutionTraceStore.getState()
      expect(nodeStatuses['node-1']).toBe('failed')
    })

    it('should track multiple node statuses independently', () => {
      const { setNodeStatus } = useExecutionTraceStore.getState()

      setNodeStatus('node-1', 'completed')
      setNodeStatus('node-2', 'running')
      setNodeStatus('node-3', 'pending')

      const { nodeStatuses } = useExecutionTraceStore.getState()
      expect(nodeStatuses['node-1']).toBe('completed')
      expect(nodeStatuses['node-2']).toBe('running')
      expect(nodeStatuses['node-3']).toBe('pending')
    })

    it('should update existing node status', () => {
      const { setNodeStatus } = useExecutionTraceStore.getState()

      setNodeStatus('node-1', 'pending')
      setNodeStatus('node-1', 'running')
      setNodeStatus('node-1', 'completed')

      const { nodeStatuses } = useExecutionTraceStore.getState()
      expect(nodeStatuses['node-1']).toBe('completed')
    })
  })

  describe('Step logs storage', () => {
    it('should initialize with empty step logs', () => {
      const { stepLogs } = useExecutionTraceStore.getState()
      expect(stepLogs).toEqual({})
    })

    it('should store step log with input data', () => {
      const { addStepLog } = useExecutionTraceStore.getState()

      const stepInfo: StepInfo = {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'http',
        status: 'completed',
        output_data: { result: 'success' },
        duration_ms: 150,
      }

      addStepLog('node-1', stepInfo)

      const { stepLogs } = useExecutionTraceStore.getState()
      expect(stepLogs['node-1']).toHaveLength(1)
      expect(stepLogs['node-1'][0]).toEqual(stepInfo)
    })

    it('should store step log with error', () => {
      const { addStepLog } = useExecutionTraceStore.getState()

      const stepInfo: StepInfo = {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'http',
        status: 'failed',
        error: 'Connection timeout',
        duration_ms: 5000,
      }

      addStepLog('node-1', stepInfo)

      const { stepLogs } = useExecutionTraceStore.getState()
      expect(stepLogs['node-1']).toHaveLength(1)
      expect(stepLogs['node-1'][0].error).toBe('Connection timeout')
    })

    it('should store multiple step logs for same node', () => {
      const { addStepLog } = useExecutionTraceStore.getState()

      const step1: StepInfo = {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'http',
        status: 'completed',
        duration_ms: 100,
      }

      const step2: StepInfo = {
        step_id: 'step-2',
        node_id: 'node-1',
        node_type: 'http',
        status: 'completed',
        duration_ms: 120,
      }

      addStepLog('node-1', step1)
      addStepLog('node-1', step2)

      const { stepLogs } = useExecutionTraceStore.getState()
      expect(stepLogs['node-1']).toHaveLength(2)
      expect(stepLogs['node-1'][0]).toEqual(step1)
      expect(stepLogs['node-1'][1]).toEqual(step2)
    })

    it('should store step logs for multiple nodes independently', () => {
      const { addStepLog } = useExecutionTraceStore.getState()

      const step1: StepInfo = {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'http',
        status: 'completed',
        duration_ms: 100,
      }

      const step2: StepInfo = {
        step_id: 'step-2',
        node_id: 'node-2',
        node_type: 'transform',
        status: 'completed',
        duration_ms: 50,
      }

      addStepLog('node-1', step1)
      addStepLog('node-2', step2)

      const { stepLogs } = useExecutionTraceStore.getState()
      expect(stepLogs['node-1']).toHaveLength(1)
      expect(stepLogs['node-2']).toHaveLength(1)
      expect(stepLogs['node-1'][0]).toEqual(step1)
      expect(stepLogs['node-2'][0]).toEqual(step2)
    })
  })

  describe('Edge animation state management', () => {
    it('should initialize with empty animated edges', () => {
      const { animatedEdges } = useExecutionTraceStore.getState()
      expect(animatedEdges).toEqual(new Set())
    })

    it('should add edge to animated edges', () => {
      const { setEdgeAnimated } = useExecutionTraceStore.getState()

      setEdgeAnimated('edge-1', true)

      const { animatedEdges } = useExecutionTraceStore.getState()
      expect(animatedEdges.has('edge-1')).toBe(true)
    })

    it('should remove edge from animated edges', () => {
      const { setEdgeAnimated } = useExecutionTraceStore.getState()

      setEdgeAnimated('edge-1', true)
      setEdgeAnimated('edge-1', false)

      const { animatedEdges } = useExecutionTraceStore.getState()
      expect(animatedEdges.has('edge-1')).toBe(false)
    })

    it('should track multiple animated edges', () => {
      const { setEdgeAnimated } = useExecutionTraceStore.getState()

      setEdgeAnimated('edge-1', true)
      setEdgeAnimated('edge-2', true)
      setEdgeAnimated('edge-3', true)

      const { animatedEdges } = useExecutionTraceStore.getState()
      expect(animatedEdges.has('edge-1')).toBe(true)
      expect(animatedEdges.has('edge-2')).toBe(true)
      expect(animatedEdges.has('edge-3')).toBe(true)
      expect(animatedEdges.size).toBe(3)
    })

    it('should handle removing non-existent edge gracefully', () => {
      const { setEdgeAnimated } = useExecutionTraceStore.getState()

      setEdgeAnimated('edge-1', false)

      const { animatedEdges } = useExecutionTraceStore.getState()
      expect(animatedEdges.has('edge-1')).toBe(false)
      expect(animatedEdges.size).toBe(0)
    })
  })

  describe('Timeline event tracking', () => {
    it('should initialize with empty timeline events', () => {
      const { timelineEvents } = useExecutionTraceStore.getState()
      expect(timelineEvents).toEqual([])
    })

    it('should add timeline event', () => {
      const { addTimelineEvent } = useExecutionTraceStore.getState()

      const event = {
        timestamp: '2025-12-17T10:00:00Z',
        nodeId: 'node-1',
        type: 'started' as const,
        message: 'Node execution started',
      }

      addTimelineEvent(event)

      const { timelineEvents } = useExecutionTraceStore.getState()
      expect(timelineEvents).toHaveLength(1)
      expect(timelineEvents[0]).toEqual(event)
    })

    it('should add multiple timeline events in order', () => {
      const { addTimelineEvent } = useExecutionTraceStore.getState()

      const event1 = {
        timestamp: '2025-12-17T10:00:00Z',
        nodeId: 'node-1',
        type: 'started' as const,
        message: 'Node 1 started',
      }

      const event2 = {
        timestamp: '2025-12-17T10:00:01Z',
        nodeId: 'node-1',
        type: 'completed' as const,
        message: 'Node 1 completed',
      }

      const event3 = {
        timestamp: '2025-12-17T10:00:02Z',
        nodeId: 'node-2',
        type: 'started' as const,
        message: 'Node 2 started',
      }

      addTimelineEvent(event1)
      addTimelineEvent(event2)
      addTimelineEvent(event3)

      const { timelineEvents } = useExecutionTraceStore.getState()
      expect(timelineEvents).toHaveLength(3)
      expect(timelineEvents[0]).toEqual(event1)
      expect(timelineEvents[1]).toEqual(event2)
      expect(timelineEvents[2]).toEqual(event3)
    })

    it('should support different event types', () => {
      const { addTimelineEvent } = useExecutionTraceStore.getState()

      const events = [
        {
          timestamp: '2025-12-17T10:00:00Z',
          nodeId: 'node-1',
          type: 'started' as const,
          message: 'Started',
        },
        {
          timestamp: '2025-12-17T10:00:01Z',
          nodeId: 'node-1',
          type: 'completed' as const,
          message: 'Completed',
        },
        {
          timestamp: '2025-12-17T10:00:02Z',
          nodeId: 'node-2',
          type: 'failed' as const,
          message: 'Failed',
        },
      ]

      events.forEach(addTimelineEvent)

      const { timelineEvents } = useExecutionTraceStore.getState()
      expect(timelineEvents[0].type).toBe('started')
      expect(timelineEvents[1].type).toBe('completed')
      expect(timelineEvents[2].type).toBe('failed')
    })
  })

  describe('Reset functionality', () => {
    it('should reset all state to initial values', () => {
      const store = useExecutionTraceStore.getState()

      // Set up some state
      store.setNodeStatus('node-1', 'running')
      store.setNodeStatus('node-2', 'completed')
      store.addStepLog('node-1', {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'http',
        status: 'completed',
        duration_ms: 100,
      })
      store.setEdgeAnimated('edge-1', true)
      store.addTimelineEvent({
        timestamp: '2025-12-17T10:00:00Z',
        nodeId: 'node-1',
        type: 'started',
        message: 'Started',
      })

      // Reset
      store.reset()

      // Verify everything is reset
      const state = useExecutionTraceStore.getState()
      expect(state.nodeStatuses).toEqual({})
      expect(state.stepLogs).toEqual({})
      expect(state.animatedEdges.size).toBe(0)
      expect(state.timelineEvents).toEqual([])
    })

    it('should allow setting state after reset', () => {
      const store = useExecutionTraceStore.getState()

      // Set, reset, set again
      store.setNodeStatus('node-1', 'running')
      store.reset()
      store.setNodeStatus('node-2', 'completed')

      const { nodeStatuses } = useExecutionTraceStore.getState()
      expect(nodeStatuses['node-1']).toBeUndefined()
      expect(nodeStatuses['node-2']).toBe('completed')
    })
  })

  describe('Current execution tracking', () => {
    it('should initialize with null execution ID', () => {
      const { currentExecutionId } = useExecutionTraceStore.getState()
      expect(currentExecutionId).toBeNull()
    })

    it('should set current execution ID', () => {
      const { setCurrentExecutionId } = useExecutionTraceStore.getState()

      setCurrentExecutionId('exec-123')

      const { currentExecutionId } = useExecutionTraceStore.getState()
      expect(currentExecutionId).toBe('exec-123')
    })

    it('should update execution ID', () => {
      const { setCurrentExecutionId } = useExecutionTraceStore.getState()

      setCurrentExecutionId('exec-123')
      setCurrentExecutionId('exec-456')

      const { currentExecutionId } = useExecutionTraceStore.getState()
      expect(currentExecutionId).toBe('exec-456')
    })

    it('should clear execution ID when set to null', () => {
      const { setCurrentExecutionId } = useExecutionTraceStore.getState()

      setCurrentExecutionId('exec-123')
      setCurrentExecutionId(null)

      const { currentExecutionId } = useExecutionTraceStore.getState()
      expect(currentExecutionId).toBeNull()
    })

    it('should reset execution ID on reset', () => {
      const { setCurrentExecutionId, reset } = useExecutionTraceStore.getState()

      setCurrentExecutionId('exec-123')
      reset()

      const { currentExecutionId } = useExecutionTraceStore.getState()
      expect(currentExecutionId).toBeNull()
    })
  })
})
