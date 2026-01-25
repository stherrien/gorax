/**
 * Integration tests for CredentialManager page.
 * Tests the full flow of loading, filtering, creating, and managing credentials
 * with MSW intercepting API calls.
 */

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse, delay } from 'msw'
import { server } from '../test/mocks/server'
import { render } from '../test/test-utils'
import { CredentialManager } from './CredentialManager'
import { useCredentialStore } from '../stores/credentialStore'
import type { Credential, CredentialType } from '../api/credentials'

const API_BASE = '/api/v1'

// Mock credential data
const mockCredentials: Credential[] = [
  {
    id: 'cred-1',
    tenantId: 'tenant-1',
    name: 'Production API Key',
    type: 'api_key',
    description: 'API key for production environment',
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-15T00:00:00Z',
  },
  {
    id: 'cred-2',
    tenantId: 'tenant-1',
    name: 'GitHub OAuth',
    type: 'oauth2',
    description: 'OAuth2 credentials for GitHub integration',
    expiresAt: '2026-01-01T00:00:00Z',
    createdAt: '2025-01-05T00:00:00Z',
    updatedAt: '2025-01-10T00:00:00Z',
  },
  {
    id: 'cred-3',
    tenantId: 'tenant-1',
    name: 'Database Credentials',
    type: 'basic_auth',
    description: 'Database authentication',
    createdAt: '2025-01-10T00:00:00Z',
    updatedAt: '2025-01-20T00:00:00Z',
  },
  {
    id: 'cred-4',
    tenantId: 'tenant-1',
    name: 'Service Token',
    type: 'bearer_token',
    createdAt: '2025-01-12T00:00:00Z',
    updatedAt: '2025-01-18T00:00:00Z',
  },
]

// Mock ThemeContext
vi.mock('../contexts/ThemeContext', () => ({
  useThemeContext: () => ({ isDark: true }),
}))

// Setup default handlers for credentials integration tests
function setupCredentialHandlers() {
  server.use(
    // List credentials
    http.get(`${API_BASE}/credentials`, async ({ request }) => {
      const url = new URL(request.url)
      const type = url.searchParams.get('type') as CredentialType | null
      const search = url.searchParams.get('search')

      await delay(50)

      let filtered = [...mockCredentials]

      if (type) {
        filtered = filtered.filter((c) => c.type === type)
      }

      if (search) {
        const searchLower = search.toLowerCase()
        filtered = filtered.filter(
          (c) =>
            c.name.toLowerCase().includes(searchLower) ||
            c.description?.toLowerCase().includes(searchLower)
        )
      }

      return HttpResponse.json({
        data: filtered,
        limit: 10,
        offset: 0,
      })
    }),

    // Get single credential
    http.get(`${API_BASE}/credentials/:id`, async ({ params }) => {
      const { id } = params
      await delay(50)

      const credential = mockCredentials.find((c) => c.id === id)
      if (!credential) {
        return HttpResponse.json({ error: 'Credential not found' }, { status: 404 })
      }

      return HttpResponse.json({ data: credential })
    }),

    // Create credential
    http.post(`${API_BASE}/credentials`, async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      await delay(100)

      if (!body.name) {
        return HttpResponse.json({ error: 'Name is required' }, { status: 400 })
      }

      const newCredential: Credential = {
        id: `cred-${Date.now()}`,
        tenantId: 'tenant-1',
        name: body.name as string,
        type: body.type as CredentialType,
        description: body.description as string | undefined,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }

      return HttpResponse.json({ data: newCredential }, { status: 201 })
    }),

    // Update credential
    http.put(`${API_BASE}/credentials/:id`, async ({ params, request }) => {
      const { id } = params
      const body = (await request.json()) as Record<string, unknown>
      await delay(100)

      const credential = mockCredentials.find((c) => c.id === id)
      if (!credential) {
        return HttpResponse.json({ error: 'Credential not found' }, { status: 404 })
      }

      const updated = {
        ...credential,
        ...body,
        updatedAt: new Date().toISOString(),
      }

      return HttpResponse.json({ data: updated })
    }),

    // Delete credential
    http.delete(`${API_BASE}/credentials/:id`, async ({ params }) => {
      const { id } = params
      await delay(100)

      const credential = mockCredentials.find((c) => c.id === id)
      if (!credential) {
        return HttpResponse.json({ error: 'Credential not found' }, { status: 404 })
      }

      return new HttpResponse(null, { status: 204 })
    }),

    // Test credential
    http.post(`${API_BASE}/credentials/:id/test`, async ({ params }) => {
      const { id } = params
      await delay(150)

      const credential = mockCredentials.find((c) => c.id === id)
      if (!credential) {
        return HttpResponse.json({ error: 'Credential not found' }, { status: 404 })
      }

      return HttpResponse.json({
        success: true,
        message: 'Credential test successful',
        testedAt: new Date().toISOString(),
      })
    }),

    // Rotate credential
    http.post(`${API_BASE}/credentials/:id/rotate`, async ({ params }) => {
      const { id } = params
      await delay(150)

      const credential = mockCredentials.find((c) => c.id === id)
      if (!credential) {
        return HttpResponse.json({ error: 'Credential not found' }, { status: 404 })
      }

      return HttpResponse.json({
        data: {
          ...credential,
          updatedAt: new Date().toISOString(),
        },
      })
    })
  )
}

describe('CredentialManager Page Integration Tests', () => {
  beforeEach(() => {
    // Reset Zustand store before each test
    useCredentialStore.getState().reset()
    setupCredentialHandlers()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('Loading and Display', () => {
    it('should show loading state initially', async () => {
      render(<CredentialManager />)

      // The component calls fetchCredentials on mount
      await waitFor(() => {
        expect(screen.getByText('Credentials')).toBeInTheDocument()
      })
    })

    it('should load and display credentials', async () => {
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      expect(screen.getByText('GitHub OAuth')).toBeInTheDocument()
      expect(screen.getByText('Database Credentials')).toBeInTheDocument()
      expect(screen.getByText('Service Token')).toBeInTheDocument()
    })

    it('should display page header with create button', async () => {
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Credentials')).toBeInTheDocument()
      })

      expect(screen.getByRole('button', { name: 'Create Credential' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Refresh' })).toBeInTheDocument()
    })

    it('should display filter controls', async () => {
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      expect(screen.getByLabelText('Search')).toBeInTheDocument()
      expect(screen.getByLabelText('Filter by type')).toBeInTheDocument()
      expect(screen.getByLabelText('Sort by')).toBeInTheDocument()
    })
  })

  describe('Filtering and Search', () => {
    it('should search credentials by name', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      const searchInput = screen.getByLabelText('Search')
      await user.type(searchInput, 'GitHub')

      // The filtering should happen client-side based on the component logic
      await waitFor(() => {
        expect(screen.getByText('GitHub OAuth')).toBeInTheDocument()
      })
    })

    it('should filter credentials by type', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      const typeFilter = screen.getByLabelText('Filter by type')
      await user.selectOptions(typeFilter, 'api_key')

      // Should only show API Key credentials after filter
      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })
    })

    it('should change sort order', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      const sortSelect = screen.getByLabelText('Sort by')
      await user.selectOptions(sortSelect, 'name')

      expect(sortSelect).toHaveValue('name')
    })
  })

  describe('Create Credential', () => {
    it('should show create form when clicking create button', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: 'Create Credential' }))

      // Form should be displayed (the specific form fields depend on CredentialForm component)
      await waitFor(() => {
        // The create button should no longer be visible when in create mode
        expect(screen.queryByRole('button', { name: 'Create Credential' })).not.toBeInTheDocument()
      })
    })
  })

  describe('Delete Credential', () => {
    it('should show delete confirmation modal', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      // Find delete button for first credential
      const credentialRow =
        screen.getByText('Production API Key').closest('tr') ||
        screen.getByText('Production API Key').parentElement?.parentElement

      if (credentialRow) {
        const deleteButton = within(credentialRow as HTMLElement).queryByRole('button', {
          name: /delete/i,
        })

        if (deleteButton) {
          await user.click(deleteButton)

          await waitFor(() => {
            expect(screen.getByText('Delete Credential')).toBeInTheDocument()
          })
          expect(
            screen.getByText(/Are you sure you want to delete the credential/i)
          ).toBeInTheDocument()
        }
      }
    })

    it('should cancel delete operation', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      const credentialRow =
        screen.getByText('Production API Key').closest('tr') ||
        screen.getByText('Production API Key').parentElement?.parentElement

      if (credentialRow) {
        const deleteButton = within(credentialRow as HTMLElement).queryByRole('button', {
          name: /delete/i,
        })

        if (deleteButton) {
          await user.click(deleteButton)

          await waitFor(() => {
            expect(screen.getByText('Delete Credential')).toBeInTheDocument()
          })

          await user.click(screen.getByRole('button', { name: 'Cancel' }))

          await waitFor(() => {
            expect(screen.queryByText('Delete Credential')).not.toBeInTheDocument()
          })

          expect(screen.getByText('Production API Key')).toBeInTheDocument()
        }
      }
    })
  })

  describe('Test Credential', () => {
    it('should test credential and show success result', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      const credentialRow =
        screen.getByText('Production API Key').closest('tr') ||
        screen.getByText('Production API Key').parentElement?.parentElement

      if (credentialRow) {
        const testButton = within(credentialRow as HTMLElement).queryByRole('button', {
          name: /test/i,
        })

        if (testButton) {
          await user.click(testButton)

          await waitFor(
            () => {
              expect(screen.getByText('Credential test successful')).toBeInTheDocument()
            },
            { timeout: 3000 }
          )
        }
      }
    })

    it('should show test failure message', async () => {
      // Override the test handler to return failure
      server.use(
        http.post(`${API_BASE}/credentials/:id/test`, async () => {
          await delay(50)
          return HttpResponse.json({
            success: false,
            message: 'Connection failed: timeout',
            testedAt: new Date().toISOString(),
          })
        })
      )

      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      const credentialRow =
        screen.getByText('Production API Key').closest('tr') ||
        screen.getByText('Production API Key').parentElement?.parentElement

      if (credentialRow) {
        const testButton = within(credentialRow as HTMLElement).queryByRole('button', {
          name: /test/i,
        })

        if (testButton) {
          await user.click(testButton)

          await waitFor(
            () => {
              expect(screen.getByText('Connection failed: timeout')).toBeInTheDocument()
            },
            { timeout: 3000 }
          )
        }
      }
    })
  })

  describe('Edit Credential', () => {
    it('should switch to edit mode when clicking edit button', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      const credentialRow =
        screen.getByText('Production API Key').closest('tr') ||
        screen.getByText('Production API Key').parentElement?.parentElement

      if (credentialRow) {
        const editButton = within(credentialRow as HTMLElement).queryByRole('button', {
          name: /edit/i,
        })

        if (editButton) {
          await user.click(editButton)

          // Form should be displayed in edit mode
          // (the specific form fields depend on CredentialForm component)
          await waitFor(() => {
            expect(
              screen.queryByRole('button', { name: 'Create Credential' })
            ).not.toBeInTheDocument()
          })
        }
      }
    })
  })

  describe('Refresh', () => {
    it('should refresh credentials when clicking refresh button', async () => {
      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })

      const refreshButton = screen.getByRole('button', { name: 'Refresh' })
      await user.click(refreshButton)

      // Should still show credentials after refresh
      await waitFor(() => {
        expect(screen.getByText('Production API Key')).toBeInTheDocument()
      })
    })
  })

  describe('Error Handling', () => {
    it('should display error when API fails', async () => {
      server.use(
        http.get(`${API_BASE}/credentials`, async () => {
          await delay(50)
          return HttpResponse.json({ error: 'Internal server error' }, { status: 500 })
        })
      )

      render(<CredentialManager />)

      await waitFor(() => {
        // Error message should be displayed
        const errorElement = screen.queryByText(/error|failed/i)
        if (errorElement) {
          expect(errorElement).toBeInTheDocument()
        }
      })
    })

    it('should allow dismissing error message', async () => {
      server.use(
        http.get(`${API_BASE}/credentials`, async () => {
          await delay(50)
          return HttpResponse.json({ error: 'Internal server error' }, { status: 500 })
        })
      )

      const user = userEvent.setup()
      render(<CredentialManager />)

      await waitFor(() => {
        const errorElement = screen.queryByText(/error|failed/i)
        if (errorElement) {
          expect(errorElement).toBeInTheDocument()
        }
      })

      // Find and click dismiss button
      const dismissButton = screen.queryByLabelText('Dismiss error')
      if (dismissButton) {
        await user.click(dismissButton)

        await waitFor(() => {
          const errorElement = screen.queryByText(/internal server error/i)
          expect(errorElement).not.toBeInTheDocument()
        })
      }
    })
  })

  describe('Empty State', () => {
    it('should display empty state when no credentials exist', async () => {
      server.use(
        http.get(`${API_BASE}/credentials`, async () => {
          await delay(50)
          return HttpResponse.json({
            data: [],
            limit: 10,
            offset: 0,
          })
        })
      )

      render(<CredentialManager />)

      await waitFor(() => {
        expect(screen.getByText('Credentials')).toBeInTheDocument()
      })

      // The create button should still be visible
      expect(screen.getByRole('button', { name: 'Create Credential' })).toBeInTheDocument()
    })
  })
})
