import { test, expect } from '@playwright/test';

test('a user can register and then log out', async ({ page }) => {
  const email = `e2e_${Date.now()}@example.com`;

  await page.goto('/login');
  // The form starts in "log in" mode; switch to sign up.
  await page.getByRole('button', { name: 'Sign up' }).click();
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill('supersecret');
  await page.getByRole('button', { name: 'Sign up' }).click();

  // Landed on home, now signed in.
  await expect(page.getByText(`Signed in as ${email}.`)).toBeVisible();

  // Open the user menu (avatar) and log out.
  await page.getByRole('button', { name: 'User menu' }).click();
  await page.getByRole('menuitem', { name: 'Log out' }).click();

  await expect(page.getByText('Plan races and track your iRacing stats.')).toBeVisible();
});
