# Gorax Example Workflows

This directory contains example workflows demonstrating various Gorax features and integrations.

## Slack Integration Examples

### 1. Hello World (`slack-hello-world.json`)
**Difficulty**: Beginner
**Actions Used**: Send Message, Add Reaction

The simplest possible Slack workflow. Perfect for:
- Testing your Slack integration setup
- Learning the basic workflow structure
- Verifying credentials are configured correctly

**Flow**:
```
Webhook → Send "Hello" → Add 👋 Reaction
```

**Quick Start**:
```bash
# 1. Update channel ID in the JSON file
# 2. Import workflow
curl -X POST http://localhost:8080/api/workflows \
  -H "Content-Type: application/json" \
  -d @slack-hello-world.json

# 3. Trigger it
curl -X POST http://localhost:8080/webhooks/hello-slack
```

---

### 2. Notification Demo (`slack-notification-workflow.json`)
**Difficulty**: Intermediate
**Actions Used**: All Slack actions + HTTP + Template Variables

Comprehensive demo showing all 4 Slack actions working together:
- Send initial message
- Add reaction (👀 "watching")
- Process data via HTTP
- Update message with results
- Add success reaction (✅)
- Send DM to admin

**Flow**:
```
Webhook → Send Message → Add Reaction → HTTP Process
                                              ↓
                                         Update Message → Add Reaction
                                              ↓
                                           Send DM
```

**Features Demonstrated**:
- Template variable interpolation (`{{steps.slack-send-1.ts}}`)
- Referencing previous step outputs
- Slack Block Kit formatting
- Custom bot username and icon
- Message threading capability

See [SLACK_DEMO_README.md](./SLACK_DEMO_README.md) for full documentation.

---

### 3. Deployment Notification (`slack-deployment-notification.json`)
**Difficulty**: Advanced
**Actions Used**: Conditionals, Error Handling, Multiple Slack Actions

Production-ready deployment notification workflow with:
- Conditional branching (success/failure paths)
- Rich Slack Block Kit formatting
- Different reactions for success/failure
- On-call engineer alerts via DM
- Interactive buttons in messages

**Flow**:
```
Webhook → Announce Start → Execute Deployment → Check Result
                                                      ↓
                                                 [Success?]
                                                   /     \
                                              TRUE         FALSE
                                               ↓             ↓
                                        Update Success  Update Failure
                                               ↓             ↓
                                          Add ✅         Add ❌
                                                           ↓
                                                    Alert On-Call
```

**Real-World Use Cases**:
- CI/CD pipeline notifications
- Infrastructure deployments
- Release management
- Incident response

---

## Configuration Guide

### Required Slack Credentials

All Slack workflows require OAuth credentials with these scopes:

```
chat:write              # Send messages
chat:write.public       # Post to public channels
reactions:write         # Add emoji reactions
users:read              # Look up users
users:read.email        # Find users by email
```

### Setting Up Credentials

1. **Create Slack App**:
   - Go to https://api.slack.com/apps
   - Click "Create New App" → "From Scratch"
   - Name it "Gorax Bot" and select your workspace

2. **Configure OAuth Scopes**:
   - Navigate to "OAuth & Permissions"
   - Add the scopes listed above
   - Install app to workspace

3. **Get OAuth Token**:
   - After installation, copy the "Bot User OAuth Token"
   - It starts with `xoxb-`

4. **Store in Gorax**:
```bash
curl -X POST http://localhost:8080/api/credentials \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Slack Bot Token",
    "type": "oauth2",
    "config": {
      "access_token": "xoxb-your-token-here"
    }
  }'
```

### Finding Channel IDs

Slack channel IDs are required for sending messages:

1. **Via Slack UI**:
   - Right-click the channel → View channel details
   - Scroll to bottom → Copy channel ID
   - Format: `C1234567890`

2. **Via API**:
```bash
curl -H "Authorization: Bearer xoxb-your-token" \
  https://slack.com/api/conversations.list
```

### Finding User IDs

For sending DMs:

1. **By Email** (easiest):
   - Just use the email address: `user@company.com`
   - Gorax will look up the user ID automatically

2. **By User ID**:
   - In Slack, click user profile → "View full profile" → More → Copy member ID
   - Format: `U1234567890`

---

## Template Variables Reference

All workflows support template variables for dynamic content:

### Trigger Data
```
{{trigger.timestamp}}          # When workflow was triggered
{{trigger.body.field}}         # Webhook payload fields
{{trigger.headers.field}}      # Request headers
```

### Workflow Metadata
```
{{workflow.id}}                # Workflow ID
{{workflow.name}}              # Workflow name
{{execution.id}}               # Unique execution ID
{{execution.started_at}}       # Execution start time
```

### Step Outputs
```
{{steps.step-id.field}}        # Any field from previous steps
{{steps.slack-send-1.ts}}      # Message timestamp
{{steps.slack-send-1.channel}} # Channel ID
{{steps.http-1.status}}        # HTTP response status
{{steps.http-1.body.data}}     # HTTP response data
```

### Environment Variables
```
${env.VARIABLE_NAME}           # Environment variable
${credential.credential_id}     # Credential value
```

---

## Slack Block Kit Guide

Slack Block Kit provides rich message formatting. Here are common patterns:

### Basic Message with Header
```json
[
  {
    "type": "header",
    "text": {
      "type": "plain_text",
      "text": "🚀 Deployment Started"
    }
  },
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "Your deployment is now in progress."
    }
  }
]
```

### Message with Fields (Side-by-Side)
```json
[
  {
    "type": "section",
    "fields": [
      {
        "type": "mrkdwn",
        "text": "*Environment:*\nProduction"
      },
      {
        "type": "mrkdwn",
        "text": "*Status:*\n✅ Success"
      }
    ]
  }
]
```

### Message with Buttons
```json
[
  {
    "type": "actions",
    "elements": [
      {
        "type": "button",
        "text": {
          "type": "plain_text",
          "text": "View Details"
        },
        "url": "https://example.com/details"
      },
      {
        "type": "button",
        "text": {
          "type": "plain_text",
          "text": "Rollback"
        },
        "style": "danger",
        "url": "https://example.com/rollback"
      }
    ]
  }
]
```

### Message with Code Block
```json
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*Error Message:*\n```Error: Connection timeout```"
    }
  }
]
```

**Tool**: Use [Slack Block Kit Builder](https://app.slack.com/block-kit-builder) to design your blocks visually.

---

## Emoji Reference

Common emojis for Slack reactions:

| Emoji | Name | Use Case |
|-------|------|----------|
| 👍 | `thumbsup` | Approval, success |
| 👎 | `thumbsdown` | Disapproval, failure |
| ✅ | `white_check_mark` | Completed successfully |
| ❌ | `x` | Failed, error |
| ⚠️ | `warning` | Warning, attention needed |
| 👀 | `eyes` | Watching, reviewing |
| 🚀 | `rocket` | Deployment, launch |
| 🎉 | `tada` | Celebration, milestone |
| 🔥 | `fire` | Hot, urgent |
| ❤️ | `heart` | Favorite, important |
| 🔄 | `arrows_counterclockwise` | In progress, processing |
| ⏸️ | `pause_button` | Paused, waiting |
| 🛑 | `stop_sign` | Stopped, blocked |

View all emojis: [Emoji Cheat Sheet](https://www.webfx.com/tools/emoji-cheat-sheet/)

---

## Troubleshooting

### "channel_not_found" Error
- Verify channel ID is correct
- Ensure bot is added to channel: `/invite @Gorax Bot`
- Check bot has `chat:write.public` scope for public channels

### "not_in_channel" Error
- Bot must be invited to private channels
- Use `/invite @Gorax Bot` in the channel

### "user_not_found" Error
- Check email address is correct
- Ensure bot has `users:read` and `users:read.email` scopes
- User must be in the workspace

### "invalid_auth" Error
- Verify OAuth token is correct
- Check token hasn't expired or been revoked
- Ensure credentials are properly stored in Gorax

### Message Not Updating
- Verify timestamp format matches original message
- Check channel ID matches original message
- Message must exist (can't update deleted messages)

### Reactions Not Working
- Ensure emoji name is correct (no colons needed)
- Check bot has `reactions:write` scope
- Message must be visible to bot

---

## Best Practices

### 1. Message Design
- Use clear, concise text
- Include relevant context
- Use emojis for visual indicators
- Structure with Block Kit for readability

### 2. Error Handling
- Always include try/catch logic
- Send alerts for failures
- Include error details in messages
- Provide actionable buttons (rollback, view logs)

### 3. Template Variables
- Validate data exists before using
- Provide fallback values
- Use descriptive step names
- Document expected data structure

### 4. Performance
- Avoid excessive API calls
- Batch updates when possible
- Use threading for related messages
- Set appropriate timeouts

### 5. Security
- Never expose credentials in messages
- Use environment variables for secrets
- Validate webhook signatures
- Restrict channel access appropriately

---

## Additional Resources

- [Slack API Documentation](https://api.slack.com/)
- [Block Kit Builder](https://app.slack.com/block-kit-builder)
- [Block Kit Reference](https://api.slack.com/reference/block-kit)
- [Gorax Documentation](../docs/README.md)
- [Slack App Management](https://api.slack.com/apps)

---

## Contributing

Have a great workflow example? Submit a PR with:
1. The workflow JSON file
2. Documentation explaining the use case
3. Configuration requirements
4. Screenshots (if applicable)

## Support

Questions or issues?
- Check the [Gorax Documentation](../docs/README.md)
- Open an issue on GitHub
- Join our community Slack (coming soon)
