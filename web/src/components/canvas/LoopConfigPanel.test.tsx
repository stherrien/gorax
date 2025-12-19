import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import LoopConfigPanel from './LoopConfigPanel'

describe('LoopConfigPanel', () => {
  describe('Basic rendering', () => {
    it('should render loop configuration panel', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '',
        itemVariable: '',
        indexVariable: '',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      expect(screen.getByText(/Loop Configuration/i)).toBeInTheDocument()
    })

    it('should render all form fields', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '',
        itemVariable: '',
        indexVariable: '',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      expect(screen.getByLabelText(/Source Array/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Item Variable/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Index Variable/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Max Iterations/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Error Strategy/i)).toBeInTheDocument()
    })
  })

  describe('Form interactions', () => {
    it('should update source when input changes', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '',
        itemVariable: 'item',
        indexVariable: 'index',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const sourceInput = screen.getByLabelText(/Source Array/i) as HTMLInputElement
      fireEvent.change(sourceInput, { target: { value: '${steps.data.items}' } })

      expect(mockOnChange).toHaveBeenCalledWith({
        ...config,
        source: '${steps.data.items}',
      })
    })

    it('should update item variable when input changes', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '${steps.data.items}',
        itemVariable: '',
        indexVariable: 'index',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const itemVarInput = screen.getByLabelText(/Item Variable/i) as HTMLInputElement
      fireEvent.change(itemVarInput, { target: { value: 'currentItem' } })

      expect(mockOnChange).toHaveBeenCalledWith({
        ...config,
        itemVariable: 'currentItem',
      })
    })

    it('should update index variable when input changes', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '${steps.data.items}',
        itemVariable: 'item',
        indexVariable: '',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const indexVarInput = screen.getByLabelText(/Index Variable/i) as HTMLInputElement
      fireEvent.change(indexVarInput, { target: { value: 'idx' } })

      expect(mockOnChange).toHaveBeenCalledWith({
        ...config,
        indexVariable: 'idx',
      })
    })

    it('should update max iterations when input changes', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '${steps.data.items}',
        itemVariable: 'item',
        indexVariable: 'index',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const maxIterInput = screen.getByLabelText(/Max Iterations/i) as HTMLInputElement
      fireEvent.change(maxIterInput, { target: { value: '500' } })

      expect(mockOnChange).toHaveBeenCalledWith({
        ...config,
        maxIterations: 500,
      })
    })

    it('should update error strategy when select changes', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '${steps.data.items}',
        itemVariable: 'item',
        indexVariable: 'index',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const errorStrategySelect = screen.getByLabelText(/Error Strategy/i) as HTMLSelectElement
      fireEvent.change(errorStrategySelect, { target: { value: 'continue' } })

      expect(mockOnChange).toHaveBeenCalledWith({
        ...config,
        onError: 'continue',
      })
    })
  })

  describe('Initial values', () => {
    it('should display provided configuration values', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '${steps.http_request.output.data}',
        itemVariable: 'item',
        indexVariable: 'idx',
        maxIterations: 500,
        onError: 'continue' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      expect(screen.getByDisplayValue('${steps.http_request.output.data}')).toBeInTheDocument()
      expect(screen.getByDisplayValue('item')).toBeInTheDocument()
      expect(screen.getByDisplayValue('idx')).toBeInTheDocument()
      expect(screen.getByDisplayValue('500')).toBeInTheDocument()

      // For select, check the value attribute
      const errorStrategySelect = screen.getByLabelText(/Error Strategy/i) as HTMLSelectElement
      expect(errorStrategySelect.value).toBe('continue')
    })
  })

  describe('Validation hints', () => {
    it('should show source format hint', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '',
        itemVariable: '',
        indexVariable: '',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      expect(screen.getByText(/\${steps\..*\.output\..*}/i)).toBeInTheDocument()
    })

    it('should show variable naming hint', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '',
        itemVariable: '',
        indexVariable: '',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      expect(screen.getByText(/Variable name to access current item/i)).toBeInTheDocument()
    })
  })

  describe('Error strategy options', () => {
    it('should have stop and continue options in error strategy select', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '',
        itemVariable: '',
        indexVariable: '',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const errorStrategySelect = screen.getByLabelText(/Error Strategy/i)
      const options = errorStrategySelect.querySelectorAll('option')

      expect(options).toHaveLength(2)
      expect(options[0]).toHaveValue('stop')
      expect(options[1]).toHaveValue('continue')
    })
  })

  describe('Accessibility', () => {
    it('should have proper labels for all inputs', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '',
        itemVariable: '',
        indexVariable: '',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      const { container } = render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const labels = container.querySelectorAll('label')
      expect(labels.length).toBeGreaterThan(0)

      // Each input should have an associated label
      const inputs = container.querySelectorAll('input, select')
      inputs.forEach((input) => {
        expect(input).toHaveAttribute('id')
      })
    })

    it('should have descriptive help text', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '',
        itemVariable: '',
        indexVariable: '',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      // Should have helpful descriptions
      expect(screen.getByText(/current item in each iteration/i)).toBeInTheDocument()
      expect(screen.getByText(/handle errors in each iteration/i)).toBeInTheDocument()
    })
  })

  describe('Max iterations bounds', () => {
    it('should enforce minimum max iterations of 1', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '${steps.data.items}',
        itemVariable: 'item',
        indexVariable: 'index',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const maxIterInput = screen.getByLabelText(/Max Iterations/i) as HTMLInputElement
      expect(maxIterInput).toHaveAttribute('min', '1')
    })

    it('should enforce maximum max iterations of 10000', () => {
      const mockOnChange = vi.fn()
      const config = {
        source: '${steps.data.items}',
        itemVariable: 'item',
        indexVariable: 'index',
        maxIterations: 1000,
        onError: 'stop' as const,
      }

      render(<LoopConfigPanel config={config} onChange={mockOnChange} />)

      const maxIterInput = screen.getByLabelText(/Max Iterations/i) as HTMLInputElement
      expect(maxIterInput).toHaveAttribute('max', '10000')
    })
  })
})
