import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { ThemeToggle } from './ThemeToggle';
import { useThemeStore } from '../model/theme';

describe('ThemeToggle', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.className = '';
    useThemeStore.setState({ theme: 'light' });
  });

  it('switches the theme when an option is clicked', async () => {
    const user = userEvent.setup();
    render(<ThemeToggle />);

    await user.click(screen.getByLabelText('Dark theme'));
    expect(useThemeStore.getState().theme).toBe('dark');
    expect(document.documentElement.classList.contains('dark')).toBe(true);

    await user.click(screen.getByLabelText('Custom theme'));
    expect(useThemeStore.getState().theme).toBe('custom');
    expect(document.documentElement.classList.contains('custom')).toBe(true);
  });

  it('marks the active theme with aria-pressed', () => {
    useThemeStore.setState({ theme: 'light' });
    render(<ThemeToggle />);
    expect(screen.getByLabelText('Light theme')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByLabelText('Dark theme')).toHaveAttribute('aria-pressed', 'false');
  });
});
