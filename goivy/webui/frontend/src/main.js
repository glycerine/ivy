import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { installIvyVueBridge } from './ivyVueBridge.js';
import { createAppServices, installAppServices } from './services/appServices.js';

async function bootstrap() {
  const app = createApp(App);
  const pinia = createPinia();
  app.use(pinia);

  installIvyVueBridge({ app, pinia });
  installAppServices(createAppServices());

  app.mount('#ivy-vue-root');
}

bootstrap().catch((err) => {
  window.__ivyInitError = err && err.message ? err.message : String(err);
  console.error('Failed to boot Ivy Vue app:', err);
});
