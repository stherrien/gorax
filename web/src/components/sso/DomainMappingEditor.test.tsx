/**
 * Tests for DomainMappingEditor component
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DomainMappingEditor } from './DomainMappingEditor';

describe('DomainMappingEditor', () => {
  it('should render existing domains', () => {
    const domains = ['example.com', 'test.com'];
    render(<DomainMappingEditor domains={domains} onChange={vi.fn()} />);

    expect(screen.getByText('example.com')).toBeInTheDocument();
    expect(screen.getByText('test.com')).toBeInTheDocument();
  });

  it('should add a new domain', async () => {
    const onChange = vi.fn();
    render(<DomainMappingEditor domains={[]} onChange={onChange} />);

    const input = screen.getByPlaceholderText(/enter domain/i);
    await userEvent.type(input, 'newdomain.com');

    const addButton = screen.getByRole('button', { name: /add domain/i });
    await userEvent.click(addButton);

    expect(onChange).toHaveBeenCalledWith(['newdomain.com']);
  });

  it('should remove a domain', async () => {
    const onChange = vi.fn();
    const domains = ['example.com', 'test.com'];
    render(<DomainMappingEditor domains={domains} onChange={onChange} />);

    const removeButtons = screen.getAllByRole('button', { name: /remove/i });
    await userEvent.click(removeButtons[0]);

    expect(onChange).toHaveBeenCalledWith(['test.com']);
  });

  it('should validate domain format', async () => {
    const onChange = vi.fn();
    render(<DomainMappingEditor domains={[]} onChange={onChange} />);

    const input = screen.getByPlaceholderText(/enter domain/i);
    await userEvent.type(input, 'invalid domain');

    const addButton = screen.getByRole('button', { name: /add domain/i });
    await userEvent.click(addButton);

    expect(screen.getByText(/invalid domain format/i)).toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
  });

  it('should not allow duplicate domains', async () => {
    const onChange = vi.fn();
    const domains = ['example.com'];
    render(<DomainMappingEditor domains={domains} onChange={onChange} />);

    const input = screen.getByPlaceholderText(/enter domain/i);
    await userEvent.type(input, 'example.com');

    const addButton = screen.getByRole('button', { name: /add domain/i });
    await userEvent.click(addButton);

    expect(screen.getByText(/domain already exists/i)).toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
  });

  it('should clear input after adding domain', async () => {
    const onChange = vi.fn();
    render(<DomainMappingEditor domains={[]} onChange={onChange} />);

    const input = screen.getByPlaceholderText(/enter domain/i) as HTMLInputElement;
    await userEvent.type(input, 'example.com');

    const addButton = screen.getByRole('button', { name: /add domain/i });
    await userEvent.click(addButton);

    expect(input.value).toBe('');
  });

  it('should show empty state when no domains', () => {
    render(<DomainMappingEditor domains={[]} onChange={vi.fn()} />);

    expect(screen.getByText(/no domains configured/i)).toBeInTheDocument();
  });

  it('should normalize domain input', async () => {
    const onChange = vi.fn();
    render(<DomainMappingEditor domains={[]} onChange={onChange} />);

    const input = screen.getByPlaceholderText(/enter domain/i);
    await userEvent.type(input, '  EXAMPLE.COM  ');

    const addButton = screen.getByRole('button', { name: /add domain/i });
    await userEvent.click(addButton);

    expect(onChange).toHaveBeenCalledWith(['example.com']);
  });

  it('should disable add button when input is empty', () => {
    render(<DomainMappingEditor domains={[]} onChange={vi.fn()} />);

    const addButton = screen.getByRole('button', { name: /add domain/i });
    expect(addButton).toBeDisabled();
  });
});
