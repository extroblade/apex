import { describe, it, expect, afterEach } from 'vitest';

import i18n, { setLanguage } from './config';

describe('i18n', () => {
  afterEach(() => setLanguage('en'));

  it('provides English by default', () => {
    expect(i18n.t('nav.home')).toBe('Home');
  });

  it('switches to Russian', () => {
    setLanguage('ru');
    expect(i18n.language).toBe('ru');
    expect(i18n.t('nav.home')).toBe('Главная');
    expect(i18n.t('nav.drivers')).toBe('Пилоты');
  });

  it('interpolates values', () => {
    setLanguage('en');
    expect(i18n.t('home.signedInAs', { email: 'a@b.com' })).toBe('Signed in as a@b.com.');
  });
});
