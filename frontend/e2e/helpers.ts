import { type Page, expect } from '@playwright/test';

/** Registers a brand-new user (unique email) and asserts the signed-in home. */
export async function registerNewUser(page: Page, prefix = 'e2e'): Promise<string> {
  const email = `${prefix}_${Date.now()}_${Math.floor(Math.random() * 100000)}@example.com`;
  await page.goto('/login');
  await page.getByRole('button', { name: 'Sign up' }).click();
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill('supersecret');
  await page.getByRole('button', { name: 'Sign up' }).click();
  await expect(page.getByText(`Signed in as ${email}.`)).toBeVisible();
  return email;
}
