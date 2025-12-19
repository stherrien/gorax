# Phase 2.4 Execution History - Usage Examples

## Overview
This document provides practical examples of using the new execution history features implemented in Phase 2.4.

## Repository Methods

### 1. List Executions with Advanced Filtering

#### Basic Listing (No Filters)
```go
ctx := context.Background()
tenantID := "tenant-123"
filter := workflow.ExecutionFilter{}
cursor := ""
limit := 20

result, err := repo.ListExecutionsAdvanced(ctx, tenantID, filter, cursor, limit)
if err != nil {
    return fmt.Errorf("list executions: %w", err)
}

fmt.Printf("Found %d executions (total: %d)\n", len(result.Data), result.TotalCount)
fmt.Printf("Has more pages: %v\n", result.HasMore)
```

#### Filter by Workflow ID
```go
filter := workflow.ExecutionFilter{
    WorkflowID: "workflow-456",
}

result, err := repo.ListExecutionsAdvanced(ctx, tenantID, filter, "", 20)
```

#### Filter by Status
```go
filter := workflow.ExecutionFilter{
    Status: "completed",
}

result, err := repo.ListExecutionsAdvanced(ctx, tenantID, filter, "", 20)
```

#### Filter by Trigger Type
```go
filter := workflow.ExecutionFilter{
    TriggerType: "webhook",
}

result, err := repo.ListExecutionsAdvanced(ctx, tenantID, filter, "", 20)
```

#### Filter by Date Range
```go
startDate := time.Now().Add(-24 * time.Hour)
endDate := time.Now()

filter := workflow.ExecutionFilter{
    StartDate: &startDate,
    EndDate:   &endDate,
}

result, err := repo.ListExecutionsAdvanced(ctx, tenantID, filter, "", 20)
```

#### Combined Filters
```go
startDate := time.Now().Add(-7 * 24 * time.Hour)
endDate := time.Now()

filter := workflow.ExecutionFilter{
    WorkflowID:  "workflow-456",
    Status:      "completed",
    TriggerType: "webhook",
    StartDate:   &startDate,
    EndDate:     &endDate,
}

result, err := repo.ListExecutionsAdvanced(ctx, tenantID, filter, "", 20)
```

### 2. Cursor-Based Pagination

#### Navigate Through Pages
```go
func listAllExecutions(ctx context.Context, repo *workflow.Repository, tenantID string) error {
    filter := workflow.ExecutionFilter{}
    cursor := ""
    limit := 50
    page := 1

    for {
        result, err := repo.ListExecutionsAdvanced(ctx, tenantID, filter, cursor, limit)
        if err != nil {
            return fmt.Errorf("list page %d: %w", page, err)
        }

        fmt.Printf("Page %d: %d executions\n", page, len(result.Data))

        // Process executions
        for _, exec := range result.Data {
            fmt.Printf("  - %s: %s (%s)\n", exec.ID, exec.Status, exec.CreatedAt)
        }

        // Check if there are more pages
        if !result.HasMore {
            fmt.Println("Reached last page")
            break
        }

        // Move to next page
        cursor = result.Cursor
        page++
    }

    return nil
}
```

#### Infinite Scroll Implementation
```go
// For a web API endpoint that supports infinite scroll
func handleListExecutions(w http.ResponseWriter, r *http.Request) {
    tenantID := getTenantID(r)
    cursor := r.URL.Query().Get("cursor")
    limit := 20

    filter := workflow.ExecutionFilter{
        WorkflowID: r.URL.Query().Get("workflow_id"),
        Status:     r.URL.Query().Get("status"),
    }

    result, err := repo.ListExecutionsAdvanced(r.Context(), tenantID, filter, cursor, limit)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return JSON response with cursor for next page
    json.NewEncoder(w).Encode(result)
}
```

### 3. Get Execution with Steps

#### Basic Usage
```go
executionID := "exec-123"

result, err := repo.GetExecutionWithSteps(ctx, tenantID, executionID)
if err != nil {
    if errors.Is(err, workflow.ErrNotFound) {
        return fmt.Errorf("execution not found")
    }
    return fmt.Errorf("get execution: %w", err)
}

fmt.Printf("Execution: %s (%s)\n", result.Execution.ID, result.Execution.Status)
fmt.Printf("Steps: %d\n", len(result.Steps))

for i, step := range result.Steps {
    fmt.Printf("  Step %d: %s - %s (%s)\n", i+1, step.NodeID, step.NodeType, step.Status)
    if step.DurationMs != nil {
        fmt.Printf("    Duration: %dms\n", *step.DurationMs)
    }
}
```

#### Calculate Total Execution Time
```go
result, err := repo.GetExecutionWithSteps(ctx, tenantID, executionID)
if err != nil {
    return 0, err
}

var totalDuration int
for _, step := range result.Steps {
    if step.DurationMs != nil {
        totalDuration += *step.DurationMs
    }
}

fmt.Printf("Total execution time: %dms\n", totalDuration)
```

#### Identify Failed Steps
```go
result, err := repo.GetExecutionWithSteps(ctx, tenantID, executionID)
if err != nil {
    return nil, err
}

var failedSteps []*workflow.StepExecution
for _, step := range result.Steps {
    if step.Status == "failed" {
        failedSteps = append(failedSteps, step)
    }
}

if len(failedSteps) > 0 {
    fmt.Printf("Found %d failed steps:\n", len(failedSteps))
    for _, step := range failedSteps {
        fmt.Printf("  - %s: %s\n", step.NodeID, *step.ErrorMessage)
    }
}
```

### 4. Count Executions

#### Basic Count
```go
filter := workflow.ExecutionFilter{}

count, err := repo.CountExecutions(ctx, tenantID, filter)
if err != nil {
    return fmt.Errorf("count executions: %w", err)
}

fmt.Printf("Total executions: %d\n", count)
```

#### Count by Filter
```go
filter := workflow.ExecutionFilter{
    Status: "failed",
}

count, err := repo.CountExecutions(ctx, tenantID, filter)
if err != nil {
    return fmt.Errorf("count failed executions: %w", err)
}

fmt.Printf("Failed executions: %d\n", count)
```

#### Calculate Success Rate
```go
func calculateSuccessRate(ctx context.Context, repo *workflow.Repository, tenantID, workflowID string) (float64, error) {
    filter := workflow.ExecutionFilter{
        WorkflowID: workflowID,
    }

    total, err := repo.CountExecutions(ctx, tenantID, filter)
    if err != nil {
        return 0, err
    }

    if total == 0 {
        return 0, nil
    }

    filter.Status = "completed"
    completed, err := repo.CountExecutions(ctx, tenantID, filter)
    if err != nil {
        return 0, err
    }

    successRate := (float64(completed) / float64(total)) * 100
    return successRate, nil
}
```

## API Handler Examples

### REST API Endpoint
```go
// GET /api/executions?workflow_id=xxx&status=yyy&cursor=zzz&limit=20
func handleListExecutions(w http.ResponseWriter, r *http.Request) {
    tenantID := getTenantFromContext(r.Context())

    // Parse query parameters
    workflowID := r.URL.Query().Get("workflow_id")
    status := r.URL.Query().Get("status")
    triggerType := r.URL.Query().Get("trigger_type")
    cursor := r.URL.Query().Get("cursor")

    limit := 20
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
            limit = parsed
        }
    }

    // Parse date filters
    var startDate, endDate *time.Time
    if start := r.URL.Query().Get("start_date"); start != "" {
        if parsed, err := time.Parse(time.RFC3339, start); err == nil {
            startDate = &parsed
        }
    }
    if end := r.URL.Query().Get("end_date"); end != "" {
        if parsed, err := time.Parse(time.RFC3339, end); err == nil {
            endDate = &parsed
        }
    }

    // Build filter
    filter := workflow.ExecutionFilter{
        WorkflowID:  workflowID,
        Status:      status,
        TriggerType: triggerType,
        StartDate:   startDate,
        EndDate:     endDate,
    }

    // Validate filter
    if err := filter.Validate(); err != nil {
        http.Error(w, fmt.Sprintf("invalid filter: %v", err), http.StatusBadRequest)
        return
    }

    // Query database
    result, err := repo.ListExecutionsAdvanced(r.Context(), tenantID, filter, cursor, limit)
    if err != nil {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        log.Printf("list executions error: %v", err)
        return
    }

    // Return response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### GraphQL Resolver
```go
type executionResolver struct {
    repo *workflow.Repository
}

func (r *executionResolver) Executions(
    ctx context.Context,
    workflowID *string,
    status *string,
    triggerType *string,
    startDate *time.Time,
    endDate *time.Time,
    cursor *string,
    limit *int,
) (*workflow.ExecutionListResult, error) {
    tenantID := getTenantFromContext(ctx)

    filter := workflow.ExecutionFilter{}
    if workflowID != nil {
        filter.WorkflowID = *workflowID
    }
    if status != nil {
        filter.Status = *status
    }
    if triggerType != nil {
        filter.TriggerType = *triggerType
    }
    if startDate != nil {
        filter.StartDate = startDate
    }
    if endDate != nil {
        filter.EndDate = endDate
    }

    cursorStr := ""
    if cursor != nil {
        cursorStr = *cursor
    }

    limitInt := 20
    if limit != nil && *limit > 0 && *limit <= 100 {
        limitInt = *limit
    }

    return r.repo.ListExecutionsAdvanced(ctx, tenantID, filter, cursorStr, limitInt)
}

func (r *executionResolver) Execution(
    ctx context.Context,
    id string,
) (*workflow.ExecutionWithSteps, error) {
    tenantID := getTenantFromContext(ctx)
    return r.repo.GetExecutionWithSteps(ctx, tenantID, id)
}
```

## Frontend Integration Examples

### React Hook for Infinite Scroll
```typescript
import { useState, useEffect } from 'react';

interface UseExecutionListOptions {
  workflowId?: string;
  status?: string;
  triggerType?: string;
  limit?: number;
}

export function useExecutionList(options: UseExecutionListOptions) {
  const [executions, setExecutions] = useState([]);
  const [cursor, setCursor] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);

  const loadMore = async () => {
    if (loading || !hasMore) return;

    setLoading(true);
    try {
      const params = new URLSearchParams({
        limit: String(options.limit || 20),
        ...(cursor && { cursor }),
        ...(options.workflowId && { workflow_id: options.workflowId }),
        ...(options.status && { status: options.status }),
        ...(options.triggerType && { trigger_type: options.triggerType }),
      });

      const response = await fetch(`/api/executions?${params}`);
      const result = await response.json();

      setExecutions(prev => [...prev, ...result.data]);
      setCursor(result.cursor);
      setHasMore(result.has_more);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadMore();
  }, []);

  return { executions, loadMore, hasMore, loading };
}
```

### Vue 3 Composable
```typescript
import { ref, computed } from 'vue';

export function useExecutionList(options: ExecutionListOptions) {
  const executions = ref<Execution[]>([]);
  const cursor = ref<string | null>(null);
  const hasMore = ref(true);
  const loading = ref(false);
  const totalCount = ref(0);

  async function loadMore() {
    if (loading.value || !hasMore.value) return;

    loading.value = true;
    try {
      const params = new URLSearchParams({
        limit: String(options.limit || 20),
        ...(cursor.value && { cursor: cursor.value }),
        ...(options.workflowId && { workflow_id: options.workflowId }),
      });

      const response = await fetch(`/api/executions?${params}`);
      const result = await response.json();

      executions.value.push(...result.data);
      cursor.value = result.cursor;
      hasMore.value = result.has_more;
      totalCount.value = result.total_count;
    } finally {
      loading.value = false;
    }
  }

  function reset() {
    executions.value = [];
    cursor.value = null;
    hasMore.value = true;
    totalCount.value = 0;
  }

  return {
    executions,
    loadMore,
    reset,
    hasMore: computed(() => hasMore.value),
    loading: computed(() => loading.value),
    totalCount: computed(() => totalCount.value),
  };
}
```

## Performance Tips

### 1. Choose Appropriate Page Size
```go
// For real-time dashboards: smaller pages, more frequent updates
limit := 10

// For bulk processing: larger pages
limit := 100

// For infinite scroll: medium pages
limit := 20
```

### 2. Use Specific Filters
```go
// GOOD: Use specific filters to reduce result set
filter := workflow.ExecutionFilter{
    WorkflowID: "workflow-123",
    Status:     "failed",
}

// AVOID: Fetching all executions without filters
filter := workflow.ExecutionFilter{}
```

### 3. Cache Total Counts
```go
// Cache total count to avoid repeated COUNT queries
var cachedCount int
var cacheTime time.Time
var cacheDuration = 5 * time.Minute

func getTotalCount(ctx context.Context) (int, error) {
    if time.Since(cacheTime) < cacheDuration {
        return cachedCount, nil
    }

    count, err := repo.CountExecutions(ctx, tenantID, filter)
    if err != nil {
        return 0, err
    }

    cachedCount = count
    cacheTime = time.Now()
    return count, nil
}
```

### 4. Parallel Queries for Aggregations
```go
func getExecutionStats(ctx context.Context) (*Stats, error) {
    var wg sync.WaitGroup
    var total, completed, failed, running int
    var errs []error

    wg.Add(4)

    go func() {
        defer wg.Done()
        if count, err := repo.CountExecutions(ctx, tenantID, workflow.ExecutionFilter{}); err == nil {
            total = count
        } else {
            errs = append(errs, err)
        }
    }()

    go func() {
        defer wg.Done()
        if count, err := repo.CountExecutions(ctx, tenantID, workflow.ExecutionFilter{Status: "completed"}); err == nil {
            completed = count
        } else {
            errs = append(errs, err)
        }
    }()

    // ... similar for failed and running

    wg.Wait()

    if len(errs) > 0 {
        return nil, errs[0]
    }

    return &Stats{Total: total, Completed: completed, Failed: failed, Running: running}, nil
}
```
