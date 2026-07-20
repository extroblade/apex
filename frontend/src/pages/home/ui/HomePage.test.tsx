import { describe, it, expect, afterEach } from 'vitest';
import { screen } from '@testing-library/react';

import { renderWithProviders } from '@/test/render';
import { i18n, setLanguage } from '@/shared/i18n';
import { ru } from '@/shared/i18n/locales/ru';
import { HomePage } from './HomePage';

describe('HomePage', () => {
  afterEach(() => setLanguage('en'));

  it('renders the brand and the signed-out tagline in English', () => {
    renderWithProviders(<HomePage />);
    expect(screen.getByRole('heading', { name: 'ContentPilot' })).toBeInTheDocument();
    expect(
      screen.getByText('Plan races and track your iRacing stats.'),
    ).toBeInTheDocument();
  });

  it('renders the tagline in Russian after switching language', async () => {
    // ru is backend-served now (not bundled); inject its authored bundle so the
    // switch resolves offline instead of hitting the network.
    i18n.addResourceBundle('ru', 'translation', ru, true, true);
    await setLanguage('ru');
    renderWithProviders(<HomePage />);
    expect(
      await screen.findByText('Планируйте гонки и следите за статистикой iRacing.'),
    ).toBeInTheDocument();
  });
});
