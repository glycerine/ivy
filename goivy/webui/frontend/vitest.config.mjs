import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { defineConfig } from 'vitest/config';

const frontendDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  cacheDir: path.resolve(frontendDir, '../node_modules/.vite-goivy-webui-test'),
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.{js,ts}'],
    clearMocks: true,
    restoreMocks: true,
    mockReset: true,
  },
});
