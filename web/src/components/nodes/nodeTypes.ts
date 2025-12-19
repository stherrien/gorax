/**
 * Node type registration for ReactFlow
 * Wraps all node components with ExecutionStatusNode for real-time execution visualization
 */

import { ExecutionStatusNode } from './ExecutionStatusNode'
import TriggerNode from './TriggerNode'
import ActionNode from './ActionNode'
import ConditionalNode from './ConditionalNode'
import LoopNode from './LoopNode'

/**
 * Node types registry for ReactFlow
 * All nodes are wrapped with ExecutionStatusNode to support execution visualization
 */
export const nodeTypes = {
  trigger: ExecutionStatusNode(TriggerNode),
  action: ExecutionStatusNode(ActionNode),
  conditional: ExecutionStatusNode(ConditionalNode),
  loop: ExecutionStatusNode(LoopNode),
} as const

/**
 * Type-safe node type keys
 */
export type NodeTypeKey = keyof typeof nodeTypes
