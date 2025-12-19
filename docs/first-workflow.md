# Your First Workflow

Let's build a simple workflow that sends a Slack message when a webhook is triggered.

## What You'll Build

A workflow that:
1. Receives an HTTP POST request (webhook)
2. Sends a notification to Slack
3. Returns a success response

**Time**: ~5 minutes

## Prerequisites

- Gorax is [installed and running](getting-started.md)
- You have a Slack workspace
- Basic understanding of webhooks

## Step 1: Set Up Slack Integration

### Create a Slack App

1. Go to [Slack API Apps](https://api.slack.com/apps)
2. Click **"Create New App"** → **"From scratch"**
3. Name it "Gorax Notifications"
4. Select your workspace

### Add Permissions

1. Go to **"OAuth & Permissions"**
2. Add these Bot Token Scopes:
   - `chat:write`
   - `chat:write.public`
3. Click **"Install to Workspace"**
4. Copy the **Bot User OAuth Token** (starts with `xoxb-`)

### Get Your Channel ID

1. Open Slack, right-click your channel
2. Select **"View channel details"**
3. Scroll down and copy the **Channel ID** (e.g., `C1234567890`)

### Store Credentials in Gorax

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

Save the returned credential ID for later.

## Step 2: Create the Workflow

### Open Gorax

1. Navigate to `http://localhost:5173`
2. Click **"New Workflow"**
3. Name it "Hello Slack"

### Add a Webhook Trigger

1. From the **Node Palette** (left sidebar), drag **"Webhook"** onto the canvas
2. Click the node to configure:
   - **Name**: "Receive Request"
   - **Path**: `/hello`
   - **Method**: POST
3. Click **"Save"**

### Add a Slack Action

1. Drag **"Slack: Send Message"** onto the canvas
2. Connect the webhook to the Slack node (drag from bottom of webhook to top of Slack node)
3. Click the Slack node to configure:
   - **Name**: "Send Notification"
   - **Channel ID**: `C1234567890` (your channel ID)
   - **Message Text**: `Hello from Gorax! 👋`
4. Click **"Save"**

### Visual Layout

Your workflow should look like this:

```
┌─────────────────┐
│  Webhook        │
│  /hello         │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Slack Send     │
│  Message        │
└─────────────────┘
```

## Step 3: Deploy the Workflow

1. Click **"Save Workflow"** (top right)
2. Click **"Activate"** to enable the workflow

Your webhook is now live at: `http://localhost:8080/webhooks/hello`

## Step 4: Test Your Workflow

### Trigger via cURL

```bash
curl -X POST http://localhost:8080/webhooks/hello \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Testing Gorax!",
    "from": "Your Name"
  }'
```

### Expected Result

✅ **In Slack**: You should see a message:
```
Gorax Bot  10:30 AM
Hello from Gorax! 👋
```

✅ **In Terminal**: You should get:
```json
{
  "execution_id": "exec_abc123",
  "status": "completed",
  "message": "Workflow executed successfully"
}
```

## Step 5: View Execution History

1. In Gorax UI, click **"Executions"** in the sidebar
2. You'll see your workflow execution
3. Click it to see details:
   - Trigger data received
   - Slack API response
   - Execution duration

## Step 6: Enhance Your Workflow

### Use Template Variables

Edit the Slack message to use data from the webhook:

```
Hello {{trigger.body.from}}!
Your message: {{trigger.body.message}}
```

Now test again:
```bash
curl -X POST http://localhost:8080/webhooks/hello \
  -H "Content-Type: application/json" \
  -d '{
    "message": "This is amazing!",
    "from": "Alice"
  }'
```

Slack will show:
```
Hello Alice!
Your message: This is amazing!
```

### Add a Reaction

1. Drag **"Slack: Add Reaction"** below the Send Message node
2. Connect them
3. Configure:
   - **Channel ID**: `{{steps.slack-send-1.channel}}`
   - **Timestamp**: `{{steps.slack-send-1.ts}}`
   - **Emoji**: `wave`

Now your message will get a 👋 reaction automatically!

## What's Next?

🎉 Congratulations! You've created your first workflow.

### Next Steps

- 📖 [Learn about Template Variables](template-variables.md)
- 🔧 [Explore More Integrations](integrations/slack.md)
- 💡 [Browse Example Workflows](../examples/)
- 🚀 [Add Control Flow (if/then, loops)](control-flow.md)

### Common Use Cases

- **DevOps**: Deploy notifications, build status updates
- **Support**: Ticket alerts, customer notifications
- **Sales**: Lead notifications, CRM updates
- **Monitoring**: Alert on errors, system health checks

## Troubleshooting

### Message Not Appearing in Slack

1. **Check bot is in channel**: Type `/invite @Gorax Bot` in your Slack channel
2. **Verify credentials**: Ensure the Bot Token is correctly stored
3. **Check permissions**: Bot needs `chat:write` scope
4. **View logs**: Check execution details in Gorax UI

### Webhook Not Triggering

1. **Verify URL**: Should be `http://localhost:8080/webhooks/hello`
2. **Check workflow is active**: Green indicator in UI
3. **Test with browser**: Try POST request from Postman

### Need Help?

- 💬 [Ask in Discussions](https://github.com/stherrien/gorax/discussions)
- 📖 [Full Documentation](README.md)
- 🐛 [Report Issues](https://github.com/stherrien/gorax/issues)
