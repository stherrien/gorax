import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import SubWorkflowNode from './SubWorkflowNode'
import { workflowAPI } from '../../api/workflows'

// Mock the workflows API
vi.mock('../../api/workflows', () => ({
  workflowAPI: {
    get: vi.fn(),
  },
}))

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  Handle: ({ type, position }: { type: string; position: string }) => (
    <div data-testid={`handle-${type}-${position}`} />
  ),
  Position: {
    Top: 'top',
    Bottom: 'bottom',
  },
}))

describe('SubWorkflowNode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders with basic data', () => {
    const data = {
      label: 'Execute Sub-Workflow',
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByText('Execute Sub-Workflow')).toBeInTheDocument()
    expect(screen.getByText('Sub-Workflow')).toBeInTheDocument()
  })

  it('displays workflow name when provided', () => {
    const data = {
      label: 'Execute Sub-Workflow',
      workflowId: 'wf-123',
      workflowName: 'Email Notification Workflow',
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByText('Email Notification Workflow')).toBeInTheDocument()
  })

  it('fetches workflow name when ID is provided but name is not', async () => {
    const mockWorkflow = {
      id: 'wf-123',
      name: 'Fetched Workflow',
      status: 'active',
      definition: { nodes: [], edges: [] },
    }

    vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)

    const data = {
      label: 'Execute Sub-Workflow',
      workflowId: 'wf-123',
    }

    render(<SubWorkflowNode data={data} />)

    await waitFor(() => {
      expect(workflowAPI.get).toHaveBeenCalledWith('wf-123')
    })

    await waitFor(() => {
      expect(screen.getByText('Fetched Workflow')).toBeInTheDocument()
    })
  })

  it('displays "Unknown Workflow" on fetch error', async () => {
    vi.mocked(workflowAPI.get).mockRejectedValue(new Error('Not found'))

    const data = {
      label: 'Execute Sub-Workflow',
      workflowId: 'wf-invalid',
    }

    render(<SubWorkflowNode data={data} />)

    await waitFor(() => {
      expect(screen.getByText('Unknown Workflow')).toBeInTheDocument()
    })
  })

  it('displays input mapping count', () => {
    const data = {
      label: 'Execute Sub-Workflow',
      inputMapping: {
        userId: '${trigger.user.id}',
        email: '${trigger.user.email}',
        orderData: '${steps.processOrder.output}',
      },
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByText('Inputs:')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('displays output mapping count', () => {
    const data = {
      label: 'Execute Sub-Workflow',
      outputMapping: {
        result: '${output.status}',
        notificationId: '${output.id}',
      },
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByText('Outputs:')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('displays synchronous execution mode', () => {
    const data = {
      label: 'Execute Sub-Workflow',
      waitForResult: true,
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByText('Mode:')).toBeInTheDocument()
    expect(screen.getByText('Sync')).toBeInTheDocument()
  })

  it('displays asynchronous execution mode', () => {
    const data = {
      label: 'Execute Sub-Workflow',
      waitForResult: false,
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByText('Mode:')).toBeInTheDocument()
    expect(screen.getByText('Async')).toBeInTheDocument()
  })

  it('displays timeout when configured', () => {
    const data = {
      label: 'Execute Sub-Workflow',
      waitForResult: true,
      timeoutMs: 5000,
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByText('Timeout:')).toBeInTheDocument()
    expect(screen.getByText('5000ms')).toBeInTheDocument()
  })

  it('does not display timeout when not configured', () => {
    const data = {
      label: 'Execute Sub-Workflow',
      waitForResult: true,
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.queryByText('Timeout:')).not.toBeInTheDocument()
  })

  it('applies selected styling when selected', () => {
    const data = {
      label: 'Execute Sub-Workflow',
    }

    const { container } = render(<SubWorkflowNode data={data} selected={true} />)

    const nodeElement = container.querySelector('.ring-2')
    expect(nodeElement).toBeInTheDocument()
  })

  it('does not apply selected styling when not selected', () => {
    const data = {
      label: 'Execute Sub-Workflow',
    }

    const { container } = render(<SubWorkflowNode data={data} selected={false} />)

    const nodeElement = container.querySelector('.ring-2')
    expect(nodeElement).not.toBeInTheDocument()
  })

  it('renders both input and output handles', () => {
    const data = {
      label: 'Execute Sub-Workflow',
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByTestId('handle-target-top')).toBeInTheDocument()
    expect(screen.getByTestId('handle-source-bottom')).toBeInTheDocument()
  })

  it('displays comprehensive configuration', () => {
    const data = {
      label: 'Execute Notification Workflow',
      workflowId: 'wf-notify-123',
      workflowName: 'Send Notification',
      inputMapping: {
        userId: '${trigger.user.id}',
        message: '${steps.prepare.message}',
      },
      outputMapping: {
        notificationId: '${output.id}',
      },
      waitForResult: true,
      timeoutMs: 10000,
    }

    render(<SubWorkflowNode data={data} />)

    expect(screen.getByText('Execute Notification Workflow')).toBeInTheDocument()
    expect(screen.getByText('Send Notification')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument() // Input count
    expect(screen.getByText('1')).toBeInTheDocument() // Output count
    expect(screen.getByText('Sync')).toBeInTheDocument()
    expect(screen.getByText('10000ms')).toBeInTheDocument()
  })
})
