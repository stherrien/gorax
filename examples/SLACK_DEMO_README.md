# Slack Integration Demo Workflow

This demo workflow showcases all four Slack actions available in Gorax:

## What It Does

This workflow demonstrates a complete notification flow:

1. **Webhook Trigger** - Receives an HTTP POST request to start the workflow
2. **Send Initial Message** - Posts a message to a Slack channel announcing the workflow has started
3. **Add Eyes Reaction** - Adds 👀 emoji to indicate the message is being processed
4. **Process Data** - Simulates data processing with an HTTP request
5. **Update Message** - Updates the original Slack message with completion status
6. **Add Success Reaction** - Adds ✅ emoji to indicate success
7. **Send DM** - Notifies an admin via direct message about the completion

## Workflow Diagram

```
Webhook
   ↓
Send Message (🚀 Processing...)
   ↓
Add Reaction (👀)
   ↓
HTTP Process ──────┐
   ↓               ↓
Update Message   Send DM
(✅ Complete)    (Notify Admin)
   ↓
Add Reaction (✅)
```

## Features Demonstrated

### 1. **Send Message** (`slack:send_message`)
- Posts messages to public/private channels
- Supports plain text and Slack Block Kit
- Custom bot username and icon
- Template variable interpolation

### 2. **Add Reaction** (`slack:add_reaction`)
- Adds emoji reactions to messages
- References previous step outputs using `{{steps.slack-send-1.ts}}`
- Useful for status indicators

### 3. **Update Message** (`slack:update_message`)
- Modifies existing messages
- Perfect for progress updates
- Updates both text and Block Kit layouts

### 4. **Send DM** (`slack:send_dm`)
- Sends direct messages to users
- Accepts email addresses or Slack user IDs
- Great for personal notifications

## Template Variables Used

The workflow demonstrates Gorax's template variable system:

- `{{steps.slack-send-1.channel}}` - Channel ID from previous step
- `{{steps.slack-send-1.ts}}` - Message timestamp for reactions/updates
- `{{trigger.timestamp}}` - When the workflow was triggered
- `{{workflow.name}}` - Current workflow name
- `{{execution.id}}` - Unique execution identifier
- `{{steps.http-process.duration}}` - Processing duration

## Configuration Required

Before using this workflow, update these values:

### Channel ID
Replace `C1234567890` with your actual Slack channel ID:
- Open Slack → Right-click channel → View channel details
- Channel ID is at the bottom

### Admin Email
Replace `admin@example.com` with the actual user email or Slack user ID:
- For email: `user@company.com`
- For user ID: `U1234567890`

### Slack Credentials
Ensure you have configured Slack OAuth credentials in Gorax:
1. Create a Slack app at https://api.slack.com/apps
2. Add OAuth scopes:
   - `chat:write` - Send messages
   - `chat:write.public` - Post to public channels
   - `reactions:write` - Add reactions
   - `users:read` - Look up users by email
   - `users:read.email` - Read user email addresses
3. Install app to workspace
4. Store OAuth token in Gorax credential service

## Testing the Workflow

### 1. Import the workflow:
```bash
curl -X POST http://localhost:8080/api/workflows \
  -H "Content-Type: application/json" \
  -d @slack-notification-workflow.json
```

### 2. Trigger via webhook:
```bash
curl -X POST http://localhost:8080/webhooks/slack-demo \
  -H "Content-Type: application/json" \
  -d '{"message": "Test notification"}'
```

### 3. Watch the Slack channel:
You should see:
1. Initial message: "🚀 Workflow started! Processing your request..."
2. Eyes reaction (👀) added
3. Message updated: "✅ Workflow completed successfully!"
4. Checkmark reaction (✅) added
5. DM sent to admin

## Slack Block Kit Examples

The workflow uses Slack's Block Kit for rich formatting. Here are the blocks used:

### Initial Message Block
```json
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*Workflow Status*\n🚀 Processing your request..."
    }
  }
]
```

### Completion Message Block
```json
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*Workflow Status*\n✅ Completed successfully!\n\n*Details:*\n• Started: {{trigger.timestamp}}\n• Duration: {{steps.http-process.duration}}ms\n• Status: Success"
    }
  }
]
```

### DM Notification Block
```json
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*Workflow Notification*\n\nThe workflow `{{workflow.name}}` has completed successfully.\n\n*Execution ID:* {{execution.id}}\n*Status:* ✅ Success"
    }
  }
]
```

## Common Use Cases

This pattern can be adapted for:

### 1. **Deployment Notifications**
- Send message when deployment starts
- Update with progress
- Add reactions for status (🚀 → ✅ or ❌)
- Notify team via DM on failure

### 2. **CI/CD Pipeline Updates**
- Initial notification on pipeline start
- Update message with test results
- Add reactions for pass/fail
- DM relevant team members

### 3. **Support Ticket Workflows**
- Notify channel of new ticket
- Update with assignment status
- Add reactions for priority
- DM assigned agent

### 4. **Data Processing Jobs**
- Announce job start
- Update with progress milestones
- Show completion status
- Alert on errors

## Error Handling

The workflow includes proper error handling:

- If a Slack action fails, the workflow stops
- Error details are captured in execution logs
- Use conditional nodes (future) to handle failures
- Retry configuration can be added to each node

## Next Steps

To extend this workflow:

1. **Add Conditionals**: Branch based on processing results
2. **Add Loops**: Process multiple items with updates
3. **Error Notifications**: Send DM on failure
4. **Threading**: Use `thread_ts` for grouped updates
5. **Interactive Buttons**: Add Block Kit buttons for user actions

## Documentation

For more details on Slack actions:
- [Slack API Documentation](https://api.slack.com/)
- [Block Kit Builder](https://app.slack.com/block-kit-builder)
- [Gorax Documentation](../docs/integrations/slack.md)

## Support

If you encounter issues:
1. Check Slack app permissions
2. Verify credential configuration
3. Review execution logs in Gorax
4. Test Slack API calls directly
