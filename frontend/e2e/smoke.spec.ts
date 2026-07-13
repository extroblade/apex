import { test, expect } from '@playwright/test';

test('home page loads with brand and nav', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Apex' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Fuel' })).toBeVisible();
});

test('fuel calculator computes a plan', async ({ page }) => {
  await page.goto('/fuel');
  await expect(page.getByRole('heading', { name: 'Fuel Planner' })).toBeVisible();
  await page.getByRole('button', { name: 'Calculate' }).click();
  // The plan card renders stint rows once the backend responds.
  await expect(page.getByText('Total laps')).toBeVisible();
});

test('fuel calculator honors mandatory stops and pit windows', async ({ page }) => {
  await page.goto('/fuel');

  // Two mandatory stops on a race that needs only one for fuel.
  await page.getByLabel('Mandatory pit stops').fill('2');
  // Constrain the first stop to laps 5..10.
  await page.getByLabel('Stop 1 window — Earliest').fill('5');
  await page.getByLabel('Stop 1 window — Latest').fill('10');
  await page.getByRole('button', { name: 'Calculate' }).click();

  // 3 stints, each showing its pit lap except the last.
  await expect(page.getByText('pit lap', { exact: false })).toHaveCount(2);

  // The strategy info dialog opens and explains the strategies.
  await page.getByRole('button', { name: 'When to use each strategy' }).click();
  await expect(page.getByText('Undercut — pit before your rival', { exact: false })).toBeVisible();
});
