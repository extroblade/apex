import { describe, it, expect, beforeEach } from 'vitest';

import { applyTheme, readStoredTheme, useThemeStore } from './theme';

describe('theme', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.className = '';
    useThemeStore.setState({ theme: 'light' });
  });

  it('applyTheme toggles the html classes', () => {
    applyTheme('dark');
    expect(document.documentElement.classList.contains('dark')).toBe(true);

    applyTheme('custom');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
    expect(document.documentElement.classList.contains('custom')).toBe(true);

    applyTheme('light');
    expect(document.documentElement.classList.contains('custom')).toBe(false);
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('setTheme persists to localStorage and applies the class', () => {
    useThemeStore.getState().setTheme('dark');

    expect(localStorage.getItem('theme')).toBe('dark');
    expect(document.documentElement.classList.contains('dark')).toBe(true);
    expect(useThemeStore.getState().theme).toBe('dark');
  });

  it('readStoredTheme falls back to light for missing/invalid values', () => {
    expect(readStoredTheme()).toBe('light');
    localStorage.setItem('theme', 'nonsense');
    expect(readStoredTheme()).toBe('light');
    localStorage.setItem('theme', 'custom');
    expect(readStoredTheme()).toBe('custom');
  });
});
