import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';

import { Button } from './button';

describe('Button', () => {
  it('renders its children', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument();
  });

  it('applies variant classes via tokens', () => {
    render(<Button variant="destructive">Delete</Button>);
    expect(screen.getByRole('button', { name: 'Delete' }).className).toContain(
      'bg-destructive',
    );
  });

  it('renders as its child element when asChild is set', () => {
    render(
      <Button asChild>
        <a href="/somewhere">A link</a>
      </Button>,
    );
    expect(screen.getByRole('link', { name: 'A link' })).toHaveAttribute(
      'href',
      '/somewhere',
    );
  });
});
