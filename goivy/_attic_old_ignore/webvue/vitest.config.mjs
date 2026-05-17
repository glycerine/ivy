import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { defineConfig } from 'vitest/config';
import vue from '@vitejs/plugin-vue';

const webvueDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [vue()],
  cacheDir: path.resolve(webvueDir, 'node_modules/.vite-goivy-webvue-test'),
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.js'],
    clearMocks: true,
    restoreMocks: true,
    mockReset: true,
  },
});
