<div align="center">

# 🚀 Gorax

### Modern Workflow Automation Platform

Build, deploy, and manage complex workflows with ease. No code required.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![Sponsor](https://img.shields.io/badge/Sponsor-%E2%9D%A4-red?logo=github)](https://github.com/sponsors/stherrien)

[Getting Started](#-quick-start) • [Documentation](docs/) • [Examples](examples/) • [Contributing](#-contributing)

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🎨 Visual Workflow Builder
Drag-and-drop interface for creating workflows. No coding required.

### 🔗 Rich Integrations
Connect with Slack, HTTP APIs, webhooks, and more. Extensible plugin system.

### ⚡ Real-time Execution
Monitor workflows as they run with live updates and detailed logs.

</td>
<td width="50%">

### 🔐 Secure by Default
Enterprise-grade security with encrypted credentials and role-based access.

### 🎯 Template Variables
Dynamic data interpolation between workflow steps.

### 📊 Complete Observability
Execution history, performance metrics, and audit trails.

</td>
</tr>
</table>

---

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/stherrien/gorax.git
cd gorax

# Start dependencies (PostgreSQL + Redis)
make dev-simple

# Configure environment
cp .env.example .env

# Run migrations
make db-migrate

# Start the API server
make run-api-dev

# In another terminal, start the frontend
cd web && npm install && npm run dev
```

**Open your browser** → `http://localhost:5173`

📖 **Full guide**: [Getting Started](docs/getting-started.md)

---

## 🎯 Use Cases

<table>
<tr>
<td>

### 🔧 DevOps Automation
- Deploy notifications
- CI/CD pipeline updates
- Infrastructure monitoring
- Automated rollbacks

</td>
<td>

### 💼 Business Processes
- Approval workflows
- Data synchronization
- Report generation
- Automated notifications

</td>
</tr>
<tr>
<td>

### 🎫 IT Operations
- Incident response
- Ticket routing
- Alert management
- Service health checks

</td>
<td>

### 📈 Customer Support
- Automated responses
- Ticket escalation
- SLA monitoring
- Customer notifications

</td>
</tr>
</table>

---

## 🔌 Integrations

### Available Now

| Integration | Send | Receive | Actions |
|------------|------|---------|---------|
| **Slack** | ✅ | ✅ | Messages, DMs, Reactions, Updates |
| **HTTP/REST** | ✅ | ✅ | Any API endpoint |
| **Webhooks** | ✅ | ✅ | Inbound & outbound |
| **JavaScript** | ✅ | - | Custom code execution |

### Coming Soon

GitHub • Jira • AWS Services • Google Workspace • Database Connectors • Email

---

## 📖 Example Workflow

Create a deployment notification in minutes:

```json
{
  "nodes": [
    {
      "type": "trigger:webhook",
      "config": { "path": "/deploy" }
    },
    {
      "type": "slack:send_message",
      "config": {
        "channel": "C1234567890",
        "text": "🚀 Deploying {{trigger.body.version}}"
      }
    }
  ]
}
```

**More examples**: [examples/](examples/)

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      React Frontend                          │
│              Visual Workflow Builder + Dashboard             │
└───────────────────────────┬─────────────────────────────────┘
                            │ REST API
┌───────────────────────────▼─────────────────────────────────┐
│                       Go Backend                             │
│            Chi Router • Executor • Integrations              │
└─────┬────────────────────────────────────────┬───────────────┘
      │                                        │
      ▼                                        ▼
┌─────────────┐                        ┌─────────────┐
│ PostgreSQL  │                        │    Redis    │
│  Workflows  │                        │    Cache    │
│  Executions │                        │   Sessions  │
└─────────────┘                        └─────────────┘
```

---

## 🛠️ Development

### Prerequisites

- Go 1.25+
- Node.js 18+
- PostgreSQL 14+
- Redis 6+

### Commands

```bash
make               # Show all available commands
make all           # Install deps and build
make test          # Run all tests
make lint          # Run linters
make build         # Build binaries
```

### Testing

```bash
# Backend tests
make test

# Backend tests with coverage
make test-coverage

# Frontend tests
cd web && npm test
```

**Development guide**: [Local Development](docs/local-development.md)

---

## 📚 Documentation

| Resource | Description |
|----------|-------------|
| [Getting Started](docs/getting-started.md) | Installation and setup |
| [Your First Workflow](docs/first-workflow.md) | Build a workflow in 5 minutes |
| [Slack Integration](docs/integrations/slack.md) | Complete Slack guide |
| [API Reference](docs/api-reference.md) | REST API documentation |
| [Contributing](docs/contributing.md) | Join the community |

**Full docs**: [docs/](docs/)

---

## 🤝 Contributing

We love contributions! Check out our [Contributing Guide](docs/contributing.md).

```bash
# 1. Fork the repo
# 2. Create your feature branch
git checkout -b feature/amazing-feature

# 3. Write tests (we follow TDD)
make test

# 4. Commit your changes
git commit -m 'Add amazing feature'

# 5. Push and open a PR
git push origin feature/amazing-feature
```

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 💖 Sponsors

Support this project by becoming a sponsor. Your logo will show up here with a link to your website.

[[Become a sponsor](https://github.com/sponsors/stherrien)]

---

## 🙏 Acknowledgments

Built with:
- [Go](https://go.dev/) - Backend
- [React](https://react.dev/) - Frontend
- [ReactFlow](https://reactflow.dev/) - Workflow canvas
- [Tailwind CSS](https://tailwindcss.com/) - Styling
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - Cache

---

## 📞 Support

- 📖 [Documentation](docs/)
- 💬 [Discussions](https://github.com/stherrien/gorax/discussions)
- 🐛 [Issues](https://github.com/stherrien/gorax/issues)
- 💌 Email: shawn@gorax.dev

---

<div align="center">

Made with ❤️ by [Shawn Therrien](https://github.com/stherrien)

⭐ Star us on GitHub — it motivates us a lot!

</div>
