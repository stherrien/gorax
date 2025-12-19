# Execution Trace UI Components

This module provides real-time execution visualization components for the gorax workflow engine.

## Components

### ExecutionTimeline
Displays a chronological list of execution events with visual indicators and timestamps.

**Features:**
- Event type icons (started, completed, failed, progress)
- Color-coded status indicators
- Auto-scrolls to latest events
- Displays metadata when available
- Fully accessible with ARIA labels

### StepLogViewer
Shows detailed logs for a selected workflow node.

**Features:**
- Displays input/output data with JSON formatting
- Shows error messages with highlighting
- Timestamps and duration tracking
- Copy to clipboard functionality
- Supports multiple execution steps per node

### ExecutionDetailsPanel
Combines timeline and log viewer in a tabbed interface.

**Features:**
- Tabbed navigation (Timeline / Logs)
- Sticky header with execution status
- Keyboard navigation support
- Passes selected node to log viewer

### ExecutionCanvas
Full execution visualization with workflow canvas and real-time updates.

**Features:**
- Read-only workflow canvas with ReactFlow
- Integrated ExecutionDetailsPanel (Timeline + LogViewer)
- WebSocket connection status indicator
- Node selection for detailed logs
- Split-view layout (canvas left, details right)
- Auto-connects to execution updates
- Resets state on unmount

## Hooks

### useExecutionTrace
Connects to WebSocket for real-time execution updates and updates the execution trace store.

**Features:**
- Auto-connects based on execution ID
- Updates store with node status, logs, and timeline events
- Handles all execution event types
- Connection state management

## Usage Example

### Using ExecutionCanvas (Recommended)

The simplest way to visualize executions with full canvas integration:

```tsx
import { ExecutionCanvas } from './components/execution'

function ExecutionView({ workflowId, executionId }: { workflowId: string; executionId: string }) {
  return (
    <div className="h-screen">
      <ExecutionCanvas
        workflowId={workflowId}
        executionId={executionId}
      />
    </div>
  )
}
```

### Using ExecutionDetailsPanel with Custom Canvas

For more control over the canvas:

```tsx
import { useState } from 'react'
import { ExecutionDetailsPanel } from './components/execution'
import { useExecutionTrace } from './hooks/useExecutionTrace'

function WorkflowExecutionView({ executionId }: { executionId: string }) {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)

  // Connect to execution updates
  const { connected, reconnecting } = useExecutionTrace(executionId)

  return (
    <div className="execution-view">
      {/* Connection indicator */}
      {reconnecting && <div className="reconnecting">Reconnecting...</div>}
      {!connected && <div className="disconnected">Disconnected</div>}

      {/* Main workflow canvas */}
      <div className="workflow-canvas">
        {/* Your workflow visualization with node selection */}
        <WorkflowCanvas onNodeSelect={setSelectedNodeId} />
      </div>

      {/* Execution details side panel */}
      <ExecutionDetailsPanel selectedNodeId={selectedNodeId} />
    </div>
  )
}
```

### Using Individual Components

```tsx
import { ExecutionTimeline, StepLogViewer } from './components/execution'
import { useExecutionTrace } from './hooks/useExecutionTrace'

function CustomExecutionView({ executionId }: { executionId: string }) {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)

  useExecutionTrace(executionId)

  return (
    <div className="custom-layout">
      {/* Timeline in sidebar */}
      <aside className="sidebar">
        <ExecutionTimeline />
      </aside>

      {/* Logs in main area */}
      <main className="main-content">
        <StepLogViewer selectedNodeId={selectedNodeId} />
      </main>
    </div>
  )
}
```

## Store Integration

The components use the `executionTraceStore` (Zustand) for state management. The store tracks:

- **Current execution ID**: Active execution being traced
- **Node statuses**: Status for each node (pending, running, completed, failed)
- **Step logs**: Detailed execution logs per node
- **Timeline events**: Chronological list of all execution events
- **Animated edges**: Set of edge IDs to animate

The `useExecutionTrace` hook automatically updates the store based on WebSocket events.

## Styling

All components use the `executionTrace.css` stylesheet which includes:

- Timeline vertical line and event dots
- Status color coding (blue=started, green=completed, red=failed, yellow=progress)
- Tab navigation styles
- Dark mode support
- Responsive design
- Accessibility enhancements
- Respects `prefers-reduced-motion`

## Accessibility

All components are fully accessible:

- Proper ARIA labels and roles
- Keyboard navigation support
- Screen reader announcements
- Focus indicators
- High contrast mode support

## Testing

Each component has comprehensive test coverage:

- **ExecutionTimeline**: 20 tests
- **StepLogViewer**: 17 tests
- **ExecutionDetailsPanel**: 17 tests
- **ExecutionCanvas**: 23 tests
- **useExecutionTrace**: 15 tests

**Total: 92 tests** covering all functionality, accessibility, and edge cases.

Run tests with:
```bash
npm test ExecutionTimeline.test.tsx
npm test StepLogViewer.test.tsx
npm test ExecutionDetailsPanel.test.tsx
npm test ExecutionCanvas.test.tsx
npm test useExecutionTrace.test.ts
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    ExecutionCanvas                          │
│  (Workflow Canvas + ExecutionDetailsPanel + WebSocket)     │
└─────────────────┬──────────────────────────┬────────────────┘
                  │                          │
                  ↓                          ↓
      ┌───────────────────┐      ┌──────────────────────┐
      │   WorkflowCanvas  │      │ ExecutionDetailsPanel│
      │   (ReactFlow)     │      │    (Tabbed UI)       │
      └───────────────────┘      └──────────────────────┘
                                            │
                                    ┌───────┴────────┐
                                    ↓                ↓
                            ┌──────────────┐  ┌─────────────────┐
                            │ExecutionTime │  │  StepLogViewer  │
                            │    line      │  │                 │
                            └──────────────┘  └─────────────────┘

                  All components read from:
                           ↓
      ┌─────────────────────────────────────────┐
      │      executionTraceStore (Zustand)      │
      │   (Centralized execution state)         │
      └───────────────┬─────────────────────────┘
                      ↑
                      │ Updates
      ┌───────────────┴─────────────────────────┐
      │         useExecutionTrace Hook           │
      │  (Connects to WebSocket, updates store) │
      └─────────────────────────────────────────┘
```

## Performance Considerations

- **Auto-scroll**: Uses `scrollIntoView` with smooth behavior (disabled if `prefers-reduced-motion`)
- **Timeline**: Efficiently renders large event lists with proper key management
- **Store updates**: Uses Zustand for optimal re-renders
- **WebSocket**: Automatic reconnection with exponential backoff

## Browser Support

- Modern browsers with WebSocket support
- Graceful degradation for older browsers
- Mobile responsive
