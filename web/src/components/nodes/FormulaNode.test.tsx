import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ReactFlowProvider } from '@xyflow/react';
import FormulaNode from './FormulaNode';

// Wrapper component that provides React Flow context
const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <ReactFlowProvider>{children}</ReactFlowProvider>
);

describe('FormulaNode', () => {
  it('renders with default label', () => {
    const data = {
      label: 'Calculate Total',
      actionType: 'formula' as const,
      config: {
        expression: 'x + y',
      },
    };

    render(<FormulaNode data={data} />, { wrapper: Wrapper });

    expect(screen.getByText('Calculate Total')).toBeInTheDocument();
  });

  it('renders with default "Formula" label when no label provided', () => {
    const data = {
      label: '',
      actionType: 'formula' as const,
    };

    render(<FormulaNode data={data} />, { wrapper: Wrapper });

    expect(screen.getByText('Formula')).toBeInTheDocument();
  });

  it('displays formula expression', () => {
    const data = {
      label: 'Test',
      actionType: 'formula' as const,
      config: {
        expression: 'upper(name)',
      },
    };

    render(<FormulaNode data={data} />, { wrapper: Wrapper });

    expect(screen.getByText('upper(name)')).toBeInTheDocument();
  });

  it('truncates long expressions', () => {
    const longExpression = 'round((price * quantity) * (1 + taxRate) + shippingCost)';
    const data = {
      label: 'Test',
      actionType: 'formula' as const,
      config: {
        expression: longExpression,
      },
    };

    render(<FormulaNode data={data} />, { wrapper: Wrapper });

    // Should be truncated with ellipsis
    const expressionElement = screen.getByTitle(longExpression);
    expect(expressionElement.textContent).toContain('...');
    expect(expressionElement.textContent?.length).toBeLessThan(longExpression.length);
  });

  it('shows "No formula set" when expression is empty', () => {
    const data = {
      label: 'Test',
      actionType: 'formula' as const,
      config: {},
    };

    render(<FormulaNode data={data} />, { wrapper: Wrapper });

    expect(screen.getByText('No formula set')).toBeInTheDocument();
  });

  it('applies selected styling when selected', () => {
    const data = {
      label: 'Test',
      actionType: 'formula' as const,
    };

    render(<FormulaNode data={data} selected={true} />, { wrapper: Wrapper });

    const node = screen.getByTestId('formula-node');
    expect(node.className).toContain('ring-2');
  });

  it('does not apply selected styling when not selected', () => {
    const data = {
      label: 'Test',
      actionType: 'formula' as const,
    };

    render(<FormulaNode data={data} selected={false} />, { wrapper: Wrapper });

    const node = screen.getByTestId('formula-node');
    expect(node.className).not.toContain('ring-2');
  });

  it('displays formula icon', () => {
    const data = {
      label: 'Test',
      actionType: 'formula' as const,
    };

    render(<FormulaNode data={data} />, { wrapper: Wrapper });

    expect(screen.getByText('🔢')).toBeInTheDocument();
  });

  it('has input and output handles', () => {
    const data = {
      label: 'Test',
      actionType: 'formula' as const,
    };

    const { container } = render(<FormulaNode data={data} />, { wrapper: Wrapper });

    // Check for Handle components (they render as divs with specific classes)
    const handles = container.querySelectorAll('.react-flow__handle');
    expect(handles.length).toBeGreaterThanOrEqual(2); // Should have at least input and output handles
  });
});
