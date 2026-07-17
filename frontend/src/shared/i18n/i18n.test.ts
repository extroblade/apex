import { describe, it, expect, afterEach, vi } from 'vitest';

import i18n, { setLanguage } from './config';

describe('i18n', () => {
  afterEach(async () => {
    await setLanguage('en'); // bundled → no fetch
    vi.unstubAllGlobals();
  });

  it('provides bundled English by default', () => {
    expect(i18n.t('nav.home')).toBe('Home');
  });

  it('interpolates values', () => {
    expect(i18n.t('home.signedInAs', { email: 'a@b.com' })).toBe('Signed in as a@b.com.');
  });

  it('fetches a non-bundled language bundle from the backend', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ nav: { home: 'Accueil' } }),
      }),
    );

    await setLanguage('fr');

    expect(fetch).toHaveBeenCalledWith('/api/locales/fr');
    expect(i18n.language).toBe('fr');
    expect(i18n.t('nav.home')).toBe('Accueil');
    // A key missing from the fetched bundle falls back to English.
    expect(i18n.t('nav.fuel')).toBe('Fuel');
  });

  it('fetches each language bundle only once', async () => {
    const f = vi
      .fn()
      .mockResolvedValue({ ok: true, json: async () => ({ nav: { home: 'Hola' } }) });
    vi.stubGlobal('fetch', f);

    await setLanguage('es');
    await setLanguage('en'); // bundled
    await setLanguage('es'); // already loaded → no refetch

    expect(f).toHaveBeenCalledTimes(1);
  });
});
