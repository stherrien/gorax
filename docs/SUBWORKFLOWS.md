# Subworkflow Actions in Gorax

## Overview

Subworkflows enable workflow composition and reusability by allowing one workflow to call another workflow as an action. This powerful feature promotes:

- **Modularity**: Break complex workflows into smaller, reusable components
- **Maintainability**: Update a workflow once and all parent workflows benefit
- **Reusability**: Create library workflows for common tasks (validation, notifications, etc.)
- **Organization**: Separate concerns and keep workflows focused

## Features

- **Synchronous & Asynchronous Execution**: Choose to wait for completion or fire-and-forget
- **Input/Output Mapping**: Pass data between parent and child workflows
- **Context Inheritance**: Optionally inherit parent workflow context
- **Circular Dependency Detection**: Prevents infinite loops
- **Depth Limiting**: Maximum 10 levels of nesting to prevent stack overflow
- **Execution Tracking**: Full parent-child relationship tracking in database
- **Timeout Support**: Set maximum execution time for synchronous calls
- **Tenant Isolation**: Subworkflows respect multi-tenant boundaries

## Configuration

### Node Type

- **Action Node**: `action:subworkflow`
- **Control Node** (legacy): `control:sub_workflow`

### Configuration Schema

```json
{
  "workflow_id": "wf-12345",
  "workflow_name": "User Validation Workflow",
  "input_mapping": {
    "email": "${trigger.body.email}",
    "userId": "${trigger.body.user_id}"
  },
  "output_mapping": {
    "is_valid": "${output.validation_status}",
    "errors": "${output.validation_errors}"
  },
  "mode": "sync",
  "timeout": "30s",
  "inherit_context": false
}
```

### Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `workflow_id` | string | ✅ | ID of the workflow to execute |
| `workflow_name` | string | ❌ | Display name (auto-fetched if not provided) |
| `input_mapping` | object | ❌ | Map parent context → child input |
| `output_mapping` | object | ❌ | Map child output → parent context |
| `mode` | string | ✅ | `"sync"` or `"async"` |
| `timeout` | string | ❌ | Timeout duration (e.g., "30s", "5m", "1h") |
| `inherit_context` | boolean | ❌ | Whether to pass parent context to child (default: false) |

## Usage Examples

### Example 1: Simple Synchronous Call

```json
{
  "type": "action:subworkflow",
  "config": {
    "workflow_id": "wf-notify-admin",
    "workflow_name": "Admin Notification",
    "mode": "sync",
    "timeout": "10s"
  }
}
```

### Example 2: With Input/Output Mapping

```json
{
  "type": "action:subworkflow",
  "config": {
    "workflow_id": "wf-validate-user",
    "workflow_name": "User Validation",
    "input_mapping": {
      "email": "${trigger.body.email}",
      "name": "${trigger.body.name}",
      "accountType": "${steps.getUserType.output.type}"
    },
    "output_mapping": {
      "is_valid": "${output.valid}",
      "validation_message": "${output.message}"
    },
    "mode": "sync",
    "timeout": "30s"
  }
}
```

### Example 3: Asynchronous Fire-and-Forget

```json
{
  "type": "action:subworkflow",
  "config": {
    "workflow_id": "wf-send-welcome-email",
    "workflow_name": "Send Welcome Email",
    "input_mapping": {
      "email": "${trigger.body.email}",
      "name": "${trigger.body.name}"
    },
    "mode": "async"
  }
}
```

### Example 4: Context Inheritance

```json
{
  "type": "action:subworkflow",
  "config": {
    "workflow_id": "wf-audit-log",
    "workflow_name": "Audit Log Workflow",
    "inherit_context": true,
    "mode": "async"
  }
}
```

## Common Use Cases

### 1. Reusable Validation Workflows

Create a shared validation workflow that multiple workflows can call:

**Validation Workflow** (`wf-validate-user`):
- Input: email, name, phone
- Logic: Check format, verify email domain, validate phone
- Output: valid (boolean), errors (array)

**Parent Workflows**:
- Registration workflow → calls validation
- Profile update workflow → calls validation
- Admin user creation → calls validation

### 2. Multi-Step Approval Process

Break approval into stages:

**Workflow Chain**:
1. Main workflow → Manager Approval workflow
2. Manager Approval → Director Approval workflow (if > $10k)
3. Director Approval → Finance Review workflow (if > $50k)

### 3. Notification Workflows

Create notification workflows for different channels:

- **Email Notification Workflow**: Formats and sends emails
- **Slack Notification Workflow**: Posts to Slack channels
- **SMS Notification Workflow**: Sends text messages

Parent workflows call the appropriate notification workflow based on user preferences.

### 4. Data Transformation Pipelines

Chain data transformation workflows:

1. **Extract workflow**: Fetch data from source
2. **Transform workflow**: Clean and format data
3. **Load workflow**: Store in destination

### 5. Error Handling Workflows

Create dedicated error handling workflows:

```json
{
  "type": "control:try",
  "config": {
    "try_nodes": ["risky-operation"],
    "catch_nodes": ["error-handler"]
  }
}

// error-handler node calls error-notification-workflow
{
  "type": "action:subworkflow",
  "config": {
    "workflow_id": "wf-error-notification",
    "input_mapping": {
      "error": "${error.message}",
      "workflow": "${env.workflow_id}",
      "execution": "${env.execution_id}"
    },
    "mode": "async"
  }
}
```

## Execution Flow

### Synchronous Mode

1. Parent workflow reaches subworkflow node
2. Parent context is mapped to child input
3. Child workflow execution created (status: pending)
4. Child workflow executes
5. Parent workflow waits (up to timeout)
6. Child completes
7. Child output is mapped to parent context
8. Parent workflow continues

### Asynchronous Mode

1. Parent workflow reaches subworkflow node
2. Parent context is mapped to child input
3. Child workflow execution created (status: pending)
4. Child workflow scheduled for background execution
5. Parent workflow immediately continues with execution ID
6. Child workflow executes independently

## Input/Output Mapping

### Expression Syntax

Use `${}` syntax to reference values:

- `${trigger.body.field}` - Access trigger data
- `${steps.stepId.output.field}` - Access previous step output
- `${env.tenant_id}` - Access environment variables
- `"literal value"` - Pass literal strings/numbers

### Input Mapping Examples

```json
{
  "input_mapping": {
    // Simple field
    "email": "${trigger.body.email}",

    // Nested field
    "userId": "${trigger.body.user.id}",

    // From previous step
    "accountType": "${steps.getUserType.output.type}",

    // Literal value
    "source": "api",

    // Number
    "priority": "1"
  }
}
```

### Output Mapping Examples

```json
{
  "output_mapping": {
    // Map single field
    "validation_result": "${output.valid}",

    // Map nested field
    "error_code": "${output.error.code}",

    // Map to different name
    "user_email": "${output.validated_email}"
  }
}
```

## Security & Best Practices

### Security

1. **Tenant Isolation**: Subworkflows can only call workflows in the same tenant
2. **Workflow Status**: Only `active` workflows can be called
3. **Permission Checking**: User must have execute permission on target workflow
4. **Depth Limiting**: Maximum 10 levels of nesting
5. **Circular Detection**: Prevents workflow A → B → A cycles

### Best Practices

1. **Keep Workflows Focused**: Each workflow should have a single responsibility
2. **Use Descriptive Names**: Name workflows clearly (e.g., "Send Welcome Email", not "Email Flow 1")
3. **Document Inputs/Outputs**: Add descriptions to workflow definitions
4. **Set Appropriate Timeouts**: Consider worst-case execution time
5. **Handle Errors**: Use try/catch for subworkflow calls that might fail
6. **Avoid Deep Nesting**: More than 3-4 levels becomes hard to debug
7. **Use Async for Long Operations**: Don't block parent workflow for slow tasks
8. **Version Workflows**: Use workflow versions to manage breaking changes

## Error Handling

### Common Errors

| Error | Description | Solution |
|-------|-------------|----------|
| Workflow not found | Invalid `workflow_id` | Verify workflow exists and is accessible |
| Workflow not active | Target workflow is draft/archived | Activate the workflow |
| Circular dependency | Workflow calls itself (directly or indirectly) | Redesign workflow chain |
| Max depth exceeded | Too many nested levels (>10) | Flatten workflow hierarchy |
| Timeout | Subworkflow exceeded timeout | Increase timeout or optimize child |
| Missing executor | Sync mode requires executor | Check configuration |

### Example Error Handling

```json
{
  "type": "control:try",
  "config": {
    "try_nodes": ["call-subworkflow"],
    "catch_nodes": ["handle-error"],
    "finally_nodes": ["cleanup"]
  }
}
```

## Monitoring & Debugging

### Execution Tracking

All subworkflow executions are tracked in the database:

```sql
SELECT
  e.id,
  e.workflow_id,
  e.parent_execution_id,
  e.execution_depth,
  e.status,
  e.created_at
FROM executions e
WHERE e.parent_execution_id = 'parent-exec-id';
```

### Execution Tree

View the full execution hierarchy:

```
Root Execution (depth=0)
├─ Child Execution 1 (depth=1)
│  └─ Grandchild Execution (depth=2)
└─ Child Execution 2 (depth=1)
```

### Debugging Tips

1. **Check Depth**: Verify execution depth in database
2. **Verify Mapping**: Log input/output values to verify mapping
3. **Test Independently**: Test child workflow independently first
4. **Use Sync Mode**: Start with sync mode for easier debugging
5. **Check Logs**: Review step execution logs for both parent and child
6. **Monitor Timeouts**: Track execution duration to set appropriate timeouts

## Performance Considerations

1. **Caching**: Workflow definitions are cached to reduce database load
2. **Async for Scale**: Use async mode for high-volume operations
3. **Timeout Tuning**: Set realistic timeouts to free up resources
4. **Depth Limiting**: Shallow hierarchies perform better
5. **Parallel vs Sequential**: Use parallel execution for independent subworkflows

## Limitations

- Maximum depth: 10 levels
- Maximum timeout: 10 minutes (configurable)
- Cannot call workflows across tenants
- Cannot call inactive workflows
- Async mode does not return output data

## Migration from Legacy Format

If you're using the legacy format, migrate to the new format:

### Old Format

```json
{
  "workflow_id": "wf-123",
  "wait_for_result": true,
  "timeout_ms": 30000
}
```

### New Format

```json
{
  "workflow_id": "wf-123",
  "mode": "sync",
  "timeout": "30s"
}
```

Both formats are supported for backward compatibility.

## API Reference

### Create Subworkflow Node (REST API)

```bash
POST /api/workflows/{workflow_id}/nodes
Content-Type: application/json

{
  "type": "action:subworkflow",
  "data": {
    "name": "Call Validation Workflow",
    "config": {
      "workflow_id": "wf-validate-user",
      "input_mapping": {
        "email": "${trigger.body.email}"
      },
      "output_mapping": {
        "is_valid": "${output.valid}"
      },
      "mode": "sync",
      "timeout": "30s"
    }
  },
  "position": {
    "x": 400,
    "y": 200
  }
}
```

### Query Subworkflow Executions

```bash
GET /api/executions?parent_execution_id={parent_id}
```

## Troubleshooting

### Subworkflow Not Executing

1. Verify workflow is active: `GET /api/workflows/{id}`
2. Check user permissions
3. Review logs for error messages
4. Verify input mapping expressions are valid

### Timeout Issues

1. Increase timeout value
2. Optimize child workflow performance
3. Consider using async mode
4. Check for blocking operations

### Circular Dependency Error

1. Review workflow chain
2. Break the cycle by restructuring
3. Use a shared data store instead

### Output Not Available

1. Verify using sync mode (async returns execution ID only)
2. Check output mapping expressions
3. Verify child workflow sets output data

## Support

For issues or questions:
- GitHub Issues: https://github.com/gorax/gorax/issues
- Documentation: https://docs.gorax.io
- Community Forum: https://community.gorax.io
