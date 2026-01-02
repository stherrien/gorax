# Gorax Workflow Specification

Complete technical specification for designing, building, and executing workflows in Gorax.

**Version:** 1.0
**Last Updated:** 2026-01-01

---

## Table of Contents

1. [Overview](#1-overview)
2. [Workflow JSON Schema](#2-workflow-json-schema)
3. [Node Types](#3-node-types)
4. [Edge Connections](#4-edge-connections)
5. [Expression Syntax](#5-expression-syntax)
6. [Validation Rules](#6-validation-rules)
7. [Workflow Variables](#7-workflow-variables)
8. [Execution Model](#8-execution-model)
9. [Examples](#9-examples)
10. [Best Practices](#10-best-practices)

---

## 1. Overview

### What is a Workflow?

A **workflow** in Gorax is a directed acyclic graph (DAG) of **nodes** connected by **edges** that defines an automated business process. Each workflow:

- Starts with a **trigger** (webhook, schedule, or manual invocation)
- Executes a series of **actions** (HTTP requests, data transformations, code execution)
- Can include **control flow** (conditionals, loops, parallel execution)
- Produces an **output** that can be used by downstream systems

### Workflow Definition Structure

```json
{
  "nodes": [ /* array of node objects */ ],
  "edges": [ /* array of edge objects */ ]
}
```

A workflow consists of:
- **Nodes**: Individual steps or operations (triggers, actions, control flow)
- **Edges**: Connections that define execution order and data flow

### Execution Model

Gorax uses a **topological sort** algorithm to determine execution order:

1. **Parse** the workflow definition (nodes + edges)
2. **Validate** the graph structure (no cycles, valid connections)
3. **Sort** nodes by dependencies (topological order)
4. **Execute** nodes sequentially, storing outputs in context
5. **Handle** errors with retry logic and circuit breakers
6. **Broadcast** execution events in real-time (optional)

---

## 2. Workflow JSON Schema

### Complete Workflow Structure

```json
{
  "nodes": [
    {
      "id": "unique-node-id",
      "type": "node:type",
      "position": { "x": 100, "y": 200 },
      "data": {
        "name": "Human-readable name",
        "config": { /* type-specific configuration */ }
      }
    }
  ],
  "edges": [
    {
      "id": "unique-edge-id",
      "source": "source-node-id",
      "target": "target-node-id",
      "sourceHandle": "optional-source-handle",
      "targetHandle": "optional-target-handle",
      "label": "optional-label"
    }
  ]
}
```

### Node Object Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier for the node (used in edges and context) |
| `type` | string | Yes | Node type (e.g., `trigger:webhook`, `action:http`) |
| `position` | object | Yes | Canvas position `{x: number, y: number}` |
| `data` | object | Yes | Node data containing `name` and `config` |
| `data.name` | string | Yes | Human-readable name for the node |
| `data.config` | object | Yes | Type-specific configuration (varies by node type) |

### Edge Object Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier for the edge |
| `source` | string | Yes | Source node ID |
| `target` | string | Yes | Target node ID |
| `sourceHandle` | string | No | Source connection point (for multi-output nodes) |
| `targetHandle` | string | No | Target connection point (for multi-input nodes) |
| `label` | string | No | Edge label (e.g., "true"/"false" for conditionals) |

### Validation Rules

1. **Node IDs must be unique** across the workflow
2. **Edge IDs must be unique** across the workflow
3. **Source and target nodes must exist** in the nodes array
4. **No cycles allowed** (except for loops which have special handling)
5. **At least one trigger node** must be present
6. **Positions must be valid numbers** (for canvas rendering)

---

## 3. Node Types

### 3.1 Trigger Nodes

Trigger nodes initiate workflow execution. Only **one trigger** executes per workflow run.

#### Webhook Trigger (`trigger:webhook`)

Starts a workflow when an HTTP request is received.

**Configuration:**

```json
{
  "type": "trigger:webhook",
  "data": {
    "name": "GitHub Webhook",
    "config": {
      "path": "/webhook/github",
      "auth_type": "signature",
      "secret": "{{credentials.github_webhook_secret}}",
      "allowed_ips": "192.30.252.0/22",
      "response_url": ""
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `path` | string | No | Custom webhook path (auto-generated if omitted) |
| `auth_type` | string | No | Authentication: `none`, `basic`, `signature`, `api_key` |
| `secret` | string | No | Secret for signature validation (HMAC-SHA256) |
| `allowed_ips` | string | No | Comma-separated list of allowed IP addresses/CIDR ranges |
| `response_url` | string | No | URL to respond to after execution |

**Output Context:**

```json
{
  "trigger": {
    "method": "POST",
    "path": "/webhook/github",
    "headers": { "Content-Type": "application/json" },
    "body": { /* parsed JSON body */ },
    "query": { /* query parameters */ }
  }
}
```

#### Schedule Trigger (`trigger:schedule`)

Starts a workflow on a schedule using cron syntax.

**Configuration:**

```json
{
  "type": "trigger:schedule",
  "data": {
    "name": "Daily Report",
    "config": {
      "cron": "0 9 * * MON-FRI",
      "timezone": "America/New_York"
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cron` | string | Yes | Cron expression (standard 5-field format) |
| `timezone` | string | No | IANA timezone (default: UTC) |

**Cron Format:** `minute hour day month weekday`

**Examples:**
- `0 9 * * *` - Daily at 9:00 AM
- `*/15 * * * *` - Every 15 minutes
- `0 0 1 * *` - First day of every month at midnight
- `0 9 * * MON-FRI` - Weekdays at 9:00 AM

**Output Context:**

```json
{
  "trigger": {
    "type": "schedule",
    "timestamp": "2026-01-01T09:00:00Z",
    "cron": "0 9 * * MON-FRI"
  }
}
```

---

### 3.2 Action Nodes

Action nodes perform operations like HTTP requests, data transformations, or code execution.

#### HTTP Request (`action:http`)

Makes an HTTP request to an external API.

**Configuration:**

```json
{
  "type": "action:http",
  "data": {
    "name": "Fetch User Data",
    "config": {
      "method": "GET",
      "url": "https://api.example.com/users/{{trigger.body.user_id}}",
      "headers": {
        "Authorization": "Bearer {{credentials.api_token}}",
        "Content-Type": "application/json"
      },
      "body": { "key": "value" },
      "timeout": 30,
      "auth": {
        "type": "bearer",
        "token": "{{credentials.api_token}}"
      },
      "follow_redirects": true
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `method` | string | Yes | HTTP method: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS |
| `url` | string | Yes | Target URL (supports template variables) |
| `headers` | object | No | HTTP headers (supports template variables) |
| `body` | object/string | No | Request body (auto-serialized to JSON) |
| `timeout` | number | No | Timeout in seconds (default: 30) |
| `auth` | object | No | Authentication configuration |
| `auth.type` | string | No | Auth type: `basic`, `bearer`, `api_key` |
| `auth.username` | string | No | Username for basic auth |
| `auth.password` | string | No | Password for basic auth |
| `auth.token` | string | No | Token for bearer auth |
| `auth.api_key` | string | No | API key for api_key auth |
| `auth.header` | string | No | Header name for api_key (default: X-API-Key) |
| `follow_redirects` | boolean | No | Follow HTTP redirects (default: true) |

**Output:**

```json
{
  "status_code": 200,
  "headers": {
    "Content-Type": "application/json",
    "Content-Length": "1234"
  },
  "body": { /* parsed response body */ }
}
```

#### Transform (`action:transform`)

Extracts and transforms data using JSONPath expressions.

**Configuration:**

```json
{
  "type": "action:transform",
  "data": {
    "name": "Extract User Info",
    "config": {
      "expression": "steps.http-1.body.users[0].name",
      "mapping": {
        "user_id": "trigger.body.id",
        "user_name": "steps.http-1.body.name",
        "user_email": "steps.http-1.body.email"
      },
      "default": null
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `expression` | string | No | JSONPath expression to extract a single value |
| `mapping` | object | No | Map of output fields to source paths |
| `default` | any | No | Default value if extraction fails |

**Use Cases:**
- **Single value extraction**: Use `expression` to extract one value
- **Object mapping**: Use `mapping` to create a new object from multiple paths
- **Fallback**: Use `default` to provide a fallback value

**Output (expression):**

```json
"John Doe"
```

**Output (mapping):**

```json
{
  "user_id": "12345",
  "user_name": "John Doe",
  "user_email": "john@example.com"
}
```

#### Formula (`action:formula`)

Evaluates mathematical or logical expressions using the Expr language.

**Configuration:**

```json
{
  "type": "action:formula",
  "data": {
    "name": "Calculate Total",
    "config": {
      "expression": "steps.http-1.body.price * steps.http-1.body.quantity * 1.08",
      "output_variable": "total_with_tax"
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `expression` | string | Yes | Expr language expression |
| `output_variable` | string | No | Variable name for the result (default: "result") |

**Supported Operations:**
- Arithmetic: `+`, `-`, `*`, `/`, `%`, `**`
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logical: `&&`, `||`, `!`
- String: `+` (concatenation), `contains()`, `startsWith()`, `endsWith()`
- Array: `len()`, `[index]`, `in`
- Math: `abs()`, `ceil()`, `floor()`, `round()`

**Output:**

```json
108.0
```

#### Code Execution (`action:code`)

Executes sandboxed JavaScript code for custom logic.

**Configuration:**

```json
{
  "type": "action:code",
  "data": {
    "name": "Custom Logic",
    "config": {
      "script": "const items = context.steps['http-1'].body.items;\nreturn items.filter(item => item.active).map(item => item.name);",
      "timeout": 30,
      "memory_limit": 128
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `script` | string | Yes | JavaScript code to execute |
| `timeout` | number | No | Timeout in seconds (default: 30, max: 60) |
| `memory_limit` | number | No | Memory limit in MB (future enhancement) |

**Available Context:**
- `context.trigger` - Trigger data
- `context.steps` - Step outputs
- `context.env` - Environment variables

**Security:**
- Sandboxed execution (no file system, network, or process access)
- Enforced timeout prevents infinite loops
- No access to Node.js modules or Go stdlib

**Output:**

```json
{
  "result": [ /* script return value */ ]
}
```

**Example Scripts:**

```javascript
// Filter and transform array
const items = context.steps['http-1'].body.items;
return items
  .filter(item => item.price > 100)
  .map(item => ({
    id: item.id,
    name: item.name,
    discounted_price: item.price * 0.9
  }));
```

```javascript
// Complex business logic
const user = context.trigger.body;
const settings = context.steps['get-settings'].body;

if (user.type === 'premium' && settings.premium_enabled) {
  return { discount: 0.2, priority: 'high' };
} else {
  return { discount: 0.1, priority: 'normal' };
}
```

---

### 3.3 Control Flow Nodes

Control flow nodes alter execution path based on conditions or repeat operations.

#### Conditional (`control:if`)

Evaluates a condition and branches execution.

**Configuration:**

```json
{
  "type": "control:if",
  "data": {
    "name": "Check Status",
    "config": {
      "condition": "steps['http-1'].body.status == 'active'",
      "true_branch": "true",
      "false_branch": "false",
      "description": "Check if user is active",
      "stop_on_true": false,
      "stop_on_false": false
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `condition` | string | Yes | Boolean expression to evaluate |
| `true_branch` | string | No | Edge label for true branch (default: "true") |
| `false_branch` | string | No | Edge label for false branch (default: "false") |
| `description` | string | No | Human-readable description |
| `stop_on_true` | boolean | No | Stop workflow if condition is true |
| `stop_on_false` | boolean | No | Stop workflow if condition is false |

**Edges:**
- Outgoing edges must have `label: "true"` or `label: "false"`
- Only nodes on the taken branch will execute

**Output:**

```json
{
  "condition": "steps['http-1'].body.status == 'active'",
  "result": true,
  "taken_branch": "true",
  "next_nodes": ["node-id-1", "node-id-2"],
  "stop_execution": false
}
```

**Example Graph:**

```
[Trigger] → [HTTP Request] → [Conditional: if status == 'active']
                                  |                     |
                              (true)                (false)
                                  |                     |
                           [Send Email]          [Log Error]
```

#### Loop (`control:loop`)

Iterates over an array, executing body nodes for each item.

**Configuration:**

```json
{
  "type": "control:loop",
  "data": {
    "name": "Process Users",
    "config": {
      "source": "steps['http-1'].body.users",
      "item_variable": "user",
      "index_variable": "index",
      "max_iterations": 1000,
      "on_error": "continue"
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | string | Yes | JSONPath to array (e.g., `steps.node1.output.items`) |
| `item_variable` | string | Yes | Variable name for current item (e.g., "item") |
| `index_variable` | string | No | Variable name for current index (e.g., "index") |
| `max_iterations` | number | No | Safety limit (default: 1000) |
| `on_error` | string | No | Error strategy: `continue` or `stop` (default: "stop") |

**Loop Body:**
- The first outgoing edge defines the loop body entrance
- All nodes reachable from the body entrance are executed for each iteration
- Loop variables are available in context as `steps.{item_variable}` and `steps.{index_variable}`

**Output:**

```json
{
  "iteration_count": 3,
  "iterations": [
    {
      "index": 0,
      "item": { "id": 1, "name": "Alice" },
      "output": { /* outputs from body nodes */ },
      "error": null
    },
    {
      "index": 1,
      "item": { "id": 2, "name": "Bob" },
      "output": { /* outputs from body nodes */ },
      "error": null
    }
  ],
  "metadata": {
    "item_variable": "user",
    "index_variable": "index",
    "on_error": "continue"
  }
}
```

**Example:**

```json
{
  "nodes": [
    {
      "id": "loop-1",
      "type": "control:loop",
      "data": {
        "config": {
          "source": "steps['get-users'].body.users",
          "item_variable": "user",
          "index_variable": "i"
        }
      }
    },
    {
      "id": "send-email",
      "type": "action:http",
      "data": {
        "config": {
          "method": "POST",
          "url": "https://api.sendgrid.com/v3/mail/send",
          "body": {
            "to": "{{steps.user.email}}",
            "subject": "Hello {{steps.user.name}}"
          }
        }
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "loop-1", "target": "send-email" }
  ]
}
```

#### Parallel (`control:parallel`)

Executes multiple branches concurrently.

**Configuration:**

```json
{
  "type": "control:parallel",
  "data": {
    "name": "Parallel API Calls",
    "config": {
      "error_strategy": "fail_fast",
      "max_concurrency": 5
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `error_strategy` | string | No | `fail_fast` (stop on first error) or `wait_all` (wait for all) |
| `max_concurrency` | number | No | Max concurrent branches (0 = unlimited) |

**Branches:**
- Each outgoing edge defines a separate branch
- Branches execute concurrently (goroutines)
- Results are collected when all branches complete

**Output:**

```json
{
  "branch_count": 3,
  "branch_results": [
    {
      "branch_index": 0,
      "output": { /* branch outputs */ },
      "error": null,
      "duration_ms": 1234
    },
    {
      "branch_index": 1,
      "output": { /* branch outputs */ },
      "error": null,
      "duration_ms": 987
    }
  ],
  "metadata": {
    "error_strategy": "fail_fast",
    "max_concurrency": 5
  }
}
```

**Example:**

```
[Trigger] → [Parallel]
                |
         +------+------+------+
         |      |      |      |
      [API 1] [API 2] [API 3] [API 4]
         |      |      |      |
         +------+------+------+
                |
            [Join] → [Process Results]
```

#### Delay (`control:delay`)

Pauses execution for a specified duration.

**Configuration:**

```json
{
  "type": "control:delay",
  "data": {
    "name": "Wait 5 Seconds",
    "config": {
      "duration": "5s"
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `duration` | string | Yes | Duration string: "5s", "1m", "2h" or template variable |

**Supported Units:**
- `s` - seconds
- `m` - minutes
- `h` - hours

**Output:**

```json
{
  "delayed_for": "5s",
  "started_at": "2026-01-01T10:00:00Z",
  "completed_at": "2026-01-01T10:00:05Z"
}
```

#### Sub-Workflow (`control:sub_workflow`)

Executes another workflow as a sub-workflow.

**Configuration:**

```json
{
  "type": "control:sub_workflow",
  "data": {
    "name": "Process Order Sub-Workflow",
    "config": {
      "workflow_id": "wf-12345",
      "input_mapping": {
        "order_id": "trigger.body.order_id",
        "customer_email": "steps['get-customer'].body.email"
      },
      "output_mapping": {
        "result": "output.status",
        "tracking_number": "output.tracking"
      },
      "wait_for_result": true,
      "timeout_ms": 60000
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `workflow_id` | string | Yes | ID of the workflow to execute |
| `input_mapping` | object | No | Map parent context to sub-workflow input |
| `output_mapping` | object | No | Map sub-workflow output to parent context |
| `wait_for_result` | boolean | Yes | Sync (true) vs async (false) execution |
| `timeout_ms` | number | No | Timeout in milliseconds (0 = no timeout) |

**Limits:**
- Maximum nesting depth: 10 levels
- Circular dependencies are detected and prevented

**Output:**

```json
{
  "workflow_id": "wf-12345",
  "execution_id": "exec-67890",
  "status": "completed",
  "output": { /* sub-workflow output */ },
  "duration_ms": 5432
}
```

---

### 3.4 Integration Nodes

Integration nodes connect to external services like Slack, AWS, email providers, etc.

#### Slack: Send Message (`slack:send_message`)

Sends a message to a Slack channel.

**Configuration:**

```json
{
  "type": "slack:send_message",
  "data": {
    "name": "Notify Team",
    "config": {
      "credential_id": "cred-slack-123",
      "channel": "#deployments",
      "text": "Deployment completed for {{trigger.body.app_name}}",
      "blocks": [
        {
          "type": "section",
          "text": {
            "type": "mrkdwn",
            "text": "*Deployment Status*\nApp: {{trigger.body.app_name}}\nStatus: Success"
          }
        }
      ],
      "thread_ts": "",
      "reply_broadcast": false
    }
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `credential_id` | string | Yes | Slack credential ID (OAuth token) |
| `channel` | string | Yes | Channel ID or name (e.g., "#general", "C0123456789") |
| `text` | string | Yes | Message text (supports template variables) |
| `blocks` | array | No | Slack Block Kit blocks for rich formatting |
| `thread_ts` | string | No | Thread timestamp to reply in a thread |
| `reply_broadcast` | boolean | No | Broadcast thread reply to channel |

**Output:**

```json
{
  "ok": true,
  "channel": "C0123456789",
  "ts": "1234567890.123456",
  "message": { /* sent message */ }
}
```

---

## 4. Edge Connections

### Edge Schema

```json
{
  "id": "edge-1",
  "source": "node-1",
  "target": "node-2",
  "sourceHandle": "output-1",
  "targetHandle": "input-1",
  "label": "true"
}
```

### Connection Rules

1. **One source, one target**: Each edge connects exactly one source to one target
2. **Source must exist before target**: Topological ordering is enforced
3. **No cycles**: DAG structure is validated (except for loops)
4. **Multiple edges allowed**: A node can have multiple incoming and outgoing edges

### Conditional Edges

Conditional nodes require edges with labels:

```json
{
  "edges": [
    {
      "id": "e1",
      "source": "conditional-1",
      "target": "action-true",
      "label": "true"
    },
    {
      "id": "e2",
      "source": "conditional-1",
      "target": "action-false",
      "label": "false"
    }
  ]
}
```

**Label values:**
- `"true"` - Taken when condition evaluates to true
- `"false"` - Taken when condition evaluates to false

### Edge Validation

1. **Source and target must exist**: Referenced node IDs must be valid
2. **No self-loops**: A node cannot connect to itself (except loop nodes)
3. **Conditional labels**: If/else nodes must have labeled edges
4. **Unique labels**: Each label from a source node must be unique

---

## 5. Expression Syntax

Gorax supports two expression syntaxes:

1. **Template variables**: `{{path.to.value}}` for simple variable interpolation
2. **Expr language**: Full expressions for conditions and formulas

### Template Variables

Template variables use double curly braces: `{{expression}}`

**Syntax:**

```
{{trigger.body.field}}
{{steps.node-id.output.data}}
{{env.tenant_id}}
{{credentials.api_token}}
```

**Access Paths:**
- `trigger.*` - Trigger data (webhook body, schedule info)
- `steps.{node-id}.*` - Output from a specific node
- `env.*` - Environment variables (tenant_id, execution_id, workflow_id)
- `credentials.{name}` - Credential values (auto-injected)

**Array Indexing:**

```
{{steps.http-1.body.users[0].name}}
{{trigger.body.items[2].price}}
```

**Nested Objects:**

```
{{steps.http-1.body.user.profile.email}}
{{trigger.body.data.nested.deep.value}}
```

### Expr Language

Used for conditions (if nodes) and formulas (formula nodes).

**Syntax Examples:**

```javascript
// Comparisons
steps['http-1'].body.status == 'active'
steps['http-1'].body.count > 100
trigger.body.age >= 18

// Logical operators
steps['http-1'].body.status == 'active' && steps['http-1'].body.verified
trigger.body.type == 'premium' || trigger.body.type == 'enterprise'
!steps['http-1'].body.disabled

// Arithmetic
steps['http-1'].body.price * 1.08
steps['http-1'].body.quantity + 10
(steps['http-1'].body.total - steps['http-1'].body.discount) * 1.05

// String operations
trigger.body.email contains '@example.com'
steps['http-1'].body.name startsWith 'John'
trigger.body.status endsWith '_pending'

// Array operations
len(steps['http-1'].body.items) > 0
'premium' in trigger.body.features
steps['http-1'].body.users[0].active

// Null checks
steps['http-1'].body.optional != null
trigger.body.field == null
```

**Built-in Functions:**

| Function | Description | Example |
|----------|-------------|---------|
| `len(arr)` | Array/string length | `len(steps.http1.body.items) > 0` |
| `contains(str, substr)` | String contains | `trigger.body.email contains '@'` |
| `startsWith(str, prefix)` | String starts with | `steps.http1.body.name startsWith 'A'` |
| `endsWith(str, suffix)` | String ends with | `trigger.body.file endsWith '.pdf'` |
| `abs(num)` | Absolute value | `abs(steps.calc.result) > 100` |
| `ceil(num)` | Round up | `ceil(steps.calc.price)` |
| `floor(num)` | Round down | `floor(steps.calc.discount)` |
| `round(num)` | Round to nearest | `round(steps.calc.total)` |

### Type Coercion

Gorax automatically converts types when possible:

- Numbers to strings: `"123"` → `123`
- Strings to numbers: `123` → `"123"`
- Booleans to strings: `true` → `"true"`

### Error Handling

If a path is not found:
- **Template variables**: Return the original template string (e.g., `{{missing.path}}`)
- **Expr expressions**: Throw an error and fail the node

---

## 6. Validation Rules

### Graph Validation

1. **No cycles**: The workflow must be a DAG (directed acyclic graph)
   - Exception: Loop nodes are allowed to create controlled cycles
   - Circular sub-workflow calls are detected and prevented

2. **Connected graph**: All nodes must be reachable from a trigger
   - Orphaned nodes (no incoming edges) are flagged as warnings
   - Dead-end nodes (no outgoing edges) are allowed

3. **Single trigger**: Each workflow must have exactly one trigger node
   - Multiple triggers are not supported in a single workflow

### Node Validation

1. **Required fields**: All required configuration fields must be present
   - `id`, `type`, `position`, `data.name`, `data.config`

2. **Type checking**: Configuration fields must match expected types
   - Strings, numbers, booleans, objects, arrays

3. **Unique IDs**: Node IDs must be unique within the workflow

4. **Valid types**: Node types must be recognized by the executor

### Edge Validation

1. **Valid endpoints**: Source and target nodes must exist

2. **No self-loops**: Edges cannot connect a node to itself (except loops)

3. **Conditional labels**: If/else nodes must have labeled edges ("true"/"false")

4. **Unique labels**: Labels from the same source must be unique

### Expression Validation

1. **Syntax**: Expressions must be valid Expr syntax

2. **Balanced brackets**: Parentheses, brackets, and braces must be balanced

3. **Valid paths**: Variable paths should follow the format:
   - `trigger.*`
   - `steps.{node-id}.*`
   - `env.*`

### Runtime Validation

1. **Timeout enforcement**: HTTP requests and code execution have timeouts

2. **Retry limits**: Failed nodes retry up to configured max retries

3. **Loop limits**: Loops have a max iteration limit (default: 1000)

4. **Sub-workflow depth**: Maximum nesting depth of 10 levels

---

## 7. Workflow Variables

### Available Context

Every node execution has access to a context object:

```json
{
  "trigger": { /* trigger data */ },
  "steps": { /* outputs from completed nodes */ },
  "env": { /* environment variables */ }
}
```

### Trigger Data

Trigger data varies by trigger type:

**Webhook:**

```json
{
  "method": "POST",
  "path": "/webhook/github",
  "headers": { "Content-Type": "application/json" },
  "body": { /* parsed JSON body */ },
  "query": { "param": "value" }
}
```

**Schedule:**

```json
{
  "type": "schedule",
  "timestamp": "2026-01-01T09:00:00Z",
  "cron": "0 9 * * *"
}
```

### Step Outputs

Step outputs are stored by node ID:

```json
{
  "steps": {
    "http-1": {
      "status_code": 200,
      "headers": { /* headers */ },
      "body": { /* response body */ }
    },
    "transform-1": {
      "user_id": "12345",
      "user_name": "John Doe"
    }
  }
}
```

**Accessing:**

```
{{steps.http-1.body.user.email}}
{{steps.transform-1.user_name}}
```

### Environment Variables

Environment variables provide workflow metadata:

```json
{
  "env": {
    "tenant_id": "tenant-abc",
    "execution_id": "exec-12345",
    "workflow_id": "wf-67890"
  }
}
```

**Accessing:**

```
{{env.tenant_id}}
{{env.execution_id}}
```

### Credential References

Credentials are referenced using template syntax:

```
{{credentials.api_token}}
{{credentials.database_password}}
```

**Security:**
- Credentials are automatically decrypted and injected at runtime
- Credential values are masked in logs and outputs
- Access is logged for audit trails

---

## 8. Execution Model

### Execution Flow

1. **Create execution record** in database (status: pending)
2. **Load workflow definition** from the workflows table
3. **Parse nodes and edges** into data structures
4. **Validate graph structure** (no cycles, valid connections)
5. **Perform topological sort** to determine execution order
6. **Update status to running**
7. **Execute nodes sequentially** in topological order
   - Inject credentials if needed
   - Build input context (trigger, steps, env)
   - Execute node logic
   - Store output in step context
   - Broadcast progress events
8. **Mark execution as completed or failed**

### Node Traversal Order

Gorax uses **topological sort** (Kahn's algorithm) to determine execution order:

1. Build adjacency list and in-degree map
2. Find nodes with zero in-degree (start nodes)
3. Process nodes in order, decrementing in-degree of neighbors
4. Add neighbors with zero in-degree to the queue
5. If all nodes are processed, the sort succeeds; otherwise, there's a cycle

**Example:**

```
[Trigger] → [HTTP-1] → [Transform] → [HTTP-2]
                  ↓
              [Parallel]
```

**Execution order:** `trigger`, `http-1`, `parallel`, `transform`, `http-2`

### Parallel Execution

Parallel nodes execute branches concurrently using goroutines:

1. Create a goroutine for each branch
2. Use semaphore for concurrency control (if `max_concurrency` is set)
3. Collect results in a thread-safe manner
4. Wait for all branches to complete (or fail fast on error)

**Error Strategies:**
- `fail_fast`: Cancel all branches on first error
- `wait_all`: Wait for all branches, then report errors

### Error Handling

Each node execution can fail due to:
- Network errors (HTTP timeouts, connection refused)
- Validation errors (invalid config, missing fields)
- Expression errors (syntax errors, missing paths)
- External errors (API returned 500, auth failed)

**Retry Logic:**

```json
{
  "retry": {
    "enabled": true,
    "max_retries": 3,
    "initial_backoff_ms": 1000,
    "max_backoff_ms": 30000,
    "backoff_multiplier": 2.0
  }
}
```

**Error Classification:**
- **Transient errors**: Retryable (network timeouts, 503 errors)
- **Permanent errors**: Not retryable (400 errors, validation failures)

**Circuit Breaker:**
- After N consecutive failures, the circuit opens
- Subsequent requests fail immediately
- After a timeout, the circuit enters half-open state
- One successful request closes the circuit

### Timeout Behavior

Each node has a timeout:

1. **HTTP actions**: Default 30s, configurable
2. **Code execution**: Default 30s, max 60s
3. **Sub-workflows**: Configurable, default: no timeout

If a node times out:
- The execution fails with a timeout error
- Retry logic may retry the operation
- Broadcast timeout event for monitoring

### Sub-Workflow Execution

Sub-workflows are executed as nested workflow runs:

1. **Create sub-execution record** with parent_execution_id
2. **Increment execution depth** (max: 10)
3. **Detect circular dependencies** using workflow chain
4. **Map input** from parent context to sub-workflow trigger
5. **Execute sub-workflow** synchronously or asynchronously
6. **Map output** back to parent context
7. **Update parent execution** with sub-workflow result

**Depth Tracking:**

```
Parent Workflow (depth: 0)
  └─ Sub-Workflow A (depth: 1)
       └─ Sub-Workflow B (depth: 2)
            └─ Sub-Workflow C (depth: 3)
```

---

## 9. Examples

### 9.1 Simple Workflow: Webhook → HTTP Request

**Use Case:** Receive a webhook and make an HTTP request to an external API.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "position": { "x": 100, "y": 100 },
      "data": {
        "name": "GitHub Webhook",
        "config": {
          "path": "/webhook/github",
          "auth_type": "signature",
          "secret": "{{credentials.github_secret}}"
        }
      }
    },
    {
      "id": "http-1",
      "type": "action:http",
      "position": { "x": 300, "y": 100 },
      "data": {
        "name": "Notify Slack",
        "config": {
          "method": "POST",
          "url": "https://slack.com/api/chat.postMessage",
          "headers": {
            "Authorization": "Bearer {{credentials.slack_token}}",
            "Content-Type": "application/json"
          },
          "body": {
            "channel": "#deployments",
            "text": "New commit: {{trigger.body.commits[0].message}}"
          }
        }
      }
    }
  ],
  "edges": [
    {
      "id": "edge-1",
      "source": "trigger-1",
      "target": "http-1"
    }
  ]
}
```

---

### 9.2 Conditional Workflow: If/Else Logic

**Use Case:** Check if a user is active, send different notifications.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "position": { "x": 100, "y": 100 },
      "data": {
        "name": "User Webhook",
        "config": {}
      }
    },
    {
      "id": "http-1",
      "type": "action:http",
      "position": { "x": 300, "y": 100 },
      "data": {
        "name": "Get User",
        "config": {
          "method": "GET",
          "url": "https://api.example.com/users/{{trigger.body.user_id}}"
        }
      }
    },
    {
      "id": "conditional-1",
      "type": "control:if",
      "position": { "x": 500, "y": 100 },
      "data": {
        "name": "Check If Active",
        "config": {
          "condition": "steps['http-1'].body.status == 'active'"
        }
      }
    },
    {
      "id": "http-2-active",
      "type": "action:http",
      "position": { "x": 700, "y": 50 },
      "data": {
        "name": "Send Welcome Email",
        "config": {
          "method": "POST",
          "url": "https://api.sendgrid.com/v3/mail/send",
          "body": {
            "to": "{{steps['http-1'].body.email}}",
            "subject": "Welcome!"
          }
        }
      }
    },
    {
      "id": "http-2-inactive",
      "type": "action:http",
      "position": { "x": 700, "y": 150 },
      "data": {
        "name": "Send Re-activation Email",
        "config": {
          "method": "POST",
          "url": "https://api.sendgrid.com/v3/mail/send",
          "body": {
            "to": "{{steps['http-1'].body.email}}",
            "subject": "Come back!"
          }
        }
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "trigger-1", "target": "http-1" },
    { "id": "e2", "source": "http-1", "target": "conditional-1" },
    { "id": "e3", "source": "conditional-1", "target": "http-2-active", "label": "true" },
    { "id": "e4", "source": "conditional-1", "target": "http-2-inactive", "label": "false" }
  ]
}
```

---

### 9.3 Loop Workflow: Iterate Over Array

**Use Case:** Fetch a list of users and send an email to each one.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:schedule",
      "position": { "x": 100, "y": 100 },
      "data": {
        "name": "Daily Digest",
        "config": {
          "cron": "0 9 * * *",
          "timezone": "America/New_York"
        }
      }
    },
    {
      "id": "http-1",
      "type": "action:http",
      "position": { "x": 300, "y": 100 },
      "data": {
        "name": "Fetch Users",
        "config": {
          "method": "GET",
          "url": "https://api.example.com/users?active=true"
        }
      }
    },
    {
      "id": "loop-1",
      "type": "control:loop",
      "position": { "x": 500, "y": 100 },
      "data": {
        "name": "For Each User",
        "config": {
          "source": "steps['http-1'].body.users",
          "item_variable": "user",
          "index_variable": "i",
          "max_iterations": 1000,
          "on_error": "continue"
        }
      }
    },
    {
      "id": "http-2",
      "type": "action:http",
      "position": { "x": 700, "y": 100 },
      "data": {
        "name": "Send Email",
        "config": {
          "method": "POST",
          "url": "https://api.sendgrid.com/v3/mail/send",
          "headers": {
            "Authorization": "Bearer {{credentials.sendgrid_token}}"
          },
          "body": {
            "personalizations": [
              {
                "to": [{ "email": "{{steps.user.email}}" }],
                "subject": "Daily Digest for {{steps.user.name}}"
              }
            ],
            "from": { "email": "noreply@example.com" },
            "content": [
              {
                "type": "text/html",
                "value": "<p>Hello {{steps.user.name}}, here is your daily digest.</p>"
              }
            ]
          }
        }
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "trigger-1", "target": "http-1" },
    { "id": "e2", "source": "http-1", "target": "loop-1" },
    { "id": "e3", "source": "loop-1", "target": "http-2" }
  ]
}
```

---

### 9.4 Parallel Workflow: Fan-out/Fan-in

**Use Case:** Call multiple APIs in parallel, then process all results.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "position": { "x": 100, "y": 200 },
      "data": {
        "name": "Trigger",
        "config": {}
      }
    },
    {
      "id": "parallel-1",
      "type": "control:parallel",
      "position": { "x": 300, "y": 200 },
      "data": {
        "name": "Fetch Data in Parallel",
        "config": {
          "error_strategy": "wait_all",
          "max_concurrency": 0
        }
      }
    },
    {
      "id": "http-1",
      "type": "action:http",
      "position": { "x": 500, "y": 100 },
      "data": {
        "name": "Fetch Users",
        "config": {
          "method": "GET",
          "url": "https://api.example.com/users"
        }
      }
    },
    {
      "id": "http-2",
      "type": "action:http",
      "position": { "x": 500, "y": 200 },
      "data": {
        "name": "Fetch Orders",
        "config": {
          "method": "GET",
          "url": "https://api.example.com/orders"
        }
      }
    },
    {
      "id": "http-3",
      "type": "action:http",
      "position": { "x": 500, "y": 300 },
      "data": {
        "name": "Fetch Products",
        "config": {
          "method": "GET",
          "url": "https://api.example.com/products"
        }
      }
    },
    {
      "id": "transform-1",
      "type": "action:transform",
      "position": { "x": 700, "y": 200 },
      "data": {
        "name": "Combine Results",
        "config": {
          "mapping": {
            "user_count": "steps['http-1'].body.count",
            "order_count": "steps['http-2'].body.count",
            "product_count": "steps['http-3'].body.count"
          }
        }
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "trigger-1", "target": "parallel-1" },
    { "id": "e2", "source": "parallel-1", "target": "http-1" },
    { "id": "e3", "source": "parallel-1", "target": "http-2" },
    { "id": "e4", "source": "parallel-1", "target": "http-3" },
    { "id": "e5", "source": "http-1", "target": "transform-1" },
    { "id": "e6", "source": "http-2", "target": "transform-1" },
    { "id": "e7", "source": "http-3", "target": "transform-1" }
  ]
}
```

---

### 9.5 Complex Workflow: Combining Multiple Patterns

**Use Case:** E-commerce order processing with conditionals, loops, and parallel execution.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "position": { "x": 100, "y": 200 },
      "data": {
        "name": "New Order Webhook",
        "config": {
          "path": "/webhook/orders",
          "auth_type": "api_key"
        }
      }
    },
    {
      "id": "http-1",
      "type": "action:http",
      "position": { "x": 300, "y": 200 },
      "data": {
        "name": "Fetch Order Details",
        "config": {
          "method": "GET",
          "url": "https://api.example.com/orders/{{trigger.body.order_id}}"
        }
      }
    },
    {
      "id": "conditional-1",
      "type": "control:if",
      "position": { "x": 500, "y": 200 },
      "data": {
        "name": "Check Order Total",
        "config": {
          "condition": "steps['http-1'].body.total > 100"
        }
      }
    },
    {
      "id": "http-2-high",
      "type": "action:http",
      "position": { "x": 700, "y": 100 },
      "data": {
        "name": "Apply Discount",
        "config": {
          "method": "POST",
          "url": "https://api.example.com/orders/{{trigger.body.order_id}}/discount",
          "body": {
            "discount_percent": 10
          }
        }
      }
    },
    {
      "id": "parallel-1",
      "type": "control:parallel",
      "position": { "x": 900, "y": 200 },
      "data": {
        "name": "Process Order",
        "config": {
          "error_strategy": "fail_fast",
          "max_concurrency": 3
        }
      }
    },
    {
      "id": "http-3",
      "type": "action:http",
      "position": { "x": 1100, "y": 100 },
      "data": {
        "name": "Charge Payment",
        "config": {
          "method": "POST",
          "url": "https://api.stripe.com/v1/charges",
          "headers": {
            "Authorization": "Bearer {{credentials.stripe_key}}"
          },
          "body": {
            "amount": "{{steps['http-1'].body.total}}",
            "currency": "usd",
            "customer": "{{steps['http-1'].body.customer_id}}"
          }
        }
      }
    },
    {
      "id": "http-4",
      "type": "action:http",
      "position": { "x": 1100, "y": 200 },
      "data": {
        "name": "Update Inventory",
        "config": {
          "method": "POST",
          "url": "https://api.example.com/inventory/decrement",
          "body": {
            "items": "{{steps['http-1'].body.items}}"
          }
        }
      }
    },
    {
      "id": "loop-1",
      "type": "control:loop",
      "position": { "x": 1100, "y": 300 },
      "data": {
        "name": "Send Notifications",
        "config": {
          "source": "steps['http-1'].body.items",
          "item_variable": "item",
          "on_error": "continue"
        }
      }
    },
    {
      "id": "slack-1",
      "type": "slack:send_message",
      "position": { "x": 1300, "y": 300 },
      "data": {
        "name": "Notify Team",
        "config": {
          "credential_id": "cred-slack",
          "channel": "#orders",
          "text": "Item shipped: {{steps.item.name}}"
        }
      }
    },
    {
      "id": "http-5",
      "type": "action:http",
      "position": { "x": 1300, "y": 200 },
      "data": {
        "name": "Send Confirmation Email",
        "config": {
          "method": "POST",
          "url": "https://api.sendgrid.com/v3/mail/send",
          "body": {
            "to": "{{steps['http-1'].body.customer_email}}",
            "subject": "Order Confirmed",
            "html": "<p>Your order has been confirmed!</p>"
          }
        }
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "trigger-1", "target": "http-1" },
    { "id": "e2", "source": "http-1", "target": "conditional-1" },
    { "id": "e3", "source": "conditional-1", "target": "http-2-high", "label": "true" },
    { "id": "e4", "source": "conditional-1", "target": "parallel-1", "label": "false" },
    { "id": "e5", "source": "http-2-high", "target": "parallel-1" },
    { "id": "e6", "source": "parallel-1", "target": "http-3" },
    { "id": "e7", "source": "parallel-1", "target": "http-4" },
    { "id": "e8", "source": "parallel-1", "target": "loop-1" },
    { "id": "e9", "source": "loop-1", "target": "slack-1" },
    { "id": "e10", "source": "http-3", "target": "http-5" },
    { "id": "e11", "source": "http-4", "target": "http-5" },
    { "id": "e12", "source": "slack-1", "target": "http-5" }
  ]
}
```

---

### 9.6 Sub-Workflow Example

**Parent Workflow:**

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "position": { "x": 100, "y": 100 },
      "data": {
        "name": "Order Webhook",
        "config": {}
      }
    },
    {
      "id": "sub-workflow-1",
      "type": "control:sub_workflow",
      "position": { "x": 300, "y": 100 },
      "data": {
        "name": "Process Payment",
        "config": {
          "workflow_id": "wf-payment-processing",
          "input_mapping": {
            "order_id": "trigger.body.order_id",
            "amount": "trigger.body.total",
            "customer_id": "trigger.body.customer_id"
          },
          "output_mapping": {
            "payment_status": "output.status",
            "transaction_id": "output.transaction_id"
          },
          "wait_for_result": true,
          "timeout_ms": 30000
        }
      }
    },
    {
      "id": "http-1",
      "type": "action:http",
      "position": { "x": 500, "y": 100 },
      "data": {
        "name": "Update Order Status",
        "config": {
          "method": "PATCH",
          "url": "https://api.example.com/orders/{{trigger.body.order_id}}",
          "body": {
            "payment_status": "{{steps['sub-workflow-1'].payment_status}}",
            "transaction_id": "{{steps['sub-workflow-1'].transaction_id}}"
          }
        }
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "trigger-1", "target": "sub-workflow-1" },
    { "id": "e2", "source": "sub-workflow-1", "target": "http-1" }
  ]
}
```

**Sub-Workflow (Payment Processing):**

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "position": { "x": 100, "y": 100 },
      "data": {
        "name": "Sub-workflow Trigger",
        "config": {}
      }
    },
    {
      "id": "http-1",
      "type": "action:http",
      "position": { "x": 300, "y": 100 },
      "data": {
        "name": "Charge Card",
        "config": {
          "method": "POST",
          "url": "https://api.stripe.com/v1/charges",
          "headers": {
            "Authorization": "Bearer {{credentials.stripe_key}}"
          },
          "body": {
            "amount": "{{trigger.body.amount}}",
            "currency": "usd",
            "customer": "{{trigger.body.customer_id}}"
          }
        }
      }
    },
    {
      "id": "transform-1",
      "type": "action:transform",
      "position": { "x": 500, "y": 100 },
      "data": {
        "name": "Format Output",
        "config": {
          "mapping": {
            "status": "steps['http-1'].body.status",
            "transaction_id": "steps['http-1'].body.id"
          }
        }
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "trigger-1", "target": "http-1" },
    { "id": "e2", "source": "http-1", "target": "transform-1" }
  ]
}
```

---

## 10. Best Practices

### 10.1 Workflow Design Patterns

#### Single Responsibility

Each workflow should have a clear, single purpose:

- **Good:** "Process new customer signup"
- **Bad:** "Process signups, send emails, update CRM, and generate reports"

Break complex processes into multiple workflows connected by sub-workflows.

#### Idempotency

Design workflows to be idempotent (safe to run multiple times):

- Use unique identifiers in API calls
- Check for existing records before creating
- Use HTTP PUT/PATCH instead of POST when possible

#### Error Recovery

Plan for failures:

- Use retry logic for transient errors
- Add conditional branches to handle error states
- Log failures for debugging
- Send notifications for critical failures

### 10.2 Error Handling Strategies

#### Retry Configuration

Configure retries for transient errors:

```json
{
  "retry": {
    "enabled": true,
    "max_retries": 3,
    "initial_backoff_ms": 1000,
    "max_backoff_ms": 30000,
    "backoff_multiplier": 2.0
  }
}
```

**When to retry:**
- Network timeouts
- 5xx server errors
- Rate limit errors (429)

**When NOT to retry:**
- 4xx client errors (except 429)
- Validation failures
- Authentication errors

#### Circuit Breakers

Use circuit breakers for external dependencies:

- After N consecutive failures, stop making requests
- Fail fast to avoid cascading failures
- Automatically recover after timeout period

#### Graceful Degradation

Handle failures gracefully:

```json
{
  "type": "control:if",
  "data": {
    "config": {
      "condition": "steps['http-1'].error == null",
      "true_branch": "true",
      "false_branch": "false"
    }
  }
}
```

### 10.3 Performance Optimization

#### Parallel Execution

Use parallel nodes for independent operations:

```json
{
  "type": "control:parallel",
  "data": {
    "config": {
      "error_strategy": "wait_all",
      "max_concurrency": 5
    }
  }
}
```

**Benefits:**
- Reduced total execution time
- Better resource utilization
- Improved throughput

#### Caching

Cache expensive operations:

- Store results in workflow context
- Reuse data across multiple nodes
- Avoid redundant API calls

#### Timeouts

Set appropriate timeouts:

- Short timeouts for fast APIs (5-10s)
- Long timeouts for slow operations (30-60s)
- No timeout for async operations

### 10.4 Testing Workflows

#### Dry-Run Testing

Use the dry-run API to validate workflows:

```bash
POST /api/v1/workflows/{id}/dry-run
{
  "test_data": {
    "trigger": {
      "body": { "user_id": "12345" }
    }
  }
}
```

**Validates:**
- Graph structure (no cycles)
- Node configuration (required fields)
- Expression syntax
- Variable references

#### Unit Testing

Test individual nodes in isolation:

- Mock external APIs
- Use test credentials
- Verify outputs match expectations

#### Integration Testing

Test complete workflows end-to-end:

- Use staging environment
- Test with real data
- Verify side effects (emails sent, records created)

### 10.5 Versioning Workflows

#### Semantic Versioning

Version workflows semantically:

- **Major:** Breaking changes (remove nodes, change outputs)
- **Minor:** New features (add nodes, new fields)
- **Patch:** Bug fixes (fix expressions, update URLs)

#### Backward Compatibility

Maintain backward compatibility:

- Don't remove fields from outputs
- Add optional fields instead of required
- Deprecate old fields before removing

#### Version History

Gorax automatically tracks workflow versions:

- Each update creates a new version
- Executions reference the version used
- Roll back to previous versions if needed

### 10.6 Security Best Practices

#### Credentials Management

- **Never hardcode secrets** in workflow definitions
- Use credential references: `{{credentials.api_token}}`
- Rotate credentials regularly
- Audit credential access logs

#### Input Validation

Validate all external inputs:

```json
{
  "type": "control:if",
  "data": {
    "config": {
      "condition": "trigger.body.email contains '@' && len(trigger.body.email) < 255"
    }
  }
}
```

#### Least Privilege

- Use dedicated credentials per integration
- Grant minimal required permissions
- Revoke unused credentials

#### Data Masking

Sensitive data is automatically masked:

- Credential values are masked in logs
- Redacted in execution history
- Encrypted at rest

---

## Appendix A: Complete Node Type Reference

| Node Type | Category | Description |
|-----------|----------|-------------|
| `trigger:webhook` | Trigger | Receive HTTP webhooks |
| `trigger:schedule` | Trigger | Schedule-based execution |
| `action:http` | Action | HTTP requests |
| `action:transform` | Action | Data transformation |
| `action:formula` | Action | Mathematical expressions |
| `action:code` | Action | JavaScript execution |
| `action:email` | Action | Send emails |
| `slack:send_message` | Integration | Slack message |
| `slack:send_dm` | Integration | Slack DM |
| `slack:update_message` | Integration | Update Slack message |
| `slack:add_reaction` | Integration | Add Slack reaction |
| `control:if` | Control Flow | Conditional branching |
| `control:loop` | Control Flow | Iterate over array |
| `control:parallel` | Control Flow | Parallel execution |
| `control:fork` | Control Flow | Fork into branches |
| `control:join` | Control Flow | Join branches |
| `control:delay` | Control Flow | Pause execution |
| `control:sub_workflow` | Control Flow | Execute sub-workflow |

---

## Appendix B: Expression Operators

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal | `a == b` |
| `!=` | Not equal | `a != b` |
| `<` | Less than | `a < b` |
| `<=` | Less than or equal | `a <= b` |
| `>` | Greater than | `a > b` |
| `>=` | Greater than or equal | `a >= b` |

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `&&` | Logical AND | `a && b` |
| `||` | Logical OR | `a || b` |
| `!` | Logical NOT | `!a` |

### Arithmetic Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `a + b` |
| `-` | Subtraction | `a - b` |
| `*` | Multiplication | `a * b` |
| `/` | Division | `a / b` |
| `%` | Modulus | `a % b` |
| `**` | Exponentiation | `a ** b` |

### String Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Concatenation | `'Hello' + ' World'` |
| `contains` | Contains substring | `str contains 'test'` |
| `startsWith` | Starts with prefix | `str startsWith 'Hello'` |
| `endsWith` | Ends with suffix | `str endsWith '.com'` |

---

## Appendix C: HTTP Status Code Handling

### Success Codes (2xx)

| Code | Meaning | Retry? |
|------|---------|--------|
| 200 | OK | No |
| 201 | Created | No |
| 202 | Accepted | No |
| 204 | No Content | No |

### Client Errors (4xx)

| Code | Meaning | Retry? |
|------|---------|--------|
| 400 | Bad Request | No |
| 401 | Unauthorized | No |
| 403 | Forbidden | No |
| 404 | Not Found | No |
| 429 | Too Many Requests | Yes (with backoff) |

### Server Errors (5xx)

| Code | Meaning | Retry? |
|------|---------|--------|
| 500 | Internal Server Error | Yes |
| 502 | Bad Gateway | Yes |
| 503 | Service Unavailable | Yes |
| 504 | Gateway Timeout | Yes |

---

## Appendix D: Cron Expression Reference

### Format

```
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-6, Sunday=0)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

### Special Characters

| Char | Description | Example |
|------|-------------|---------|
| `*` | Any value | `* * * * *` (every minute) |
| `,` | Value list | `0,15,30,45 * * * *` (every 15 min) |
| `-` | Range | `0 9-17 * * *` (9 AM to 5 PM) |
| `/` | Step values | `*/5 * * * *` (every 5 minutes) |

### Common Examples

| Expression | Description |
|------------|-------------|
| `0 0 * * *` | Daily at midnight |
| `0 */6 * * *` | Every 6 hours |
| `0 9 * * 1-5` | Weekdays at 9 AM |
| `0 0 1 * *` | First of every month |
| `0 0 * * 0` | Every Sunday at midnight |

---

**End of Document**
