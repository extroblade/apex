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
