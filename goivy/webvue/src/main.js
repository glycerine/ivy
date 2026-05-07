import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';

export function mountIvyWebvue(target = '#app') {
  const app = createApp(App);
  app.use(createPinia());
  app.mount(target);
  return app;
}

if (typeof document !== 'undefined') {
  mountIvyWebvue();
}
