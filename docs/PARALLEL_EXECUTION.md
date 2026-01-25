# Parallel Execution in Gorax Workflows

## Overview

Parallel execution allows workflows to execute multiple actions concurrently, significantly improving performance when tasks are independent and can run simultaneously.

## Key Concepts

### Named Branches

Parallel execution is organized into **named branches**. Each branch:
- Has a unique name for identification
- Contains one or more nodes that execute sequentially within that branch
- Executes concurrently with other branches

### Wait Modes

Control how the parallel node waits for branch completion:

#### `all` (Default)
- Waits for **all branches** to complete before continuing
- Returns results from all branches
- Use when you need outputs from every branch

**Example Use Cases:**
- Send notifications via email, SMS, and Slack simultaneously
- Query multiple data sources and aggregate results
- Process multiple files in parallel

#### `first`
- Returns as soon as the **first branch** completes
- Cancels remaining branches
- Use for race conditions or when first result is sufficient

**Example Use Cases:**
- Query multiple mirror servers, use whichever responds first
- Try multiple AI providers simultaneously, use fastest response
- Health checks across multiple endpoints

### Failure Modes

Control how errors are handled:

#### `stop_all` (Default)
- Stops all branches immediately when any branch fails
- Returns error from the first failing branch
- Use when all branches must succeed

#### `continue`
- Continues executing all branches even if some fail
- Collects results from successful branches
- Returns partial results
- Use for best-effort scenarios

**Example Use Cases:**
- Send notifications across channels (continue even if one fails)
- Process multiple records (collect successes, log failures)
- Non-critical parallel operations

### Concurrency Limiting

Control the maximum number of branches that execute simultaneously:

```json
{
  "max_concurrency": 0  // unlimited (default)
  "max_concurrency": 5  // max 5 branches at once
}
```

**When to Use:**
- Respect API rate limits
- Control resource consumption (memory, connections)
- Prevent overwhelming downstream services
- Batch processing with controlled parallelism

### Timeout

Set a maximum duration for parallel execution:

```json
{
  "timeout": "30s"   // 30 seconds
  "timeout": "2m"    // 2 minutes
  "timeout": "1h"    // 1 hour
}
```

## Configuration

### Complete Configuration Example

```json
{
  "type": "control:parallel",
  "config": {
    "branches": [
      {
        "name": "email_notification",
        "nodes": ["compose_email", "send_email"]
      },
      {
        "name": "sms_notification",
        "nodes": ["send_sms"]
      },
      {
        "name": "slack_notification",
        "nodes": ["format_message", "post_to_slack"]
      }
    ],
    "wait_mode": "all",
    "max_concurrency": 0,
    "timeout": "30s",
    "failure_mode": "continue"
  }
}
```

### Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `branches` | Array | **Required** | Named branches to execute in parallel |
| `wait_mode` | String | `"all"` | Wait for `"all"` or `"first"` branch |
| `max_concurrency` | Integer | `0` | Max concurrent branches (0 = unlimited) |
| `timeout` | String | `""` | Max execution time (e.g., "30s", "1m") |
| `failure_mode` | String | `"stop_all"` | Error handling: `"stop_all"` or `"continue"` |

### Branch Definition

```json
{
  "name": "branch_name",      // Unique identifier
  "nodes": ["node1", "node2"] // Nodes to execute (sequential within branch)
}
```

## Common Use Cases

### 1. Multi-Channel Notifications

Send the same message across multiple channels simultaneously:

```json
{
  "type": "control:parallel",
  "config": {
    "branches": [
      {"name": "email", "nodes": ["send_email"]},
      {"name": "sms", "nodes": ["send_sms"]},
      {"name": "slack", "nodes": ["post_slack"]},
      {"name": "webhook", "nodes": ["call_webhook"]}
    ],
    "wait_mode": "all",
    "failure_mode": "continue",  // Continue even if one channel fails
    "timeout": "30s"
  }
}
```

### 2. Concurrent API Queries

Query multiple APIs and aggregate results:

```json
{
  "type": "control:parallel",
  "config": {
    "branches": [
      {"name": "weather_api", "nodes": ["fetch_weather"]},
      {"name": "traffic_api", "nodes": ["fetch_traffic"]},
      {"name": "news_api", "nodes": ["fetch_news"]}
    ],
    "wait_mode": "all",
    "failure_mode": "stop_all",
    "timeout": "5s"
  }
}
```

### 3. Race Condition (First Wins)

Use the fastest response from multiple sources:

```json
{
  "type": "control:parallel",
  "config": {
    "branches": [
      {"name": "primary_server", "nodes": ["query_primary"]},
      {"name": "backup_server", "nodes": ["query_backup"]},
      {"name": "cache_server", "nodes": ["query_cache"]}
    ],
    "wait_mode": "first",      // Use whichever responds first
    "failure_mode": "continue",
    "timeout": "10s"
  }
}
```

### 4. Rate-Limited Batch Processing

Process many items with controlled concurrency:

```json
{
  "type": "control:parallel",
  "config": {
    "branches": [
      {"name": "batch_1", "nodes": ["process_batch_1"]},
      {"name": "batch_2", "nodes": ["process_batch_2"]},
      {"name": "batch_3", "nodes": ["process_batch_3"]},
      {"name": "batch_4", "nodes": ["process_batch_4"]},
      {"name": "batch_5", "nodes": ["process_batch_5"]}
    ],
    "wait_mode": "all",
    "max_concurrency": 2,  // Only 2 batches at a time
    "failure_mode": "continue",
    "timeout": "5m"
  }
}
```

### 5. Fan-Out / Fan-In Pattern

Process data in parallel, then aggregate:

```workflow
[Trigger]
    ↓
[Split Data]
    ↓
[Parallel Processing] → {branch1, branch2, branch3}
    ↓
[Aggregate Results]
    ↓
[Continue Workflow]
```

### 6. Parallel Data Enrichment

Enrich a record by querying multiple services:

```json
{
  "type": "control:parallel",
  "config": {
    "branches": [
      {"name": "user_profile", "nodes": ["fetch_user"]},
      {"name": "order_history", "nodes": ["fetch_orders"]},
      {"name": "preferences", "nodes": ["fetch_preferences"]},
      {"name": "analytics", "nodes": ["fetch_analytics"]}
    ],
    "wait_mode": "all",
    "failure_mode": "continue",  // Partial enrichment is OK
    "timeout": "3s"
  }
}
```

## Performance Considerations

### When to Use Parallel Execution

✅ **Good Use Cases:**
- Independent operations (no dependencies between branches)
- I/O-bound operations (API calls, database queries, file operations)
- Network requests to external services
- Multiple notifications or updates
- Data fetching from multiple sources

❌ **Avoid For:**
- CPU-intensive operations (limited by CPU cores)
- Operations with dependencies (use sequential execution)
- Single fast operation (overhead not worth it)
- Operations that share state or resources

### Performance Best Practices

1. **Set Appropriate Concurrency Limits**
   ```json
   "max_concurrency": 10  // Don't overwhelm downstream services
   ```

2. **Always Set Timeouts**
   ```json
   "timeout": "30s"  // Prevent hanging workflows
   ```

3. **Use `continue` for Non-Critical Operations**
   ```json
   "failure_mode": "continue"  // Don't let one failure block everything
   ```

4. **Monitor Parallel Execution Duration**
   - Execution time should be ~max(branch_times), not sum(branch_times)
   - If not seeing speedup, check for:
     - Resource contention
     - Concurrency limits
     - Shared dependencies

### Resource Consumption

Parallel execution uses more resources:
- **Memory**: Each branch maintains its own execution context
- **Connections**: Multiple concurrent connections to databases/APIs
- **Goroutines**: One goroutine per concurrent branch

**Recommendations:**
- Use `max_concurrency` to control resource usage
- Monitor memory and connection pool usage
- Set reasonable timeouts
- Consider using queues for large-scale parallelism

## Error Handling

### Error Propagation

#### `stop_all` Mode
```
Branch 1: Running → Success
Branch 2: Running → Error (stops all)
Branch 3: Cancelled
Result: Error returned, branches 1 and 3 cancelled
```

#### `continue` Mode
```
Branch 1: Running → Success
Branch 2: Running → Error (logged)
Branch 3: Running → Success
Result: Partial results with error information
```

### Error Information

Results include error details for each branch:

```json
{
  "branches": [
    {
      "branch_name": "email",
      "output": {...},
      "error": null,
      "duration_ms": 150
    },
    {
      "branch_name": "sms",
      "output": null,
      "error": "API rate limit exceeded",
      "duration_ms": 100
    }
  ],
  "total_branches": 2,
  "completed_branches": 2
}
```

## Advanced Patterns

### Nested Parallel Execution

Parallel nodes within parallel branches:

```json
{
  "branches": [
    {
      "name": "notifications_group",
      "nodes": ["nested_parallel_notifications"]
    },
    {
      "name": "data_processing_group",
      "nodes": ["nested_parallel_processing"]
    }
  ]
}
```

**Use Cases:**
- Hierarchical task organization
- Multi-level parallelism
- Complex workflow orchestration

### Conditional Parallel Execution

Use with conditional nodes to execute parallel branches based on conditions:

```workflow
[Condition: priority == "high"]
    ├─ true  → [Parallel: fast_notifications]
    └─ false → [Sequential: standard_process]
```

### Parallel with Retry

Combine with retry logic for resilient parallel execution:

```json
{
  "type": "control:parallel",
  "config": {
    "branches": [...],
    "retry_config": {
      "max_attempts": 3,
      "strategy": "exponential"
    }
  }
}
```

## Frontend Integration

### React Component

```tsx
import { ParallelNode } from '@/components/nodes/ParallelNode';

// ParallelNode displays:
// - Branch count
// - Wait mode indicator
// - Concurrency limit
// - Timeout
// - Visual branch connections
```

### Configuration Panel

```tsx
import { ParallelConfigPanel } from '@/components/canvas/ParallelConfigPanel';

// Features:
// - Add/remove branches
// - Name branches
// - Assign nodes to branches
// - Configure wait mode
// - Set concurrency limits
// - Set timeout
// - Choose failure mode
```

### Canvas Visualization

Parallel nodes display with:
- Multiple connection points (one per branch)
- Branch name labels
- Color-coded branches
- Collapse/expand for readability

## Troubleshooting

### Parallel Execution Not Faster

**Possible Causes:**
1. Operations are CPU-bound, not I/O-bound
2. Concurrency limit too low
3. Shared resource contention (database connection pool, etc.)
4. Operations have hidden dependencies

**Solutions:**
- Profile to identify bottlenecks
- Increase `max_concurrency`
- Increase resource pool sizes
- Review dependencies

### Timeout Issues

**Symptoms:**
- Parallel execution times out
- Some branches never complete

**Solutions:**
- Increase timeout value
- Reduce concurrency (less contention)
- Optimize slow branches
- Add branch-level timeouts

### Memory Issues

**Symptoms:**
- High memory usage during parallel execution
- Out of memory errors

**Solutions:**
- Reduce `max_concurrency`
- Optimize branch execution
- Stream large data instead of loading all at once
- Monitor and set memory limits

### Inconsistent Results

**Symptoms:**
- Different results on different runs
- Race conditions

**Solutions:**
- Ensure branches are truly independent
- Avoid shared mutable state
- Use proper synchronization if needed
- Review data dependencies

## Backward Compatibility

### Legacy `error_strategy` Field

For backward compatibility, the old `error_strategy` field is still supported:

```json
{
  "error_strategy": "fail_fast"   // Maps to failure_mode: "stop_all"
  "error_strategy": "wait_all"    // Maps to failure_mode: "continue"
}
```

**Recommendation:** Use the new `failure_mode` field for clearer semantics.

### Migration Guide

**Old Format:**
```json
{
  "error_strategy": "fail_fast",
  "max_concurrency": 5
}
```

**New Format:**
```json
{
  "branches": [
    {"name": "branch1", "nodes": ["node1"]},
    {"name": "branch2", "nodes": ["node2"]}
  ],
  "failure_mode": "stop_all",
  "wait_mode": "all",
  "max_concurrency": 5,
  "timeout": "30s"
}
```

## API Reference

### Configuration Interface

```go
type ParallelConfig struct {
    Branches       []ParallelBranch `json:"branches"`
    WaitMode       string            `json:"wait_mode"`        // "all" | "first"
    MaxConcurrency int               `json:"max_concurrency"`  // 0 = unlimited
    Timeout        string            `json:"timeout"`          // Duration string
    FailureMode    string            `json:"failure_mode"`     // "stop_all" | "continue"
}

type ParallelBranch struct {
    Name  string   `json:"name"`   // Branch identifier
    Nodes []string `json:"nodes"`  // Node IDs to execute
}
```

### Result Interface

```go
type ParallelResult struct {
    Branches          []BranchResult `json:"branches"`
    TotalBranches     int            `json:"total_branches"`
    CompletedBranches int            `json:"completed_branches"`
}

type BranchResult struct {
    BranchName string                 `json:"branch_name"`
    Output     map[string]interface{} `json:"output"`
    Error      string                 `json:"error"`
    DurationMs int64                  `json:"duration_ms"`
}
```

## Examples Repository

See the [examples directory](../examples/) for complete workflow examples:

- `parallel-notifications.json` - Multi-channel notification pattern
- `parallel-data-fetching.json` - Concurrent API queries
- `parallel-race-condition.json` - First-wins pattern
- `parallel-batch-processing.json` - Rate-limited processing
- `parallel-nested.json` - Nested parallel execution

## Related Documentation

- [Workflow Execution](./WORKFLOW_EXECUTION.md)
- [Error Handling](./ERROR_HANDLING.md)
- [Performance Optimization](./PERFORMANCE.md)
- [Testing Workflows](./TESTING.md)
