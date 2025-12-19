# gorax - Quick Start Guide

## Overview

gorax is a workflow automation platform (Tines replacement) built with Go + PostgreSQL backend and React + TypeScript frontend.

## Architecture

- **Frontend**: Vite + React 18 + TypeScript + ReactFlow
- **Backend**: Go 1.22+ with Chi router
- **Database**: PostgreSQL 16 (port 5433)
- **Cache**: Redis 7
- **Auth**: Development mode (bypasses Ory Kratos for local testing)

## Prerequisites

- Go 1.22+
- Node.js 18+
- Docker & Docker Compose
- Make

## Quick Start (3 steps)

### 1. Start Backend Services

```bash
make dev-simple
```

This starts:
- PostgreSQL on `localhost:5433` (user: postgres, password: postgres, db: gorax)
- Redis on `localhost:6379`
- Auto-applies database migrations with seed data

### 2. Start API Server

```bash
make run-api-dev
```

This starts the Go API server on `http://localhost:8080` with development authentication (no Kratos required).

### 3. Start Frontend

```bash
cd web
npm install  # First time only
npm run dev
```

Frontend runs on `http://localhost:5173` (or 5174 if 5173 is in use).

## Access the Application

1. Open browser to `http://localhost:5173`
2. You'll see the gorax dashboard with:
   - Pre-seeded "Hello World Workflow"
   - Test tenant already configured
   - Development authentication enabled

## Available Endpoints

### API Endpoints
- `GET /health` - Health check
- `GET /api/v1/workflows` - List workflows
- `POST /api/v1/workflows` - Create workflow
- `GET /api/v1/workflows/{id}` - Get workflow
- `PUT /api/v1/workflows/{id}` - Update workflow
- `DELETE /api/v1/workflows/{id}` - Delete workflow
- `POST /api/v1/workflows/{id}/execute` - Execute workflow
- `GET /api/v1/executions` - List executions
- `GET /api/v1/executions/{id}` - Get execution details

### Development Details
- **Tenant ID**: `00000000-0000-0000-0000-000000000001`
- **User ID**: `00000000-0000-0000-0000-000000000002`
- All API requests automatically include `X-Tenant-ID` header in development mode

## Testing

### Frontend Tests
```bash
cd web
npm test
```

**Results**: 216 tests passing ✅

### Backend Tests
```bash
go test ./...
```

## Project Structure

```
gorax/
├── cmd/
│   ├── api/          # API server entry point
│   └── worker/       # Background worker (future)
├── internal/
│   ├── api/          # HTTP handlers & routing
│   ├── workflow/     # Workflow domain logic
│   ├── tenant/       # Multi-tenancy
│   └── config/       # Configuration
├── migrations/       # Database migrations
├── web/             # React frontend
│   ├── src/
│   │   ├── pages/   # Page components
│   │   ├── components/
│   │   │   └── canvas/  # ReactFlow workflow editor
│   │   ├── api/     # API client
│   │   └── hooks/   # React hooks
│   └── tests/       # Frontend tests
└── docker-compose.dev.yml  # Simplified dev stack
```

## Useful Commands

```bash
# Backend
make dev-simple          # Start postgres + redis
make dev-simple-down     # Stop services
make run-api-dev         # Run API in development mode
make build               # Build Go binaries
make test                # Run Go tests

# Frontend
cd web && npm run dev    # Start dev server
cd web && npm test       # Run tests
cd web && npm run build  # Production build

# Database
make db-seed             # Re-seed database with sample data
docker exec -it gorax-postgres-dev psql -U postgres -d gorax
```

## Development Features

### Development Authentication
- Bypasses Ory Kratos for faster development
- Uses `X-Tenant-ID` header for tenant resolution
- Default test tenant and user pre-configured

### Hot Reload
- Frontend: Vite HMR (instant updates)
- Backend: Use `make run-api-dev` (manual restart needed)

### Sample Data
- 1 test tenant ("Development Tenant")
- 1 sample workflow ("Hello World Workflow")
- Sample webhook trigger with HTTP and Transform actions

## Next Steps

1. **Explore the UI**: Open `http://localhost:5173` and navigate through:
   - Dashboard
   - Workflows list
   - Workflow editor (visual canvas)
   - Executions list

2. **Create a Workflow**: Click "New Workflow" and use the drag-and-drop editor

3. **Test Execution**: Execute a workflow and view the execution details

4. **Read the Docs**: Check `/docs` folder for detailed documentation

## Troubleshooting

### Port Already in Use
- Postgres: Change `5433:5432` in `docker-compose.dev.yml`
- API: Set `SERVER_ADDRESS=:8081` env var
- Frontend: Vite will auto-select next available port

### Database Connection Issues
```bash
# Reset database
make dev-simple-down
docker volume rm gorax_postgres_dev_data
make dev-simple
```

### Frontend Not Connecting to API
- Check `web/.env.local` has `VITE_API_URL=http://localhost:8080`
- Verify API is running: `curl http://localhost:8080/health`

## Production Deployment

For production deployment with Ory Kratos authentication:

```bash
make docker-up  # Full stack with Kratos
```

See `docs/DEPLOYMENT.md` for details (coming soon).

## Support

- **Issues**: https://github.com/gorax/gorax/issues
- **Documentation**: `/docs` folder
- **Plan**: `/docs/PLAN.md`

---

**Status**: ✅ MVP Complete - Frontend + Backend fully functional!
