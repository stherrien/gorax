import { Node, Edge } from '@xyflow/react'

/**
 * Detects cycles in a directed graph using DFS
 * @param nodes - Array of graph nodes
 * @param edges - Array of graph edges
 * @returns Array of cycle paths (each path is an array of node IDs)
 */
export function detectCycles(nodes: Node[], edges: Edge[]): string[][] {
  if (nodes.length === 0) {
    return []
  }

  // Build adjacency list
  const adjacencyList = buildAdjacencyList(nodes, edges)

  // Track visited nodes and nodes in current path
  const visited = new Set<string>()
  const inPath = new Set<string>()
  const cycles: string[][] = []
  const currentPath: string[] = []

  function dfs(nodeId: string): void {
    visited.add(nodeId)
    inPath.add(nodeId)
    currentPath.push(nodeId)

    const neighbors = adjacencyList.get(nodeId) || []
    for (const neighbor of neighbors) {
      if (!visited.has(neighbor)) {
        dfs(neighbor)
      } else if (inPath.has(neighbor)) {
        // Found a cycle - extract the cycle path
        const cycleStartIndex = currentPath.indexOf(neighbor)
        const cyclePath = [...currentPath.slice(cycleStartIndex), neighbor]
        cycles.push(cyclePath)
      }
    }

    currentPath.pop()
    inPath.delete(nodeId)
  }

  // Visit all nodes to handle disconnected components
  for (const node of nodes) {
    if (!visited.has(node.id)) {
      dfs(node.id)
    }
  }

  return cycles
}

/**
 * Checks if the graph is a valid DAG (no cycles)
 * @param nodes - Array of graph nodes
 * @param edges - Array of graph edges
 * @returns true if graph is a valid DAG, false otherwise
 */
export function isValidDAG(nodes: Node[], edges: Edge[]): boolean {
  const cycles = detectCycles(nodes, edges)
  return cycles.length === 0
}

/**
 * Result of topological sort operation
 */
export interface TopologicalOrderResult {
  success: boolean
  order: string[]
  error?: string
}

/**
 * Performs topological sort using Kahn's algorithm
 * @param nodes - Array of graph nodes
 * @param edges - Array of graph edges
 * @returns Result object with success status and ordered node IDs or error
 */
export function getTopologicalOrder(nodes: Node[], edges: Edge[]): TopologicalOrderResult {
  if (nodes.length === 0) {
    return { success: true, order: [] }
  }

  // Build adjacency list and in-degree map
  const adjacencyList = buildAdjacencyList(nodes, edges)
  const inDegree = buildInDegreeMap(nodes, edges)

  // Find all nodes with in-degree 0
  const queue: string[] = []
  for (const node of nodes) {
    if (inDegree.get(node.id) === 0) {
      queue.push(node.id)
    }
  }

  const order: string[] = []

  while (queue.length > 0) {
    const nodeId = queue.shift()!
    order.push(nodeId)

    const neighbors = adjacencyList.get(nodeId) || []
    for (const neighbor of neighbors) {
      const newInDegree = (inDegree.get(neighbor) || 0) - 1
      inDegree.set(neighbor, newInDegree)

      if (newInDegree === 0) {
        queue.push(neighbor)
      }
    }
  }

  // If we couldn't process all nodes, there's a cycle
  if (order.length !== nodes.length) {
    const cycles = detectCycles(nodes, edges)
    const cycleDescription = cycles.length > 0
      ? `Cycle detected: ${cycles[0].join(' -> ')}`
      : 'Graph contains cycles'

    return {
      success: false,
      order: [],
      error: cycleDescription,
    }
  }

  return { success: true, order }
}

/**
 * Builds an adjacency list representation of the graph
 * Time complexity: O(V + E)
 */
function buildAdjacencyList(nodes: Node[], edges: Edge[]): Map<string, string[]> {
  const adjacencyList = new Map<string, string[]>()

  // Initialize with empty arrays for all nodes
  for (const node of nodes) {
    adjacencyList.set(node.id, [])
  }

  // Add edges
  for (const edge of edges) {
    const neighbors = adjacencyList.get(edge.source) || []
    neighbors.push(edge.target)
    adjacencyList.set(edge.source, neighbors)
  }

  return adjacencyList
}

/**
 * Builds an in-degree map for all nodes
 * Time complexity: O(V + E)
 */
function buildInDegreeMap(nodes: Node[], edges: Edge[]): Map<string, number> {
  const inDegree = new Map<string, number>()

  // Initialize with 0 for all nodes
  for (const node of nodes) {
    inDegree.set(node.id, 0)
  }

  // Count incoming edges
  for (const edge of edges) {
    inDegree.set(edge.target, (inDegree.get(edge.target) || 0) + 1)
  }

  return inDegree
}
