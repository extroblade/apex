import { test, expect, devices } from '@playwright/test';
import { registerNewUser } from './helpers';

// Pixel = Chromium-based device profile (WebKit isn't installed in this setup).
test.use({ ...devices['Pixel 7'] });

test('mobile: bottom bar fits items without More; profile stays in the header', async ({
  page,
}) => {
  await registerNewUser(page, 'mob');

  const bar = page.getByRole('navigation', { name: 'Primary' });
  await expect(bar).toBeVisible();

  // With iRacing off there are 4 items — they all fit, so no "More" button.
  await expect(bar.getByRole('button', { name: 'More' })).toHaveCount(0);
  await expect(bar.getByRole('link', { name: 'Garage', exact: true })).toBeVisible();

  // Primary items navigate.
  await bar.getByRole('link', { name: 'Planner', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Race Planner' })).toBeVisible();

  // Profile lives in the HEADER on mobile (not in the bottom bar).
  await expect(bar.getByRole('button', { name: 'User menu' })).toHaveCount(0);
  await page.getByRole('banner').getByRole('button', { name: 'User menu' }).click();
  await expect(page.getByRole('menuitem', { name: 'Profile' })).toBeVisible();
});
