import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

const frontendDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [vue()],
  root: frontendDir,
  cacheDir: path.resolve(frontendDir, '../../../node_modules/.vite-goivy-webui-vue'),
  build: {
    outDir: path.resolve(frontendDir, '../static/dist'),
    emptyOutDir: true,
    sourcemap: true,
    rollupOptions: {
      input: path.resolve(frontendDir, 'src/main.js'),
      output: {
        entryFileNames: 'ivyweb.js',
        chunkFileNames: 'ivyweb-[name].js',
        assetFileNames: 'ivyweb.[ext]',
      },
    },
  },
});
