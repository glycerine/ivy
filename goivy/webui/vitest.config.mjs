import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    include: ['js_test/**/*.test.mjs'],
    clearMocks: true,
    restoreMocks: true,
    mockReset: true,
  },
});
