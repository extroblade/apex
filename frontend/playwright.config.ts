import { defineConfig, devices } from '@playwright/test';

/**
 * E2E runs against the running stack (docker: `./dev.sh up`, app on :3000, which
 * proxies /api to the backend). Override the target with E2E_BASE_URL.
 *
 * First-time setup: `npx playwright install chromium`.
 */
export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: 'list',
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://localhost:3000',
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});
