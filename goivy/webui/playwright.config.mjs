import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './pw_test',
  fullyParallel: true,
  timeout: 30_000,
  workers: 4,
  reporter: [['list']],
  webServer: {
    command: 'go run ../cmd/ivyweb -addr 127.0.0.1:18089',
    env: {
      GOCACHE: '/private/tmp/goivy-webui-playwright-gocache',
    },
    url: 'http://127.0.0.1:18089',
    reuseExistingServer: false,
    timeout: 120_000,
    stdout: 'ignore',
    stderr: 'pipe',
  },
  use: {
    baseURL: 'http://127.0.0.1:18089',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'off',
    viewport: { width: 1280, height: 800 },
  },
  projects: [
    {
      name: 'chrome',
      use: {
        ...devices['Desktop Chrome'],
        channel: 'chrome',
      },
    },
  ],
});
