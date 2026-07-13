import { test, expect } from '@playwright/test';
import { registerNewUser } from './helpers';

test('season planner: favorite a series in the garage, see its season row', async ({
  page,
}) => {
  await registerNewUser(page, 'plan');

  // Garage: favorite the first series (catalog is seeded, no iRacing needed).
  await page.getByRole('link', { name: 'Garage', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Garage' })).toBeVisible();
  await page.getByRole('button', { name: 'Series', exact: true }).click();
  const firstSeries = page.locator('input[type="checkbox"]').first();
  await firstSeries.click();
  await expect(firstSeries).toBeChecked();

  // Planner: the season view shows the favorites grid.
  await page.getByRole('link', { name: 'Planner', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Race Planner' })).toBeVisible();
  await expect(page.getByText('/13 weeks', { exact: false }).first()).toBeVisible();

  // The grid has 13 week columns.
  await expect(page.getByRole('columnheader', { name: 'W13' })).toBeVisible();

  // The current-week races live on their own page, reachable from the toolbar.
  await page.getByRole('link', { name: 'This week' }).click();
  await expect(page.getByRole('heading', { name: 'This Week' })).toBeVisible();
  await page.getByRole('link', { name: 'Full season' }).click();
  await expect(page.getByRole('heading', { name: 'Race Planner' })).toBeVisible();

  // Click a schedule cell to plan that race; it flips to pressed state.
  const firstCell = page.locator('tbody button[aria-pressed]').first();
  await firstCell.click();
  await expect(firstCell).toHaveAttribute('aria-pressed', 'true');
  await expect(page.getByText('1 planned')).toBeVisible();

  // Transpose swaps the axes: weeks become rows.
  await page.getByRole('button', { name: 'Swap rows/columns' }).click();
  await expect(page.getByRole('rowheader', { name: 'W13' })).toBeVisible();

  // Switching to all series shows every seeded series (more columns than favorites).
  await page.getByRole('button', { name: 'All series' }).click();
  expect(await page.locator('thead th').count()).toBeGreaterThan(5);
});
