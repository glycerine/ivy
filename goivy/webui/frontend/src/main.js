import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { installIvyVueBridge } from './ivyVueBridge.js';

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);

installIvyVueBridge({ app, pinia });

app.mount('#ivy-vue-root');
