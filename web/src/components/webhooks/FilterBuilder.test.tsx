import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import FilterBuilder from './FilterBuilder'
import { webhookAPI } from '../../api/webhooks'
import type { WebhookFilter } from '../../api/webhooks'

vi.mock('../../api/webhooks', () => ({
  webhookAPI: {
    getFilters: vi.fn(),
    createFilter: vi.fn(),
    updateFilter: vi.fn(),
    deleteFilter: vi.fn(),
    testFilters: vi.fn(),
  },
}))

describe('FilterBuilder', () => {
  const mockWebhookId = 'webhook-123'
  const mockFilters: WebhookFilter[] = [
    {
      id: 'filter-1',
      webhookId: mockWebhookId,
      fieldPath: '$.data.status',
      operator: 'equals',
      value: 'active',
      logicGroup: 0,
      enabled: true,
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
    {
      id: 'filter-2',
      webhookId: mockWebhookId,
      fieldPath: '$.data.amount',
      operator: 'gt',
      value: 100,
      logicGroup: 0,
      enabled: true,
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Initial Rendering', () => {
    it('renders filter builder with title', async () => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: [],
        total: 0,
      })

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByText('Webhook Filters')).toBeInTheDocument()
      })
    })

    it('loads and displays existing filters', async () => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: mockFilters,
        total: mockFilters.length,
      })

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByDisplayValue('$.data.status')).toBeInTheDocument()
        expect(screen.getByDisplayValue('$.data.amount')).toBeInTheDocument()
      })
    })

    it('displays loading state while fetching filters', () => {
      vi.mocked(webhookAPI.getFilters).mockImplementation(
        () => new Promise(() => {})
      )

      render(<FilterBuilder webhookId={mockWebhookId} />)

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('displays error message when fetching filters fails', async () => {
      vi.mocked(webhookAPI.getFilters).mockRejectedValue(
        new Error('Failed to fetch filters')
      )

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByText(/failed to fetch filters/i)).toBeInTheDocument()
      })
    })
  })

  describe('Adding Filters', () => {
    beforeEach(() => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: [],
        total: 0,
      })
    })

    it('shows add filter button', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add filter/i })).toBeInTheDocument()
      })
    })

    it('adds a new empty filter rule when add button is clicked', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add filter/i })).toBeInTheDocument()
      })

      const addButton = screen.getByRole('button', { name: /add filter/i })
      fireEvent.click(addButton)

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/\$\.data\.status/i)).toBeInTheDocument()
      })
    })

    it('creates filter on server when save is clicked', async () => {
      const newFilter = {
        id: 'filter-new',
        webhookId: mockWebhookId,
        fieldPath: '$.event.type',
        operator: 'equals' as const,
        value: 'payment',
        logicGroup: 0,
        enabled: true,
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
      }

      vi.mocked(webhookAPI.createFilter).mockResolvedValue(newFilter)

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add filter/i })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: /add filter/i }))

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/\$\.data\.status/i)).toBeInTheDocument()
      })

      const fieldPathInput = screen.getByPlaceholderText(/\$\.data\.status/i)
      fireEvent.change(fieldPathInput, { target: { value: '$.event.type' } })

      const valueInput = screen.getByPlaceholderText(/value/i)
      fireEvent.change(valueInput, { target: { value: 'payment' } })

      const saveButton = screen.getByRole('button', { name: /save/i })
      fireEvent.click(saveButton)

      await waitFor(() => {
        expect(webhookAPI.createFilter).toHaveBeenCalledWith(mockWebhookId, {
          fieldPath: '$.event.type',
          operator: 'equals',
          value: 'payment',
          logicGroup: 0,
          enabled: true,
        })
      })
    })
  })

  describe('Editing Filters', () => {
    beforeEach(() => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: mockFilters,
        total: mockFilters.length,
      })
    })

    it('allows editing field path', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        const fieldPathInput = screen.getByDisplayValue('$.data.status')
        fireEvent.change(fieldPathInput, { target: { value: '$.data.newField' } })
        expect(fieldPathInput).toHaveValue('$.data.newField')
      })
    })

    it('allows changing operator', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        const operatorSelects = screen.getAllByRole('combobox')
        const firstOperatorSelect = operatorSelects.find(
          select => (select as HTMLSelectElement).value === 'equals'
        )
        expect(firstOperatorSelect).toBeInTheDocument()
        if (firstOperatorSelect) {
          fireEvent.change(firstOperatorSelect, { target: { value: 'contains' } })
          expect(firstOperatorSelect).toHaveValue('contains')
        }
      })
    })

    it('updates filter on server when save is clicked', async () => {
      const updatedFilter = { ...mockFilters[0], value: 'inactive' }
      vi.mocked(webhookAPI.updateFilter).mockResolvedValue(updatedFilter)

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(async () => {
        const valueInputs = screen.getAllByDisplayValue('active')
        fireEvent.change(valueInputs[0], { target: { value: 'inactive' } })

        const saveButtons = screen.getAllByRole('button', { name: /save/i })
        fireEvent.click(saveButtons[0])

        await waitFor(() => {
          expect(webhookAPI.updateFilter).toHaveBeenCalledWith(
            mockWebhookId,
            'filter-1',
            expect.objectContaining({ value: 'inactive' })
          )
        })
      })
    })
  })

  describe('Deleting Filters', () => {
    beforeEach(() => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: mockFilters,
        total: mockFilters.length,
      })
    })

    it('shows delete button for each filter', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
        expect(deleteButtons).toHaveLength(mockFilters.length)
      })
    })

    it('deletes filter from server when delete is clicked', async () => {
      vi.mocked(webhookAPI.deleteFilter).mockResolvedValue()

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(async () => {
        const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
        fireEvent.click(deleteButtons[0])

        await waitFor(() => {
          expect(webhookAPI.deleteFilter).toHaveBeenCalledWith(
            mockWebhookId,
            'filter-1'
          )
        })
      })
    })
  })

  describe('Operator Support', () => {
    beforeEach(() => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: [],
        total: 0,
      })
    })

    it('displays all supported operators in dropdown', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add filter/i })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: /add filter/i }))

      await waitFor(() => {
        expect(screen.getByLabelText(/operator/i)).toBeInTheDocument()
      })

      const operatorSelect = screen.getByLabelText(/operator/i)
      const options = Array.from(operatorSelect.querySelectorAll('option'))

      const expectedOperators = [
        'equals',
        'not_equals',
        'contains',
        'not_contains',
        'starts_with',
        'ends_with',
        'regex',
        'gt',
        'gte',
        'lt',
        'lte',
        'in',
        'not_in',
        'exists',
        'not_exists',
      ]

      expectedOperators.forEach(op => {
        expect(options.some(option => option.value === op)).toBe(true)
      })
    })
  })

  describe('Logic Groups', () => {
    it('allows setting logic group number', async () => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: [],
        total: 0,
      })

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add filter/i })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: /add filter/i }))

      await waitFor(() => {
        expect(screen.getByLabelText(/logic group/i)).toBeInTheDocument()
      })

      const logicGroupInput = screen.getByLabelText(/logic group/i)
      fireEvent.change(logicGroupInput, { target: { value: '1' } })

      expect(logicGroupInput).toHaveValue(1)
    })

    it('displays visual indication of AND/OR logic', async () => {
      const filtersWithGroups: WebhookFilter[] = [
        { ...mockFilters[0], logicGroup: 0 },
        { ...mockFilters[1], logicGroup: 0 },
        {
          ...mockFilters[0],
          id: 'filter-3',
          logicGroup: 1,
          fieldPath: '$.event.type',
        },
      ]

      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: filtersWithGroups,
        total: filtersWithGroups.length,
      })

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByText(/AND/)).toBeInTheDocument()
        expect(screen.getByText(/OR/)).toBeInTheDocument()
      })
    })
  })

  describe('Test Filter', () => {
    beforeEach(() => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: mockFilters,
        total: mockFilters.length,
      })
    })

    it('shows test filter button', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByText(/test filters/i)).toBeInTheDocument()
      })
    })

    it('allows testing filters with sample payload', async () => {
      vi.mocked(webhookAPI.testFilters).mockResolvedValue({
        passed: true,
        reason: 'All filters passed',
        details: {},
      })

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByText(/test filters/i)).toBeInTheDocument()
      })

      const testButton = screen.getByText(/test filters/i)
      fireEvent.click(testButton)

      await waitFor(() => {
        expect(screen.getByLabelText(/test payload/i)).toBeInTheDocument()
      })
    })

    it('displays test results', async () => {
      vi.mocked(webhookAPI.testFilters).mockResolvedValue({
        passed: true,
        reason: 'All filters passed',
        details: {},
      })

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByText(/test filters/i)).toBeInTheDocument()
      })

      const testButton = screen.getByText(/test filters/i)
      fireEvent.click(testButton)

      await waitFor(() => {
        expect(screen.getByLabelText(/test payload/i)).toBeInTheDocument()
      })

      const payloadInput = screen.getByLabelText(/test payload/i)
      fireEvent.change(payloadInput, {
        target: { value: '{"data":{"status":"active","amount":200}}' },
      })

      const runTestButton = screen.getByRole('button', { name: /run test/i })
      fireEvent.click(runTestButton)

      await waitFor(() => {
        expect(screen.getByText('All filters passed')).toBeInTheDocument()
      })
    })

    it('displays error for invalid JSON in test payload', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByText(/test filters/i)).toBeInTheDocument()
      })

      const testButton = screen.getByText(/test filters/i)
      fireEvent.click(testButton)

      await waitFor(() => {
        expect(screen.getByLabelText(/test payload/i)).toBeInTheDocument()
      })

      const payloadInput = screen.getByLabelText(/test payload/i)
      fireEvent.change(payloadInput, { target: { value: 'not valid json syntax' } })

      const runTestButton = screen.getByRole('button', { name: /run test/i })
      fireEvent.click(runTestButton)

      await waitFor(() => {
        expect(screen.getByText('Invalid JSON format')).toBeInTheDocument()
      })
    })
  })

  describe('Validation', () => {
    beforeEach(() => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: [],
        total: 0,
      })
    })

    it('shows error for empty field path', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add filter/i })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: /add filter/i }))

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
      })

      const fieldPathInput = screen.getByPlaceholderText(/\$\.data\.status/i)
      fireEvent.change(fieldPathInput, { target: { value: '' } })

      const saveButton = screen.getByRole('button', { name: /save/i })
      fireEvent.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText(/field path is required/i)).toBeInTheDocument()
      })
    })

    it('shows error for invalid JSON path format', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add filter/i })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: /add filter/i }))

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/\$\.data\.status/i)).toBeInTheDocument()
      })

      const fieldPathInput = screen.getByPlaceholderText(/\$\.data\.status/i)
      fireEvent.change(fieldPathInput, { target: { value: 'invalid path' } })

      const saveButton = screen.getByRole('button', { name: /save/i })
      fireEvent.click(saveButton)

      await waitFor(() => {
        expect(
          screen.getByText(/field path must start with/i)
        ).toBeInTheDocument()
      })
    })

    it('validates regex patterns for regex operator', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add filter/i })).toBeInTheDocument()
      })

      fireEvent.click(screen.getByRole('button', { name: /add filter/i }))

      await waitFor(() => {
        expect(screen.getByLabelText(/operator/i)).toBeInTheDocument()
      })

      const fieldPathInput = screen.getByPlaceholderText(/\$\.data\.status/i)
      fireEvent.change(fieldPathInput, { target: { value: '$.test' } })

      const operatorSelect = screen.getByLabelText(/operator/i)
      fireEvent.change(operatorSelect, { target: { value: 'regex' } })

      const valueInput = screen.getByPlaceholderText(/value/i)
      fireEvent.change(valueInput, { target: { value: '[invalid(regex' } })

      const saveButton = screen.getByRole('button', { name: /save/i })
      fireEvent.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText('Invalid regex pattern')).toBeInTheDocument()
      })
    })
  })

  describe('Enable/Disable Filters', () => {
    beforeEach(() => {
      vi.mocked(webhookAPI.getFilters).mockResolvedValue({
        filters: mockFilters,
        total: mockFilters.length,
      })
    })

    it('shows enabled toggle for each filter', async () => {
      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(() => {
        const toggles = screen.getAllByRole('checkbox', { name: /enabled/i })
        expect(toggles).toHaveLength(mockFilters.length)
      })
    })

    it('updates filter enabled state', async () => {
      const updatedFilter = { ...mockFilters[0], enabled: false }
      vi.mocked(webhookAPI.updateFilter).mockResolvedValue(updatedFilter)

      render(<FilterBuilder webhookId={mockWebhookId} />)

      await waitFor(async () => {
        const toggles = screen.getAllByRole('checkbox', { name: /enabled/i })
        fireEvent.click(toggles[0])

        await waitFor(() => {
          expect(webhookAPI.updateFilter).toHaveBeenCalledWith(
            mockWebhookId,
            'filter-1',
            expect.objectContaining({ enabled: false })
          )
        })
      })
    })
  })
})
