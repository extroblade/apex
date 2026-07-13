import { describe, it, expect } from 'vitest';

import { initials } from './initials';

describe('initials', () => {
  it('takes the first two letters of a single name', () => {
    expect(initials('Speedy')).toBe('SP');
  });

  it('combines the first letters of two words', () => {
    expect(initials('Ada Lovelace')).toBe('AL');
    expect(initials('ada.lovelace')).toBe('AL');
  });

  it('uses the local part of an email', () => {
    expect(initials('pilot@example.com')).toBe('PI');
  });

  it('falls back to ? for empty input', () => {
    expect(initials('   ')).toBe('?');
  });
});
