import { test, expect } from '@playwright/test';
import { registerNewUser } from './helpers';

test('user can update their nickname from the profile page', async ({ page }) => {
  await registerNewUser(page, 'prof');

  // Open the avatar user menu and go to Profile.
  await page.getByRole('button', { name: 'User menu' }).click();
  await page.getByRole('menuitem', { name: 'Profile' }).click();

  await expect(page.getByRole('heading', { name: 'Profile' })).toBeVisible();
  await page.getByLabel('Nickname').fill('Speedy');
  await page.getByRole('button', { name: 'Save changes' }).click();

  await expect(page.getByText('Saved.')).toBeVisible();
});

test('user can download their data export from the profile page', async ({ page }) => {
  await registerNewUser(page, 'exp');

  await page.getByRole('button', { name: 'User menu' }).click();
  await page.getByRole('menuitem', { name: 'Profile' }).click();

  await expect(page.getByRole('heading', { name: 'Profile' })).toBeVisible();

  // The "Your data" card should be visible, with a download button.
  await expect(page.getByRole('heading', { name: 'Your data' })).toBeVisible();

  // Set up a download listener before clicking, since the download starts
  // immediately on click.
  const downloadPromise = page.waitForEvent('download');
  await page.getByRole('button', { name: 'Download my data' }).click();
  const download = await downloadPromise;

  expect(download.suggestedFilename()).toMatch(/apex-account-export-.*\.json/);
});
