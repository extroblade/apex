import { describe, it, expect, afterEach } from 'vitest';
import { screen } from '@testing-library/react';

import { renderWithProviders } from '@/test/render';
import { setLanguage } from '@/shared/i18n';
import { HomePage } from './HomePage';

describe('HomePage', () => {
  afterEach(() => setLanguage('en'));

  it('renders the brand and the signed-out tagline in English', () => {
    renderWithProviders(<HomePage />);
    expect(screen.getByRole('heading', { name: 'Apex' })).toBeInTheDocument();
    expect(
      screen.getByText('Plan races and track your iRacing stats.'),
    ).toBeInTheDocument();
  });

  it('renders the tagline in Russian after switching language', async () => {
    setLanguage('ru');
    renderWithProviders(<HomePage />);
    expect(
      await screen.findByText('Планируйте гонки и следите за статистикой iRacing.'),
    ).toBeInTheDocument();
  });
});
