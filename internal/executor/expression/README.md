# Expression Package

The expression package provides powerful expression evaluation capabilities for conditional actions in workflows. It supports boolean expressions with JSONPath-like variable access and comprehensive operator support.

## Features

- **Safe Expression Evaluation**: Uses the `expr-lang/expr` library for secure expression evaluation
- **JSONPath Support**: Access nested data with dot notation (e.g., `steps.step1.output.field`)
- **Array Indexing**: Access array elements (e.g., `steps.step1.data[0].name`)
- **Rich Operators**: Support for comparison, logical, and string operators
- **Template Syntax**: Optional `{{expression}}` wrapper for consistency with other template systems

## Usage

### Basic Evaluation

```go
import "github.com/gorax/gorax/internal/executor/expression"

// Create evaluator
evaluator := expression.NewEvaluator()

// Build context from execution data
context := expression.BuildContext(
    triggerData,
    stepOutputs,
    envVariables,
)

// Evaluate a boolean condition
result, err := evaluator.EvaluateCondition(
    "steps.step1.status == \"success\" && trigger.body.count > 10",
    context,
)

if err != nil {
    log.Fatal(err)
}

if result {
    // Take true branch
} else {
    // Take false branch
}
```

### Expression Context

The expression context provides access to three main data sources:

1. **`trigger`**: Data from the workflow trigger
   ```
   trigger.body.field
   trigger.headers.contentType
   ```

2. **`steps`**: Output data from previous workflow steps
   ```
   steps.step1.output.count
   steps.http_request.status
   steps.transform.result[0].name
   ```

3. **`env`**: Environment variables from the execution context
   ```
   env.tenant_id
   env.execution_id
   env.workflow_id
   ```

### Supported Operators

#### Comparison Operators
- `==` or `equals`: Equality
- `!=` or `not_equals`: Inequality
- `>` or `greater_than`: Greater than
- `>=` or `greater_or_equal`: Greater than or equal
- `<` or `less_than`: Less than
- `<=` or `less_or_equal`: Less than or equal

#### Logical Operators
- `&&` or `and`: Logical AND
- `||` or `or`: Logical OR
- `!` or `not`: Logical NOT

#### String Operators
- `contains`: Check if string contains substring
- `starts_with`: Check if string starts with prefix
- `ends_with`: Check if string ends with suffix

### Example Expressions

#### Simple Comparisons
```javascript
steps.step1.status == "success"
trigger.body.count > 10
steps.http.statusCode < 400
```

#### Complex Conditions
```javascript
// Multiple conditions with AND
steps.step1.status == "success" && trigger.body.count > 10

// Multiple conditions with OR
trigger.body.type == "urgent" || trigger.body.priority > 5

// Negation
!(steps.step1.failed == true)

// Grouped conditions
(steps.step1.status == "success" && steps.step1.count > 10) || trigger.body.override == true
```

#### Array Access
```javascript
// Access array element
trigger.body.items[0].status == "active"

// Access nested arrays
steps.step1.data[0].users[1].name == "Alice"
```

#### String Operations
```javascript
// Contains
trigger.body.message contains "error"

// Starts with
steps.http.url starts_with "https://"

// Ends with
trigger.body.filename ends_with ".json"
```

## Parsing Expressions

The parser can extract variable paths and validate syntax:

```go
parser := expression.NewParser()

// Parse an expression
expr, err := parser.Parse("{{steps.step1.status}} == \"success\"")
if err != nil {
    log.Fatal(err)
}

// Extract all variable paths
paths := parser.ExtractPaths("steps.step1.output.count > 10 && trigger.body.type == \"webhook\"")
// Returns: ["steps.step1.output.count", "trigger.body.type"]

// Validate expression syntax
err = parser.ValidateExpression("steps.step1.count > 10")
if err != nil {
    log.Fatal(err)
}
```

## Workflow Integration

### Conditional Node Configuration

In workflow definitions, conditional nodes use the `ConditionalActionConfig`:

```json
{
  "id": "condition-1",
  "type": "control:if",
  "data": {
    "name": "Check Status",
    "config": {
      "condition": "steps.http.status == 200",
      "true_branch": "success-action",
      "false_branch": "error-handler",
      "description": "Check if HTTP request was successful",
      "stop_on_true": false,
      "stop_on_false": false
    }
  }
}
```

### Edge Labels

Edges from conditional nodes must be labeled to indicate which branch they represent:

```json
{
  "id": "e1",
  "source": "condition-1",
  "target": "success-action",
  "label": "true"
},
{
  "id": "e2",
  "source": "condition-1",
  "target": "error-handler",
  "label": "false"
}
```

## Performance Considerations

### Expression Compilation

For workflows that execute frequently, expressions can be pre-compiled:

```go
evaluator := expression.NewEvaluator()

// Compile expression once
program, err := evaluator.CompileExpression(
    "steps.step1.count > 10",
    mockContext,
)

// Run compiled expression multiple times
for _, execution := range executions {
    result, err := evaluator.EvaluateWithProgram(program, execution.Context)
    // Use result...
}
```

### Caching

The evaluator uses the `expr-lang/expr` library which performs internal optimizations. For best performance:

1. Reuse evaluator instances across multiple evaluations
2. Pre-compile expressions for frequently-executed workflows
3. Build context maps once per execution rather than per expression

## Error Handling

The expression evaluator provides detailed error messages:

```go
result, err := evaluator.EvaluateCondition(expr, context)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "failed to parse"):
        // Expression syntax error
    case strings.Contains(err.Error(), "failed to compile"):
        // Invalid expression structure
    case strings.Contains(err.Error(), "failed to evaluate"):
        // Runtime evaluation error (e.g., type mismatch)
    case strings.Contains(err.Error(), "did not evaluate to boolean"):
        // Expression returned non-boolean result
    }
}
```

## Security

The expression evaluator is designed to be safe:

1. **Sandboxed Execution**: Uses `expr-lang/expr` which prevents code injection
2. **Type Safety**: Enforces type checking at compile time
3. **No Side Effects**: Expressions are read-only and cannot modify data
4. **Resource Limits**: Expression evaluation is bounded and cannot cause infinite loops

## Testing

The package includes comprehensive tests for:

- Expression parsing and validation
- Boolean condition evaluation
- Array access and nested objects
- All supported operators
- Edge cases and error conditions

Run tests with:
```bash
go test ./internal/executor/expression/... -v
```

## Examples

See `evaluator_test.go` and `parser_test.go` for extensive examples of expression usage.
