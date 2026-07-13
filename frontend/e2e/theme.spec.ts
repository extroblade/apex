import { test, expect } from '@playwright/test';

test('theme can be switched from the preferences menu', async ({ page }) => {
  await page.goto('/');

  // Signed out: the settings menu still exposes theme + language.
  await page.getByRole('button', { name: 'Preferences' }).click();
  await page.getByRole('menuitemradio', { name: 'Dark' }).click();

  await expect(page.locator('html')).toHaveClass(/dark/);
});
