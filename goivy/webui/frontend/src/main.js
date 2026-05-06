import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { useDetailsStore } from './stores/detailsStore.js';
import { useEditorStore } from './stores/editorStore.js';
import { useSessionStore } from './stores/sessionStore.js';
import { useStateRelationsStore } from './stores/stateRelationsStore.js';

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);

const editorStore = useEditorStore(pinia);
const detailsStore = useDetailsStore(pinia);
const sessionStore = useSessionStore(pinia);
const stateRelationsStore = useStateRelationsStore(pinia);
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
  updateConstraintFacts(facts, callback) {
    detailsStore.setConstraintFacts(facts, callback);
  },
  setCheckTraceAction(action) {
    detailsStore.setTraceAction(action);
  },
  updateStateRelations(rows, onToggle) {
    stateRelationsStore.setRows(rows, onToggle);
  },
  clearStateRelations() {
    stateRelationsStore.clear();
  },
  updateStateLabel(value) {
    stateRelationsStore.setStateLabel(value);
  },
  updateStatus(message, level = '') {
    sessionStore.setStatus(message, level);
  },
};

app.mount('#ivy-vue-root');
