/**
 * Node type registration for ReactFlow v12
 * Wraps all node components with ExecutionStatusNode for real-time execution visualization
 */

import { ComponentType } from 'react'
import { ExecutionStatusNode } from './ExecutionStatusNode'
import TriggerNode from './TriggerNode'
import ActionNode from './ActionNode'
import AINode from './AINode'
import ConditionalNode from './ConditionalNode'
import LoopNode from './LoopNode'
import ParallelNode from './ParallelNode'
import ForkNode from './ForkNode'
import JoinNode from './JoinNode'

/**
 * Node types registry for ReactFlow v12
 * All nodes are wrapped with ExecutionStatusNode to support execution visualization
 *
 * Note: We use a custom type here because ReactFlow v12's NodeTypes has strict
 * type requirements that our node components satisfy at runtime but TypeScript
 * can't verify due to the HOC pattern.
 */
export const nodeTypes: Record<string, ComponentType<any>> = {
  trigger: ExecutionStatusNode(TriggerNode),
  action: ExecutionStatusNode(ActionNode),
  ai: ExecutionStatusNode(AINode),
  conditional: ExecutionStatusNode(ConditionalNode),
  loop: ExecutionStatusNode(LoopNode),
  parallel: ExecutionStatusNode(ParallelNode),
  fork: ExecutionStatusNode(ForkNode),
  join: ExecutionStatusNode(JoinNode),
}

/**
 * Type-safe node type keys
 */
export type NodeTypeKey = 'trigger' | 'action' | 'ai' | 'conditional' | 'loop' | 'parallel' | 'fork' | 'join'
