import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import App from './App'

// Mock all page components
vi.mock('./pages/Dashboard', () => ({
  default: () => <div>Dashboard Page</div>,
}))

vi.mock('./pages/WorkflowList', () => ({
  default: () => <div>Workflow List Page</div>,
}))

vi.mock('./pages/WorkflowEditor', () => ({
  default: () => <div>Workflow Editor Page</div>,
}))

vi.mock('./pages/WebhookList', () => ({
  default: () => <div>Webhook List Page</div>,
}))

vi.mock('./pages/WebhookDetail', () => ({
  default: () => <div>Webhook Detail Page</div>,
}))

vi.mock('./pages/Executions', () => ({
  default: () => <div>Executions Page</div>,
}))

vi.mock('./pages/ExecutionDetail', () => ({
  default: () => <div>Execution Detail Page</div>,
}))

vi.mock('./pages/Schedules', () => ({
  default: () => <div>Schedules Page</div>,
}))

vi.mock('./pages/CreateSchedule', () => ({
  default: () => <div>Create Schedule Page</div>,
}))

vi.mock('./pages/EditSchedule', () => ({
  default: () => <div>Edit Schedule Page</div>,
}))

vi.mock('./pages/CredentialManager', () => ({
  CredentialManager: () => <div>Credential Manager Page</div>,
}))

vi.mock('./pages/AIWorkflowBuilder', () => ({
  default: () => <div>AI Workflow Builder Page</div>,
}))

vi.mock('./pages/Marketplace', () => ({
  default: () => <div>Marketplace Page</div>,
}))

vi.mock('./pages/Analytics', () => ({
  default: () => <div>Analytics Page</div>,
}))

vi.mock('./pages/Documentation', () => ({
  default: () => <div>Documentation Page</div>,
}))

vi.mock('./pages/OAuthConnections', () => ({
  OAuthConnections: () => <div>OAuth Connections Page</div>,
}))

vi.mock('./pages/admin/SSOSettings', () => ({
  SSOSettings: () => <div>SSO Settings Page</div>,
}))

vi.mock('./pages/admin/AuditLogs', () => ({
  AuditLogs: () => <div>Audit Logs Page</div>,
}))

vi.mock('./components/oauth/OAuthCallback', () => ({
  OAuthCallback: () => <div>OAuth Callback Page</div>,
}))

vi.mock('./components/Layout', () => ({
  default: () => {
    const { Outlet } = require('react-router-dom')
    return <div><Outlet /></div>
  },
}))

describe('App Routing', () => {
  describe('Route Order - Specific Routes Before Dynamic', () => {
    it('should route /workflows/new to WorkflowEditor (not treat "new" as ID)', () => {
      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('Workflow Editor Page')).toBeInTheDocument()
    })

    it('should route /workflows/:id to WorkflowEditor with ID', () => {
      render(
        <MemoryRouter initialEntries={['/workflows/550e8400-e29b-41d4-a716-446655440000']}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('Workflow Editor Page')).toBeInTheDocument()
    })

    it('should route /schedules/new to CreateSchedule (not treat "new" as ID)', () => {
      render(
        <MemoryRouter initialEntries={['/schedules/new']}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('Create Schedule Page')).toBeInTheDocument()
    })

    it('should route /schedules/:id/edit to EditSchedule with ID', () => {
      render(
        <MemoryRouter initialEntries={['/schedules/550e8400-e29b-41d4-a716-446655440000/edit']}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('Edit Schedule Page')).toBeInTheDocument()
    })
  })

  describe('All Routes Load Correctly', () => {
    const routes = [
      { path: '/', expectedText: 'Dashboard Page' },
      { path: '/workflows', expectedText: 'Workflow List Page' },
      { path: '/workflows/new', expectedText: 'Workflow Editor Page' },
      { path: '/webhooks', expectedText: 'Webhook List Page' },
      { path: '/executions', expectedText: 'Executions Page' },
      { path: '/schedules', expectedText: 'Schedules Page' },
      { path: '/schedules/new', expectedText: 'Create Schedule Page' },
      { path: '/credentials', expectedText: 'Credential Manager Page' },
      { path: '/oauth/connections', expectedText: 'OAuth Connections Page' },
      { path: '/ai/builder', expectedText: 'AI Workflow Builder Page' },
      { path: '/marketplace', expectedText: 'Marketplace Page' },
      { path: '/analytics', expectedText: 'Analytics Page' },
      { path: '/docs', expectedText: 'Documentation Page' },
      { path: '/admin/sso', expectedText: 'SSO Settings Page' },
      { path: '/admin/audit-logs', expectedText: 'Audit Logs Page' },
    ]

    routes.forEach(({ path, expectedText }) => {
      it(`should render ${path} correctly`, () => {
        render(
          <MemoryRouter initialEntries={[path]}>
            <App />
          </MemoryRouter>
        )

        expect(screen.getByText(expectedText)).toBeInTheDocument()
      })
    })
  })

  describe('Dynamic Routes with Valid UUIDs', () => {
    const validUUID = '550e8400-e29b-41d4-a716-446655440000'

    it('should route to WebhookDetail with valid UUID', () => {
      render(
        <MemoryRouter initialEntries={[`/webhooks/${validUUID}`]}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('Webhook Detail Page')).toBeInTheDocument()
    })

    it('should route to ExecutionDetail with valid UUID', () => {
      render(
        <MemoryRouter initialEntries={[`/executions/${validUUID}`]}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('Execution Detail Page')).toBeInTheDocument()
    })

    it('should route to WorkflowEditor with valid UUID', () => {
      render(
        <MemoryRouter initialEntries={[`/workflows/${validUUID}`]}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('Workflow Editor Page')).toBeInTheDocument()
    })

    it('should route to EditSchedule with valid UUID', () => {
      render(
        <MemoryRouter initialEntries={[`/schedules/${validUUID}/edit`]}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('Edit Schedule Page')).toBeInTheDocument()
    })
  })

  describe('OAuth Callback Route', () => {
    it('should route to OAuthCallback with provider parameter', () => {
      render(
        <MemoryRouter initialEntries={['/oauth/callback/google']}>
          <App />
        </MemoryRouter>
      )

      expect(screen.getByText('OAuth Callback Page')).toBeInTheDocument()
    })
  })
})
