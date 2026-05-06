import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { useDetailsStore } from './stores/detailsStore.js';
import { useEditorStore } from './stores/editorStore.js';
import { useSessionStore } from './stores/sessionStore.js';

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);

const editorStore = useEditorStore(pinia);
const detailsStore = useDetailsStore(pinia);
const sessionStore = useSessionStore(pinia);
window.__ivyVueBridge = {
  updateEditor(snapshot) {
    editorStore.applyLegacySnapshot(snapshot);
  },
  updateDetails(details) {
    detailsStore.setDetails(details);
  },
  clearDetails() {
    detailsStore.clear();
  },
  setCheckTraceAction(action) {
    detailsStore.setTraceAction(action);
  },
  updateStatus(message, level = '') {
    sessionStore.setStatus(message, level);
  },
};

app.mount('#ivy-vue-root');
