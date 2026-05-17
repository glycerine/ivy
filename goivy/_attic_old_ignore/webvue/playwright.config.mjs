import { defineConfig, devices } from '@playwright/test';

const channel = process.env.PLAYWRIGHT_CHANNEL || 'chrome';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  fullyParallel: true,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    ...devices['Desktop Chrome'],
    channel: channel === 'bundled' ? undefined : channel,
    headless: true,
  },
});
