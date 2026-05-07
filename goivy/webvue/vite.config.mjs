import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

const webvueDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [vue()],
  root: webvueDir,
  cacheDir: path.resolve(webvueDir, 'node_modules/.vite-goivy-webvue'),
  build: {
    outDir: path.resolve(webvueDir, 'static/dist'),
    emptyOutDir: true,
    sourcemap: true,
    rollupOptions: {
      input: path.resolve(webvueDir, 'src/main.js'),
      output: {
        entryFileNames: 'ivywebvue.js',
        chunkFileNames: 'ivywebvue-[name].js',
        assetFileNames: 'ivywebvue.[ext]',
      },
    },
  },
});
