import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { defineConfig } from 'vite';

const frontendDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  root: frontendDir,
  base: '/static/dist/',
  cacheDir: path.resolve(frontendDir, '../node_modules/.vite-goivy-webui'),
  build: {
    outDir: path.resolve(frontendDir, '../static/dist'),
    emptyOutDir: true,
    sourcemap: true,
    chunkSizeWarningLimit: 2000,
    rollupOptions: {
      input: path.resolve(frontendDir, 'src/main.ts'),
      output: {
        entryFileNames: 'ivyweb.js',
        chunkFileNames: 'ivyweb-[name].js',
        assetFileNames: 'ivyweb.[ext]',
      },
    },
  },
  worker: {
    format: 'es',
  },
});
