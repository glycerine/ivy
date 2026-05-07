import { defineConfig, devices } from '@playwright/test';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const controlDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(controlDir, '../..');
const goBuildCache = path.join(repoRoot, '.cache', 'go-build-control-playwright');

export default defineConfig({
  testDir: './pw_test',
  fullyParallel: true,
  timeout: 30_000,
  workers: 1,
  reporter: [['list']],
  webServer: {
    command: 'go run ../cmd/ivy-control -addr 127.0.0.1:18080',
    env: {
      GOCACHE: goBuildCache,
      IVY_CONTROL_OIDC_AUTH_URL: 'http://127.0.0.1:18082/login/oauth/authorize',
      IVY_CONTROL_OIDC_CLIENT_ID: 'ivy-control-local',
      IVY_CONTROL_OIDC_REDIRECT_URL: 'http://127.0.0.1:18080/auth/callback',
    },
    url: 'http://127.0.0.1:18080/healthz',
    reuseExistingServer: false,
    timeout: 120_000,
    stdout: 'ignore',
    stderr: 'pipe',
  },
  use: {
    baseURL: 'http://127.0.0.1:18080',
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
