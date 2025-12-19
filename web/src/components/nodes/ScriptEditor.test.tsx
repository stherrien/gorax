import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ScriptEditor } from './ScriptEditor';

describe('ScriptEditor', () => {
  it('renders with initial script value', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="return 42;" onChange={onChange} />);

    const textarea = screen.getByTestId('script-input') as HTMLTextAreaElement;
    expect(textarea.value).toBe('return 42;');
  });

  it('calls onChange when script changes', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} />);

    const textarea = screen.getByTestId('script-input');
    fireEvent.change(textarea, { target: { value: 'return "Hello";' } });

    expect(onChange).toHaveBeenCalledWith('return "Hello";');
  });

  it('renders timeout input with default value', () => {
    const onChange = vi.fn();
    const onTimeoutChange = vi.fn();
    render(
      <ScriptEditor
        value=""
        onChange={onChange}
        timeout={30}
        onTimeoutChange={onTimeoutChange}
      />
    );

    const timeoutInput = screen.getByTestId('timeout-input') as HTMLInputElement;
    expect(timeoutInput.value).toBe('30');
  });

  it('calls onTimeoutChange when timeout value changes', () => {
    const onChange = vi.fn();
    const onTimeoutChange = vi.fn();
    render(
      <ScriptEditor
        value=""
        onChange={onChange}
        timeout={30}
        onTimeoutChange={onTimeoutChange}
      />
    );

    const timeoutInput = screen.getByTestId('timeout-input');
    fireEvent.change(timeoutInput, { target: { value: '60' } });

    expect(onTimeoutChange).toHaveBeenCalledWith(60);
  });

  it('displays error message when error prop is provided', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} error="Syntax error" />);

    const error = screen.getByTestId('script-error');
    expect(error).toHaveTextContent('Syntax error');
  });

  it('applies error styling when error is present', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} error="Invalid" />);

    const textarea = screen.getByTestId('script-input');
    expect(textarea.className).toContain('border-red-500');
  });

  it('shows examples when Show Examples button is clicked', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} />);

    const showButton = screen.getByText('Show Examples');
    fireEvent.click(showButton);

    expect(screen.getByText('Example Scripts')).toBeInTheDocument();
    expect(screen.getByText(/Simple calculation/)).toBeInTheDocument();
  });

  it('hides examples when Hide Examples button is clicked', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} />);

    // Show examples first
    const showButton = screen.getByText('Show Examples');
    fireEvent.click(showButton);
    expect(screen.getByText('Example Scripts')).toBeInTheDocument();

    // Hide examples
    const hideButton = screen.getByText('Hide Examples');
    fireEvent.click(hideButton);
    expect(screen.queryByText('Example Scripts')).not.toBeInTheDocument();
  });

  it('shows API help when Show API button is clicked', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} />);

    const showButton = screen.getByText('Show API');
    fireEvent.click(showButton);

    expect(screen.getByText('Available JavaScript API')).toBeInTheDocument();
    expect(screen.getByText('Access data from the workflow trigger')).toBeInTheDocument();
  });

  it('applies example script when Use button is clicked', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} />);

    // Show examples
    fireEvent.click(screen.getByText('Show Examples'));

    // Click first Use button
    const useButtons = screen.getAllByText('Use');
    fireEvent.click(useButtons[0]);

    expect(onChange).toHaveBeenCalled();
    // Check that onChange was called with a non-empty string
    expect(onChange.mock.calls[0][0]).toBeTruthy();
  });

  it('displays placeholder text', () => {
    const onChange = vi.fn();
    const placeholder = 'Enter your JavaScript code';
    render(<ScriptEditor value="" onChange={onChange} placeholder={placeholder} />);

    const textarea = screen.getByTestId('script-input');
    expect(textarea).toHaveAttribute('placeholder', placeholder);
  });

  it('displays available variables from context', () => {
    const onChange = vi.fn();
    const context = {
      trigger: { name: 'webhook', id: 123 },
      steps: { step1: { result: 'success' } },
    };
    render(<ScriptEditor value="" onChange={onChange} context={context} />);

    const summary = screen.getByText('Available Context');
    fireEvent.click(summary);

    expect(screen.getByText(/"trigger"/)).toBeInTheDocument();
    expect(screen.getByText(/"steps"/)).toBeInTheDocument();
  });

  it('does not display context section when context is empty', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} context={{}} />);

    expect(screen.queryByText('Available Context')).not.toBeInTheDocument();
  });

  it('displays security notice', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="" onChange={onChange} />);

    expect(screen.getByText(/Sandboxed Environment/)).toBeInTheDocument();
    expect(screen.getByText(/No file system or network access/)).toBeInTheDocument();
  });

  it('renders with monospace font for code', () => {
    const onChange = vi.fn();
    render(<ScriptEditor value="return 42;" onChange={onChange} />);

    const textarea = screen.getByTestId('script-input');
    expect(textarea.className).toContain('font-mono');
  });

  it('handles empty timeout gracefully', () => {
    const onChange = vi.fn();
    const onTimeoutChange = vi.fn();
    render(
      <ScriptEditor
        value=""
        onChange={onChange}
        onTimeoutChange={onTimeoutChange}
      />
    );

    const timeoutInput = screen.getByTestId('timeout-input');
    fireEvent.change(timeoutInput, { target: { value: '' } });

    // Should call with 0 or not throw error
    expect(onTimeoutChange).toHaveBeenCalled();
  });
});
