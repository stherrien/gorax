import { create } from 'zustand'
import type { StepInfo } from '../lib/websocket'

// Node execution status
export type NodeStatus = 'pending' | 'running' | 'completed' | 'failed'

// Timeline event types
export type TimelineEventType = 'started' | 'completed' | 'failed' | 'progress'

// Timeline event structure
export interface TimelineEvent {
  timestamp: string
  nodeId: string
  type: TimelineEventType
  message: string
  metadata?: Record<string, unknown>
}

// Store state interface
interface ExecutionTraceState {
  // Current execution being traced
  currentExecutionId: string | null

  // Node status tracking: nodeId -> status
  nodeStatuses: Record<string, NodeStatus>

  // Step logs: nodeId -> array of step info
  stepLogs: Record<string, StepInfo[]>

  // Animated edges (Set for O(1) lookup)
  animatedEdges: Set<string>

  // Timeline of execution events
  timelineEvents: TimelineEvent[]

  // Actions
  setCurrentExecutionId: (executionId: string | null) => void
  setNodeStatus: (nodeId: string, status: NodeStatus) => void
  addStepLog: (nodeId: string, stepInfo: StepInfo) => void
  setEdgeAnimated: (edgeId: string, animated: boolean) => void
  addTimelineEvent: (event: TimelineEvent) => void
  reset: () => void
}

// Initial state
const initialState = {
  currentExecutionId: null,
  nodeStatuses: {},
  stepLogs: {},
  animatedEdges: new Set<string>(),
  timelineEvents: [],
}

/**
 * Execution trace store - tracks real-time execution state
 * for visual feedback in the workflow canvas
 */
export const useExecutionTraceStore = create<ExecutionTraceState>((set) => ({
  ...initialState,

  setCurrentExecutionId: (executionId) =>
    set({ currentExecutionId: executionId }),

  setNodeStatus: (nodeId, status) =>
    set((state) => ({
      nodeStatuses: {
        ...state.nodeStatuses,
        [nodeId]: status,
      },
    })),

  addStepLog: (nodeId, stepInfo) =>
    set((state) => ({
      stepLogs: {
        ...state.stepLogs,
        [nodeId]: [...(state.stepLogs[nodeId] || []), stepInfo],
      },
    })),

  setEdgeAnimated: (edgeId, animated) =>
    set((state) => {
      const newAnimatedEdges = new Set(state.animatedEdges)
      if (animated) {
        newAnimatedEdges.add(edgeId)
      } else {
        newAnimatedEdges.delete(edgeId)
      }
      return { animatedEdges: newAnimatedEdges }
    }),

  addTimelineEvent: (event) =>
    set((state) => ({
      timelineEvents: [...state.timelineEvents, event],
    })),

  reset: () =>
    set({
      currentExecutionId: null,
      nodeStatuses: {},
      stepLogs: {},
      animatedEdges: new Set<string>(),
      timelineEvents: [],
    }),
}))
