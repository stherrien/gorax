# Slack Integration Quick Start

Get started with Slack workflows in 5 minutes!

## 📋 Prerequisites Checklist

- [ ] Slack workspace access
- [ ] Admin permissions to install apps
- [ ] Gorax instance running
- [ ] Channel ID where you want to post

## 🚀 Quick Setup (5 minutes)

### Step 1: Create Slack App (2 min)
1. Go to https://api.slack.com/apps
2. Click **"Create New App"** → **"From Scratch"**
3. Name: `Gorax Bot`
4. Select your workspace

### Step 2: Add Permissions (1 min)
1. Go to **"OAuth & Permissions"**
2. Scroll to **"Scopes"** → **"Bot Token Scopes"**
3. Add these scopes:
   ```
   chat:write
   chat:write.public
   reactions:write
   users:read
   users:read.email
   ```
4. Click **"Install to Workspace"**
5. Copy the **Bot User OAuth Token** (starts with `xoxb-`)

### Step 3: Configure Gorax (1 min)
```bash
# Store Slack credentials
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

### Step 4: Get Channel ID (30 sec)
1. Open Slack
2. Right-click your channel → **"View channel details"**
3. Scroll to bottom → Copy **Channel ID** (e.g., `C1234567890`)

### Step 5: Run Hello World (30 sec)
```bash
cd /Users/shawntherrien/Projects/gorax/examples

# 1. Edit slack-hello-world.json
# Replace "C1234567890" with your actual channel ID

# 2. Import workflow
curl -X POST http://localhost:8080/api/workflows \
  -H "Content-Type: application/json" \
  -d @slack-hello-world.json

# 3. Trigger it!
curl -X POST http://localhost:8080/webhooks/hello-slack
```

### Expected Result ✅
In your Slack channel, you should see:
```
Gorax Bot  10:30 AM
Hello from Gorax! 👋
```
With a 👋 reaction added automatically!

---

## 🎯 Next Steps

### Try the Full Demo
```bash
# Edit channel ID in the file first!
curl -X POST http://localhost:8080/api/workflows \
  -d @slack-notification-workflow.json

# Trigger it
curl -X POST http://localhost:8080/webhooks/slack-demo \
  -H "Content-Type: application/json" \
  -d '{"message": "Testing full demo"}'
```

### Build Your Own Workflow

**Use Case**: Deploy notification

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "data": {
        "name": "Deploy Hook",
        "config": { "path": "/deploy" }
      }
    },
    {
      "id": "notify-1",
      "type": "slack:send_message",
      "data": {
        "name": "Notify Team",
        "config": {
          "channel": "YOUR_CHANNEL_ID",
          "text": "🚀 Deploying {{trigger.body.version}}"
        }
      }
    }
  ],
  "edges": [
    { "id": "e1", "source": "trigger-1", "target": "notify-1" }
  ]
}
```

---

## 📚 Learn More

| Resource | Description |
|----------|-------------|
| [README.md](./README.md) | Complete guide with all examples |
| [SLACK_DEMO_README.md](./SLACK_DEMO_README.md) | Detailed demo walkthrough |
| [WORKFLOW_DIAGRAMS.md](./WORKFLOW_DIAGRAMS.md) | Visual workflow diagrams |

---

## 🔧 Troubleshooting

### "channel_not_found"
**Solution**: Invite bot to channel
```
/invite @Gorax Bot
```

### "not_authed" or "invalid_auth"
**Solution**: Check your token
1. Verify token starts with `xoxb-`
2. Check it's properly stored in Gorax credentials
3. Ensure token hasn't been revoked

### Message not appearing
**Solution**: Check bot permissions
1. Verify bot has `chat:write` scope
2. For public channels, add `chat:write.public`
3. Reinstall app if needed

### Can't find channel ID
**Solution**: Use Slack API
```bash
curl -H "Authorization: Bearer xoxb-YOUR-TOKEN" \
  https://slack.com/api/conversations.list \
  | grep -A5 "channel-name"
```

---

## 💡 Tips

### 1. Test in a Private Channel
Create a test channel to avoid spamming your team while developing.

### 2. Use Template Variables
Reference previous steps:
```
"text": "Processing completed in {{steps.process.duration}}ms"
```

### 3. Add Error Handling
Always include failure paths in production workflows.

### 4. Format with Block Kit
Use the [Block Kit Builder](https://app.slack.com/block-kit-builder) to design rich messages.

### 5. Keep Messages Concise
- Use emojis for quick visual feedback
- Structure with sections
- Include only actionable information

---

## 📞 Support

- **Documentation**: [README.md](./README.md)
- **Issues**: [GitHub Issues](https://github.com/gorax/gorax/issues)
- **Slack API Docs**: https://api.slack.com/

---

## 🎉 You're Ready!

Start building powerful Slack workflows with Gorax. The three example workflows provided cover:

1. **slack-hello-world.json** - Simple message + reaction
2. **slack-notification-workflow.json** - Complete notification flow
3. **slack-deployment-notification.json** - Production deployment example

Mix and match the patterns to create your own custom workflows!
