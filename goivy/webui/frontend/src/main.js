import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { installIvyVueBridge } from './ivyVueBridge.js';
import { loadLegacyIvyRuntime } from './legacyScripts.js';

async function bootstrap() {
  await loadLegacyIvyRuntime();

  const app = createApp(App);
  const pinia = createPinia();
  app.use(pinia);

  installIvyVueBridge({ app, pinia });

  app.mount('#ivy-vue-root');
}

bootstrap().catch((err) => {
  window.__ivyInitError = err && err.message ? err.message : String(err);
  console.error('Failed to boot Ivy Vue app:', err);
});
