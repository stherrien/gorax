# Slack Integration

Connect Gorax to Slack to send messages, notifications, and more.

## Features

- ✅ Send messages to channels
- ✅ Send direct messages to users
- ✅ Update existing messages
- ✅ Add emoji reactions
- ✅ Thread support
- ✅ Rich formatting with Block Kit

## Setup

### 1. Create Slack App

1. Go to [Slack API Apps](https://api.slack.com/apps)
2. Click **"Create New App"** → **"From scratch"**
3. Name it (e.g., "Gorax Bot")
4. Select your workspace

### 2. Configure Permissions

Add these **Bot Token Scopes**:

| Scope | Purpose |
|-------|---------|
| `chat:write` | Send messages |
| `chat:write.public` | Post to public channels without joining |
| `reactions:write` | Add emoji reactions |
| `users:read` | Look up users by email |
| `users:read.email` | Read user email addresses |

### 3. Install to Workspace

1. Click **"Install to Workspace"**
2. Authorize the app
3. Copy the **Bot User OAuth Token** (starts with `xoxb-`)

### 4. Store Credentials in Gorax

```bash
curl -X POST http://localhost:8080/api/credentials \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Slack Bot Token",
    "type": "oauth2",
    "config": {
      "access_token": "xoxb-YOUR-TOKEN-HERE"
    }
  }'
```

## Actions

### Send Message

Send a message to a Slack channel.

**Node Type**: `slack:send_message`

**Configuration**:

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `channel` | ✅ | Channel ID | `C1234567890` |
| `text` | No | Plain text message | `Hello World!` |
| `blocks` | No | [Block Kit](https://api.slack.com/block-kit) JSON | See below |
| `thread_ts` | No | Reply in thread | `{{steps.send.ts}}` |
| `username` | No | Custom bot name | `Deploy Bot` |
| `icon_emoji` | No | Custom bot icon | `:robot_face:` |

**Example**:
```json
{
  "channel": "C1234567890",
  "text": "Deployment completed successfully! 🚀",
  "username": "Deploy Bot",
  "icon_emoji": ":rocket:"
}
```

**Returns**:
```json
{
  "channel": "C1234567890",
  "ts": "1503435956.000247",
  "message": { ... }
}
```

### Send Direct Message

Send a DM to a specific user.

**Node Type**: `slack:send_dm`

**Configuration**:

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `user` | ✅ | Email or User ID | `user@company.com` or `U1234567890` |
| `text` | No | Plain text message | `Your report is ready` |
| `blocks` | No | Block Kit JSON | See below |

**Example**:
```json
{
  "user": "alice@company.com",
  "text": "Your workflow completed successfully!"
}
```

### Update Message

Modify an existing message.

**Node Type**: `slack:update_message`

**Configuration**:

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `channel` | ✅ | Channel ID | `{{steps.send.channel}}` |
| `ts` | ✅ | Message timestamp | `{{steps.send.ts}}` |
| `text` | No | New message text | `Updated!` |
| `blocks` | No | New Block Kit JSON | See below |

**Example**:
```json
{
  "channel": "{{steps.send-message.channel}}",
  "ts": "{{steps.send-message.ts}}",
  "text": "✅ Deployment completed!"
}
```

### Add Reaction

Add an emoji reaction to a message.

**Node Type**: `slack:add_reaction`

**Configuration**:

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `channel` | ✅ | Channel ID | `{{steps.send.channel}}` |
| `timestamp` | ✅ | Message timestamp | `{{steps.send.ts}}` |
| `emoji` | ✅ | Emoji name (no colons) | `thumbsup` |

**Common Emojis**:
- `thumbsup` 👍
- `white_check_mark` ✅
- `rocket` 🚀
- `eyes` 👀
- `fire` 🔥
- `tada` 🎉

**Example**:
```json
{
  "channel": "{{steps.send-message.channel}}",
  "timestamp": "{{steps.send-message.ts}}",
  "emoji": "white_check_mark"
}
```

## Block Kit

Slack's [Block Kit](https://api.slack.com/block-kit) enables rich message formatting.

### Simple Section

```json
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*Deployment Status*\n✅ Success"
    }
  }
]
```

### Header + Fields

```json
[
  {
    "type": "header",
    "text": {
      "type": "plain_text",
      "text": "🚀 Deployment Complete"
    }
  },
  {
    "type": "section",
    "fields": [
      {
        "type": "mrkdwn",
        "text": "*Environment:*\nProduction"
      },
      {
        "type": "mrkdwn",
        "text": "*Version:*\nv2.1.0"
      }
    ]
  }
]
```

### With Actions (Buttons)

```json
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "Deployment needs approval"
    }
  },
  {
    "type": "actions",
    "elements": [
      {
        "type": "button",
        "text": {
          "type": "plain_text",
          "text": "Approve"
        },
        "style": "primary",
        "url": "https://example.com/approve"
      },
      {
        "type": "button",
        "text": {
          "type": "plain_text",
          "text": "Reject"
        },
        "style": "danger",
        "url": "https://example.com/reject"
      }
    ]
  }
]
```

Use the [Block Kit Builder](https://app.slack.com/block-kit-builder) to design your blocks visually.

## Common Patterns

### Deploy Notification

```
Webhook → Send Message → HTTP Deploy → Update Message → Add Reaction
```

1. Announce deployment start
2. Run deployment
3. Update message with result
4. Add ✅ or ❌ reaction

### Error Alert

```
Error Webhook → Send DM to On-Call
```

Send immediate DM when errors occur.

### Progress Updates

```
Start → Send Message → Loop → Update Message (each iteration)
```

Update a single message as workflow progresses.

## Template Variables

Reference previous step outputs:

```
Channel: {{steps.send-message.channel}}
Timestamp: {{steps.send-message.ts}}
Trigger Data: {{trigger.body.message}}
Environment: {{env.APP_ENV}}
```

## Troubleshooting

### "channel_not_found"

**Solution**: Invite bot to channel
```
/invite @Gorax Bot
```

### "not_in_channel"

**Solution**: Bot must be in private channels
```
/invite @Gorax Bot
```

### "user_not_found"

**Causes**:
- Wrong email address
- User not in workspace
- Missing `users:read.email` scope

### "invalid_auth"

**Causes**:
- Token expired
- Token revoked
- Wrong credential in Gorax

### Message Not Updating

**Causes**:
- Wrong timestamp format
- Message deleted
- Channel ID mismatch

## Best Practices

### 1. Use Threading

Keep related messages together:
```json
{
  "channel": "C1234567890",
  "text": "Follow-up message",
  "thread_ts": "{{steps.initial-message.ts}}"
}
```

### 2. Progressive Updates

Update a single message instead of spamming:
```
Send: "🔄 Starting..."
Update: "⏳ In progress..."
Update: "✅ Complete!"
```

### 3. Error Handling

Always include error notifications:
```
Send Message
  ↓
[Try] Deploy
  ↓
[If Success] Update → Add ✅
[If Failure] Update → Add ❌ → Send DM to Team
```

### 4. Rate Limiting

Slack has rate limits:
- 1 message per second per channel
- 20 API calls per minute

Gorax handles this automatically with retries.

## Examples

See [examples/](../../examples/) for complete workflow examples:
- [Hello World](../../examples/slack-hello-world.json)
- [Deployment Notifications](../../examples/slack-deployment-notification.json)
- [Full Demo](../../examples/slack-notification-workflow.json)

## API Reference

- [Slack API Docs](https://api.slack.com/)
- [Block Kit Reference](https://api.slack.com/reference/block-kit)
- [Block Kit Builder](https://app.slack.com/block-kit-builder)
