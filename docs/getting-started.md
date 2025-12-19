# Getting Started with Gorax

This guide will walk you through installing Gorax and creating your first workflow.

## Prerequisites

Before you begin, ensure you have:

- **Go** 1.25 or higher ([Download](https://go.dev/dl/))
- **PostgreSQL** 14+ ([Download](https://www.postgresql.org/download/))
- **Redis** 6+ ([Download](https://redis.io/download))
- **Node.js** 18+ ([Download](https://nodejs.org/))
- **Git** ([Download](https://git-scm.com/downloads))

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/stherrien/gorax.git
cd gorax
```

### 2. Start Dependencies

Using Docker (recommended):

```bash
docker-compose up -d postgres redis
```

Or install PostgreSQL and Redis manually.

### 3. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your settings:

```env
# Database
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gorax

# Redis
REDIS_ADDRESS=localhost:6379

# Server
SERVER_ADDRESS=:8080
APP_ENV=development
```

### 4. Run Database Migrations

```bash
make migrate-up
```

### 5. Start the API Server

```bash
make run-api
```

The API will be available at `http://localhost:8080`

### 6. Start the Frontend

In a new terminal:

```bash
cd web
npm install
npm run dev
```

The UI will be available at `http://localhost:5173`

## Verify Installation

### Check API Health

```bash
curl http://localhost:8080/health
```

You should see:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### Open the Web UI

Navigate to `http://localhost:5173` in your browser. You should see the Gorax dashboard.

## Next Steps

- 📖 [Create Your First Workflow](first-workflow.md)
- 🔧 [Configure Integrations](integrations/slack.md)
- 💡 [Explore Examples](../examples/)

## Troubleshooting

### Database Connection Failed

**Problem**: Cannot connect to PostgreSQL

**Solution**:
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -p 5433 -U postgres -d gorax
```

### Redis Connection Failed

**Problem**: Cannot connect to Redis

**Solution**:
```bash
# Check if Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6379 ping
```

### Port Already in Use

**Problem**: Address already in use

**Solution**:
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Frontend Build Errors

**Problem**: npm install fails

**Solution**:
```bash
# Clear cache and reinstall
cd web
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

## Development Tools

### Makefile Commands

```bash
make help          # Show all available commands
make build         # Build the application
make test          # Run tests
make lint          # Run linters
make migrate-up    # Run database migrations
make migrate-down  # Rollback migrations
```

### Hot Reload

For development, use:

```bash
# Backend (install air first: go install github.com/cosmtrek/air@latest)
air

# Frontend (already has hot reload)
cd web && npm run dev
```

## Production Deployment

For production deployment, see the [Deployment Guide](deployment.md).

## Need Help?

- 📖 [Full Documentation](README.md)
- 💬 [Community Discussions](https://github.com/stherrien/gorax/discussions)
- 🐛 [Report Issues](https://github.com/stherrien/gorax/issues)
