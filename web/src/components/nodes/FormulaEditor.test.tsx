import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { FormulaEditor } from './FormulaEditor';

describe('FormulaEditor', () => {
  it('renders with initial value', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="x + y" onChange={onChange} />);

    const input = screen.getByTestId('formula-input') as HTMLTextAreaElement;
    expect(input.value).toBe('x + y');
  });

  it('calls onChange when value changes', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} />);

    const input = screen.getByTestId('formula-input');
    fireEvent.change(input, { target: { value: 'upper("hello")' } });

    expect(onChange).toHaveBeenCalledWith('upper("hello")');
  });

  it('displays error message when error prop is provided', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} error="Invalid expression" />);

    const error = screen.getByTestId('formula-error');
    expect(error).toHaveTextContent('Invalid expression');
  });

  it('applies error styling when error is present', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} error="Invalid" />);

    const input = screen.getByTestId('formula-input');
    expect(input.className).toContain('border-red-500');
  });

  it('shows examples when Show Examples button is clicked', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} />);

    const showButton = screen.getByText('Show Examples');
    fireEvent.click(showButton);

    expect(screen.getByText('Example Formulas')).toBeInTheDocument();
    expect(screen.getByText(/String manipulation/)).toBeInTheDocument();
  });

  it('hides examples when Hide Examples button is clicked', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} />);

    // Show examples first
    const showButton = screen.getByText('Show Examples');
    fireEvent.click(showButton);
    expect(screen.getByText('Example Formulas')).toBeInTheDocument();

    // Hide examples
    const hideButton = screen.getByText('Hide Examples');
    fireEvent.click(hideButton);
    expect(screen.queryByText('Example Formulas')).not.toBeInTheDocument();
  });

  it('shows function help when Show Functions button is clicked', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} />);

    const showButton = screen.getByText('Show Functions');
    fireEvent.click(showButton);

    expect(screen.getByText('Available Functions')).toBeInTheDocument();
    expect(screen.getByText('upper()')).toBeInTheDocument();
    expect(screen.getByText(/Converts string to uppercase/)).toBeInTheDocument();
  });

  it('applies example formula when Use button is clicked', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} />);

    // Show examples
    fireEvent.click(screen.getByText('Show Examples'));

    // Click first Use button
    const useButtons = screen.getAllByText('Use');
    fireEvent.click(useButtons[0]);

    expect(onChange).toHaveBeenCalledWith('upper(trim(name))');
  });

  it('displays placeholder text', () => {
    const onChange = vi.fn();
    const placeholder = 'Enter your formula';
    render(<FormulaEditor value="" onChange={onChange} placeholder={placeholder} />);

    const input = screen.getByTestId('formula-input');
    expect(input).toHaveAttribute('placeholder', placeholder);
  });

  it('displays available variables from context', () => {
    const onChange = vi.fn();
    const context = {
      x: 10,
      user: { name: 'John' },
    };
    render(<FormulaEditor value="" onChange={onChange} context={context} />);

    const summary = screen.getByText('Available Variables');
    fireEvent.click(summary);

    expect(screen.getByText(/"x": 10/)).toBeInTheDocument();
    expect(screen.getByText(/"name": "John"/)).toBeInTheDocument();
  });

  it('does not display variables section when context is empty', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} context={{}} />);

    expect(screen.queryByText('Available Variables')).not.toBeInTheDocument();
  });

  it('displays all built-in functions in help section', () => {
    const onChange = vi.fn();
    render(<FormulaEditor value="" onChange={onChange} />);

    fireEvent.click(screen.getByText('Show Functions'));

    // Check for some key functions
    expect(screen.getByText('upper()')).toBeInTheDocument();
    expect(screen.getByText('lower()')).toBeInTheDocument();
    expect(screen.getByText('round()')).toBeInTheDocument();
    expect(screen.getByText('len()')).toBeInTheDocument();
  });
});
