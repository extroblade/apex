import { test, expect } from '@playwright/test';
import { registerNewUser } from './helpers';

test('iRacing nav is hidden while the feature is off; planner/garage remain', async ({
  page,
}) => {
  await registerNewUser(page, 'feat');

  // Non-iRacing sections are available.
  await expect(page.getByRole('link', { name: 'Planner', exact: true })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Garage', exact: true })).toBeVisible();

  // iRacing-only sections are hidden behind the feature flag.
  await expect(page.getByRole('link', { name: 'Drivers', exact: true })).toHaveCount(0);
  await expect(page.getByRole('link', { name: 'Dashboard', exact: true })).toHaveCount(0);
  await expect(page.getByRole('link', { name: 'Compare', exact: true })).toHaveCount(0);
});
