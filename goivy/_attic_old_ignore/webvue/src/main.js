import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { authClientKey, passkeyBrowserKey } from './app/dependencies.js';
import { createAuthClient } from './api/authClient.js';
import { createPasskeyBrowser } from './api/passkeyBrowser.js';

export function mountIvyWebvue(target = '#app') {
  const app = createApp(App);
  app.use(createPinia());
  app.provide(authClientKey, createAuthClient());
  app.provide(passkeyBrowserKey, createPasskeyBrowser());
  app.mount(target);
  return app;
}

if (typeof document !== 'undefined') {
  mountIvyWebvue();
}
