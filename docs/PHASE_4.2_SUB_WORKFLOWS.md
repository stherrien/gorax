# Phase 4.2: Sub-Workflows Implementation

## Overview
Implemented sub-workflow functionality allowing workflows to call other workflows as nested executions with parameter passing, recursion protection, and both synchronous and asynchronous execution modes.

## Implementation Summary

### 1. Database Schema (Migration 012)

**File:** `/Users/shawntherrien/Projects/gorax/migrations/012_subworkflow_support.sql`

Added support for tracking sub-workflow execution relationships:
- `parent_execution_id`: References parent execution for nested workflows
- `execution_depth`: Tracks depth of nested execution (0 = root workflow)
- Indexes for efficient parent-child queries

### 2. Backend Model Updates

**File:** `/Users/shawntherrien/Projects/gorax/internal/workflow/model.go`

Added:
- `NodeTypeControlSubWorkflow`: New node type constant
- `SubWorkflowConfig`: Configuration structure with:
  - `WorkflowID`: Target workflow to execute
  - `InputMapping`: Maps parent context to sub-workflow inputs
  - `OutputMapping`: Maps sub-workflow outputs back to parent
  - `WaitForResult`: Synchronous vs asynchronous execution
  - `TimeoutMs`: Timeout for synchronous execution

Updated `Execution` model:
- `ParentExecutionID`: Link to parent execution
- `ExecutionDepth`: Nesting depth for recursion protection

### 3. Execution Context Enhancement

**File:** `/Users/shawntherrien/Projects/gorax/internal/executor/executor.go`

Enhanced `ExecutionContext` with:
- `Depth`: Current execution depth
- `WorkflowChain`: Chain of workflow IDs to detect circular dependencies
- `ParentExecutionID`: Reference to parent execution
- `MaxSubWorkflowDepth` constant: Maximum nesting depth (10 levels)

### 4. Sub-Workflow Action Implementation

**Files:**
- `/Users/shawntherrien/Projects/gorax/internal/executor/actions/subworkflow.go`
- `/Users/shawntherrien/Projects/gorax/internal/executor/actions/subworkflow_test.go`

**Features:**
- **Input/Output Mapping**: Simple expression evaluation (${path.to.value})
- **Synchronous Execution**: Wait for sub-workflow completion with timeout
- **Asynchronous Execution**: Fire-and-forget mode
- **Recursion Protection**:
  - Maximum depth checking (10 levels)
  - Circular dependency detection via workflow chain tracking
- **Tenant Isolation**: Sub-workflows inherit parent tenant context
- **Error Handling**: Comprehensive error reporting with context

**Test Coverage:**
- Input/output mapping with nested paths
- Circular dependency detection
- Maximum depth enforcement
- Synchronous vs asynchronous modes
- Timeout handling
- Tenant isolation

### 5. Executor Integration

**File:** `/Users/shawntherrien/Projects/gorax/internal/executor/subworkflow.go`

Created:
- `executeSubWorkflowAction`: Executor method to handle sub-workflow nodes
- `workflowRepositoryAdapter`: Adapter pattern to bridge repository interfaces
- Integrated into main executor switch statement

### 6. Frontend Component

**Files:**
- `/Users/shawntherrien/Projects/gorax/web/src/components/nodes/SubWorkflowNode.tsx`
- `/Users/shawntherrien/Projects/gorax/web/src/components/nodes/SubWorkflowNode.test.tsx`

**Features:**
- Visual display of sub-workflow configuration
- Async workflow name fetching
- Input/output mapping count display
- Sync/async mode indicator with color coding
- Timeout display
- Comprehensive test coverage (18 test cases)

**Visual Design:**
- Gradient background (indigo to blue)
- Icon: 🔗 (link symbol)
- Displays workflow name, I/O counts, execution mode, timeout
- Selected state highlighting

## API Endpoints

No new API endpoints required. Sub-workflows use existing workflow execution infrastructure.

## Usage Example

```json
{
  "id": "node-123",
  "type": "control:sub_workflow",
  "data": {
    "name": "Send Notification",
    "config": {
      "workflow_id": "wf-notify-456",
      "input_mapping": {
        "user_id": "${trigger.user.id}",
        "message": "${steps.prepare.message}",
        "channel": "${trigger.channel}"
      },
      "output_mapping": {
        "notification_id": "${output.id}",
        "status": "${output.status}"
      },
      "wait_for_result": true,
      "timeout_ms": 5000
    }
  }
}
```

## Recursion Protection

### Maximum Depth
- Hard limit: 10 levels of nesting
- Prevents stack overflow and infinite recursion
- Configurable via `MaxSubWorkflowDepth` constant

### Circular Dependency Detection
- Maintains workflow chain in execution context
- Checks if target workflow already exists in chain
- Returns error before execution if circular reference detected

Example:
```
Workflow A → Workflow B → Workflow C → Workflow A (REJECTED)
```

## Execution Modes

### Synchronous (wait_for_result: true)
- Parent workflow waits for sub-workflow completion
- Returns sub-workflow output via output mapping
- Supports timeout configuration
- Use for: Sequential processing, data dependencies

### Asynchronous (wait_for_result: false)
- Parent workflow continues immediately
- Sub-workflow executes in background
- Returns execution ID and "started" status
- Use for: Fire-and-forget notifications, parallel processing

## Testing

### Backend Tests
All tests passing (9 test cases):
- ✅ Execute success with sync mode
- ✅ Input parameter mapping (simple, multiple, nested)
- ✅ Output parameter mapping
- ✅ Circular dependency detection
- ✅ Maximum depth enforcement
- ✅ Async execution mode
- ✅ Sync timeout handling
- ✅ Missing workflow error
- ✅ Tenant isolation

### Frontend Tests
All tests passing (18 test cases):
- ✅ Basic rendering
- ✅ Workflow name display
- ✅ Async workflow name fetching
- ✅ Error handling (unknown workflow)
- ✅ Input/output mapping counts
- ✅ Sync/async mode display
- ✅ Timeout display
- ✅ Selected state styling
- ✅ Handle rendering
- ✅ Comprehensive configuration

## Security Considerations

1. **Tenant Isolation**: Sub-workflows inherit parent tenant context
2. **Depth Protection**: Prevents resource exhaustion via deep nesting
3. **Circular Protection**: Prevents infinite execution loops
4. **Timeout Enforcement**: Prevents indefinite waiting in sync mode
5. **Workflow Status Validation**: Only active workflows can be executed

## Performance Considerations

1. **Async Mode**: Reduces parent workflow execution time
2. **Timeout Configuration**: Prevents long-running blocks
3. **Indexed Queries**: Parent-child relationship queries are indexed
4. **Depth Tracking**: O(1) depth checking via counter

## Future Enhancements

1. **Enhanced Expression Language**: Support for complex transformations
2. **Conditional Execution**: Execute sub-workflow based on conditions
3. **Batch Sub-Workflow Execution**: Execute multiple workflows in parallel
4. **Sub-Workflow Output Caching**: Cache results for repeated calls
5. **Workflow Schema Validation**: Validate input/output contracts
6. **Visual Workflow Linking**: Click-through to sub-workflow in UI
7. **Execution Tracing**: Parent-child execution visualization
8. **Resource Limits**: Per-tenant limits on sub-workflow depth/count

## Files Modified

### Backend
- `migrations/012_subworkflow_support.sql` (new)
- `internal/workflow/model.go` (modified)
- `internal/executor/executor.go` (modified)
- `internal/executor/subworkflow.go` (new)
- `internal/executor/actions/subworkflow.go` (new)
- `internal/executor/actions/subworkflow_test.go` (new)

### Frontend
- `web/src/components/nodes/SubWorkflowNode.tsx` (new)
- `web/src/components/nodes/SubWorkflowNode.test.tsx` (new)

## Dependencies

No new external dependencies required. Uses existing:
- Backend: Go standard library, existing workflow/executor packages
- Frontend: React, @xyflow/react (already in use)

## Next Steps

1. Run database migration: `012_subworkflow_support.sql`
2. Register `SubWorkflowNode` in frontend node types registry
3. Add UI for configuring sub-workflow nodes in workflow editor
4. Update documentation with sub-workflow examples
5. Consider implementing workflow schema validation for type safety
