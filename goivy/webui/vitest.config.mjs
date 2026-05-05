import { defineConfig } from 'vitest/config';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');

export default defineConfig({
  cacheDir: path.join(repoRoot, 'node_modules/.vite-goivy-webui'),
  test: {
    environment: 'jsdom',
    include: ['js_test/**/*.test.mjs'],
    clearMocks: true,
    restoreMocks: true,
    mockReset: true,
  },
});
