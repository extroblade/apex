import { describe, it, expect, beforeEach } from 'vitest';

import {
  readCustomVars,
  saveCustomVars,
  applyCustomVars,
  DEFAULT_CUSTOM_VARS,
} from './custom-vars';

describe('custom theme vars', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('style');
  });

  it('round-trips through localStorage with defaults for missing keys', () => {
    expect(readCustomVars()).toEqual(DEFAULT_CUSTOM_VARS);
    saveCustomVars({ ...DEFAULT_CUSTOM_VARS, primary: '#00ff00' });
    expect(readCustomVars().primary).toBe('#00ff00');
    expect(readCustomVars().font).toBe(DEFAULT_CUSTOM_VARS.font);
  });

  it('applies overrides only for the custom theme and clears them otherwise', () => {
    saveCustomVars({ ...DEFAULT_CUSTOM_VARS, primary: '#123456' });

    applyCustomVars('custom');
    expect(document.documentElement.style.getPropertyValue('--primary')).toBe('#123456');
    expect(
      document.documentElement.style.getPropertyValue('--app-font'),
    ).not.toBe('');

    applyCustomVars('light');
    expect(document.documentElement.style.getPropertyValue('--primary')).toBe('');
  });
});
