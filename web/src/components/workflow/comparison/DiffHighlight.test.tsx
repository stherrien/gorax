import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { WorkflowDefinition } from '../../../types/workflow'
import type { DiffSummary, NodeDiff, PropertyChange } from '../../../types/diff'
import {
  computeWorkflowDiff,
  DiffBadge,
  DiffContainer,
  DiffSummaryDisplay,
  PropertyChangeDisplay,
} from './DiffHighlight'

describe('computeWorkflowDiff', () => {
  const createNode = (id: string, label: string) => ({
    id,
    type: 'action',
    position: { x: 0, y: 0 },
    data: { label },
  })

  const createEdge = (id: string, source: string, target: string) => ({
    id,
    source,
    target,
  })

  it('should detect no changes for identical definitions', () => {
    const definition: WorkflowDefinition = {
      nodes: [createNode('n1', 'Node 1')],
      edges: [createEdge('e1', 'n1', 'n2')],
    }

    const diff = computeWorkflowDiff(definition, definition, 1, 2)

    expect(diff.summary.totalChanges).toBe(0)
    expect(diff.summary.nodesUnchanged).toBe(1)
    expect(diff.summary.edgesUnchanged).toBe(1)
  })

  it('should detect added nodes', () => {
    const base: WorkflowDefinition = {
      nodes: [createNode('n1', 'Node 1')],
      edges: [],
    }
    const compare: WorkflowDefinition = {
      nodes: [createNode('n1', 'Node 1'), createNode('n2', 'Node 2')],
      edges: [],
    }

    const diff = computeWorkflowDiff(base, compare, 1, 2)

    expect(diff.summary.nodesAdded).toBe(1)
    expect(diff.nodeDiffs.find((d) => d.nodeId === 'n2')?.status).toBe('added')
  })

  it('should detect removed nodes', () => {
    const base: WorkflowDefinition = {
      nodes: [createNode('n1', 'Node 1'), createNode('n2', 'Node 2')],
      edges: [],
    }
    const compare: WorkflowDefinition = {
      nodes: [createNode('n1', 'Node 1')],
      edges: [],
    }

    const diff = computeWorkflowDiff(base, compare, 1, 2)

    expect(diff.summary.nodesRemoved).toBe(1)
    expect(diff.nodeDiffs.find((d) => d.nodeId === 'n2')?.status).toBe('removed')
  })

  it('should detect modified nodes', () => {
    const base: WorkflowDefinition = {
      nodes: [createNode('n1', 'Node 1')],
      edges: [],
    }
    const compare: WorkflowDefinition = {
      nodes: [createNode('n1', 'Modified Node 1')],
      edges: [],
    }

    const diff = computeWorkflowDiff(base, compare, 1, 2)

    expect(diff.summary.nodesModified).toBe(1)
    const nodeDiff = diff.nodeDiffs.find((d) => d.nodeId === 'n1')
    expect(nodeDiff?.status).toBe('modified')
    expect(nodeDiff?.propertyChanges).toBeDefined()
    expect(nodeDiff?.propertyChanges?.length).toBeGreaterThan(0)
  })

  it('should detect added edges', () => {
    const base: WorkflowDefinition = {
      nodes: [],
      edges: [createEdge('e1', 'n1', 'n2')],
    }
    const compare: WorkflowDefinition = {
      nodes: [],
      edges: [createEdge('e1', 'n1', 'n2'), createEdge('e2', 'n2', 'n3')],
    }

    const diff = computeWorkflowDiff(base, compare, 1, 2)

    expect(diff.summary.edgesAdded).toBe(1)
    expect(diff.edgeDiffs.find((d) => d.edgeId === 'e2')?.status).toBe('added')
  })

  it('should detect removed edges', () => {
    const base: WorkflowDefinition = {
      nodes: [],
      edges: [createEdge('e1', 'n1', 'n2'), createEdge('e2', 'n2', 'n3')],
    }
    const compare: WorkflowDefinition = {
      nodes: [],
      edges: [createEdge('e1', 'n1', 'n2')],
    }

    const diff = computeWorkflowDiff(base, compare, 1, 2)

    expect(diff.summary.edgesRemoved).toBe(1)
    expect(diff.edgeDiffs.find((d) => d.edgeId === 'e2')?.status).toBe('removed')
  })

  it('should detect settings changes', () => {
    const base: WorkflowDefinition = {
      nodes: [],
      edges: [],
      settings: { timeout: 1000 },
    }
    const compare: WorkflowDefinition = {
      nodes: [],
      edges: [],
      settings: { timeout: 2000 },
    }

    const diff = computeWorkflowDiff(base, compare, 1, 2)

    expect(diff.settingsChanged).toBe(true)
  })

  it('should detect variables changes', () => {
    const base: WorkflowDefinition = {
      nodes: [],
      edges: [],
      variables: [{ name: 'var1', type: 'string' }],
    }
    const compare: WorkflowDefinition = {
      nodes: [],
      edges: [],
      variables: [{ name: 'var1', type: 'number' }],
    }

    const diff = computeWorkflowDiff(base, compare, 1, 2)

    expect(diff.variablesChanged).toBe(true)
  })

  it('should handle empty definitions', () => {
    const empty: WorkflowDefinition = {
      nodes: [],
      edges: [],
    }

    const diff = computeWorkflowDiff(empty, empty, 1, 2)

    expect(diff.summary.totalChanges).toBe(0)
    expect(diff.nodeDiffs).toHaveLength(0)
    expect(diff.edgeDiffs).toHaveLength(0)
  })

  it('should include version numbers in diff result', () => {
    const definition: WorkflowDefinition = {
      nodes: [],
      edges: [],
    }

    const diff = computeWorkflowDiff(definition, definition, 5, 10)

    expect(diff.baseVersion).toBe(5)
    expect(diff.compareVersion).toBe(10)
  })

  it('should calculate correct total changes', () => {
    const base: WorkflowDefinition = {
      nodes: [createNode('n1', 'Node 1'), createNode('n2', 'Node 2')],
      edges: [createEdge('e1', 'n1', 'n2')],
    }
    const compare: WorkflowDefinition = {
      nodes: [createNode('n1', 'Modified'), createNode('n3', 'New')],
      edges: [createEdge('e1', 'n1', 'n2'), createEdge('e2', 'n1', 'n3')],
    }

    const diff = computeWorkflowDiff(base, compare, 1, 2)

    // n1 modified, n2 removed, n3 added = 3 node changes
    // e1 unchanged, e2 added = 1 edge change
    expect(diff.summary.nodesModified).toBe(1)
    expect(diff.summary.nodesRemoved).toBe(1)
    expect(diff.summary.nodesAdded).toBe(1)
    expect(diff.summary.edgesAdded).toBe(1)
    expect(diff.summary.totalChanges).toBe(4)
  })
})

describe('DiffBadge', () => {
  it('should render Added badge', () => {
    render(<DiffBadge status="added" />)
    expect(screen.getByText('Added')).toBeInTheDocument()
  })

  it('should render Removed badge', () => {
    render(<DiffBadge status="removed" />)
    expect(screen.getByText('Removed')).toBeInTheDocument()
  })

  it('should render Modified badge', () => {
    render(<DiffBadge status="modified" />)
    expect(screen.getByText('Modified')).toBeInTheDocument()
  })

  it('should render Unchanged badge', () => {
    render(<DiffBadge status="unchanged" />)
    expect(screen.getByText('Unchanged')).toBeInTheDocument()
  })

  it('should apply custom className', () => {
    render(<DiffBadge status="added" className="custom-class" />)
    const badge = screen.getByText('Added')
    expect(badge).toHaveClass('custom-class')
  })
})

describe('DiffContainer', () => {
  it('should render children', () => {
    render(
      <DiffContainer status="added">
        <span>Test content</span>
      </DiffContainer>
    )
    expect(screen.getByText('Test content')).toBeInTheDocument()
  })

  it('should apply status-based styling', () => {
    const { container } = render(
      <DiffContainer status="added">Content</DiffContainer>
    )
    const containerEl = container.firstChild
    expect(containerEl).toHaveClass('bg-green-900/30')
    expect(containerEl).toHaveClass('border-green-500')
  })

  it('should apply custom className', () => {
    const { container } = render(
      <DiffContainer status="added" className="custom-class">
        Content
      </DiffContainer>
    )
    expect(container.firstChild).toHaveClass('custom-class')
  })
})

describe('DiffSummaryDisplay', () => {
  it('should show nodes added count', () => {
    const summary: DiffSummary = {
      nodesAdded: 3,
      nodesRemoved: 0,
      nodesModified: 0,
      nodesUnchanged: 0,
      edgesAdded: 0,
      edgesRemoved: 0,
      edgesModified: 0,
      edgesUnchanged: 0,
      totalChanges: 3,
    }

    render(<DiffSummaryDisplay summary={summary} />)
    expect(screen.getByText('3 node(s) added')).toBeInTheDocument()
  })

  it('should show nodes removed count', () => {
    const summary: DiffSummary = {
      nodesAdded: 0,
      nodesRemoved: 2,
      nodesModified: 0,
      nodesUnchanged: 0,
      edgesAdded: 0,
      edgesRemoved: 0,
      edgesModified: 0,
      edgesUnchanged: 0,
      totalChanges: 2,
    }

    render(<DiffSummaryDisplay summary={summary} />)
    expect(screen.getByText('2 node(s) removed')).toBeInTheDocument()
  })

  it('should show nodes modified count', () => {
    const summary: DiffSummary = {
      nodesAdded: 0,
      nodesRemoved: 0,
      nodesModified: 5,
      nodesUnchanged: 0,
      edgesAdded: 0,
      edgesRemoved: 0,
      edgesModified: 0,
      edgesUnchanged: 0,
      totalChanges: 5,
    }

    render(<DiffSummaryDisplay summary={summary} />)
    expect(screen.getByText('5 node(s) modified')).toBeInTheDocument()
  })

  it('should show edges added and removed', () => {
    const summary: DiffSummary = {
      nodesAdded: 0,
      nodesRemoved: 0,
      nodesModified: 0,
      nodesUnchanged: 0,
      edgesAdded: 2,
      edgesRemoved: 1,
      edgesModified: 0,
      edgesUnchanged: 0,
      totalChanges: 3,
    }

    render(<DiffSummaryDisplay summary={summary} />)
    expect(screen.getByText('2 connection(s) added')).toBeInTheDocument()
    expect(screen.getByText('1 connection(s) removed')).toBeInTheDocument()
  })

  it('should show no changes message when totalChanges is 0', () => {
    const summary: DiffSummary = {
      nodesAdded: 0,
      nodesRemoved: 0,
      nodesModified: 0,
      nodesUnchanged: 5,
      edgesAdded: 0,
      edgesRemoved: 0,
      edgesModified: 0,
      edgesUnchanged: 3,
      totalChanges: 0,
    }

    render(<DiffSummaryDisplay summary={summary} />)
    expect(screen.getByText('No changes detected')).toBeInTheDocument()
  })

  it('should not show zero counts', () => {
    const summary: DiffSummary = {
      nodesAdded: 1,
      nodesRemoved: 0,
      nodesModified: 0,
      nodesUnchanged: 0,
      edgesAdded: 0,
      edgesRemoved: 0,
      edgesModified: 0,
      edgesUnchanged: 0,
      totalChanges: 1,
    }

    render(<DiffSummaryDisplay summary={summary} />)
    expect(screen.getByText('1 node(s) added')).toBeInTheDocument()
    expect(screen.queryByText(/removed/)).not.toBeInTheDocument()
    expect(screen.queryByText(/modified/)).not.toBeInTheDocument()
  })
})

describe('PropertyChangeDisplay', () => {
  it('should render property changes', () => {
    const changes: PropertyChange[] = [
      { path: 'data.label', baseValue: 'Old', compareValue: 'New', type: 'modified' },
    ]

    render(<PropertyChangeDisplay changes={changes} />)
    expect(screen.getByText('Property Changes')).toBeInTheDocument()
    expect(screen.getByText('data.label:')).toBeInTheDocument()
  })

  it('should show added property correctly', () => {
    const changes: PropertyChange[] = [
      { path: 'data.newProp', baseValue: undefined, compareValue: 'value', type: 'added' },
    ]

    render(<PropertyChangeDisplay changes={changes} />)
    expect(screen.getByText(/\+ "value"/)).toBeInTheDocument()
  })

  it('should show removed property correctly', () => {
    const changes: PropertyChange[] = [
      { path: 'data.oldProp', baseValue: 'value', compareValue: undefined, type: 'removed' },
    ]

    render(<PropertyChangeDisplay changes={changes} />)
    expect(screen.getByText(/- "value"/)).toBeInTheDocument()
  })

  it('should show modified property with old and new values', () => {
    const changes: PropertyChange[] = [
      { path: 'data.prop', baseValue: 'old', compareValue: 'new', type: 'modified' },
    ]

    render(<PropertyChangeDisplay changes={changes} />)
    expect(screen.getByText(/"old"/)).toBeInTheDocument()
    expect(screen.getByText('→')).toBeInTheDocument()
    expect(screen.getByText(/"new"/)).toBeInTheDocument()
  })

  it('should return null for empty changes', () => {
    const { container } = render(<PropertyChangeDisplay changes={[]} />)
    expect(container.firstChild).toBeNull()
  })

  it('should format different value types correctly', () => {
    const changes: PropertyChange[] = [
      { path: 'number', baseValue: 42, compareValue: 100, type: 'modified' },
      { path: 'boolean', baseValue: true, compareValue: false, type: 'modified' },
      { path: 'null', baseValue: null, compareValue: 'value', type: 'modified' },
    ]

    render(<PropertyChangeDisplay changes={changes} />)
    expect(screen.getByText('42')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
    expect(screen.getByText('true')).toBeInTheDocument()
    expect(screen.getByText('false')).toBeInTheDocument()
    expect(screen.getByText('null')).toBeInTheDocument()
  })
})
