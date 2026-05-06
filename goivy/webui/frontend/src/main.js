import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { useContextMenuStore } from './stores/contextMenuStore.js';
import { useDetailsStore } from './stores/detailsStore.js';
import { useDialogStore } from './stores/dialogStore.js';
import { useEditorStore } from './stores/editorStore.js';
import { useMenuDescriptorStore } from './stores/menuDescriptorStore.js';
import { useGraphStore } from './stores/graphStore.js';
import { useRecentFilesStore } from './stores/recentFilesStore.js';
import { useSessionStore } from './stores/sessionStore.js';
import { useSheetStore } from './stores/sheetStore.js';
import { useStateRelationsStore } from './stores/stateRelationsStore.js';

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);

const editorStore = useEditorStore(pinia);
const contextMenuStore = useContextMenuStore(pinia);
const detailsStore = useDetailsStore(pinia);
const dialogStore = useDialogStore(pinia);
const sessionStore = useSessionStore(pinia);
const stateRelationsStore = useStateRelationsStore(pinia);
const menuDescriptorStore = useMenuDescriptorStore(pinia);
const sheetStore = useSheetStore(pinia);
const graphStore = useGraphStore(pinia);
const recentFilesStore = useRecentFilesStore(pinia);

function syncContextMenuElement(visible, x = 0, y = 0) {
  const menu = document.getElementById('context-menu');
  if (!menu) return;
  menu.style.display = visible ? 'block' : 'none';
  if (visible) {
    menu.style.left = `${Number(x) || 0}px`;
    menu.style.top = `${Number(y) || 0}px`;
  }
}

window.__ivyVueBridge = {
  showContextMenu(x, y, actions) {
    contextMenuStore.show(x, y, actions);
    syncContextMenuElement(true, x, y);
  },
  hideContextMenu() {
    contextMenuStore.hide();
    syncContextMenuElement(false);
  },
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
  showLoading(message) {
    sessionStore.showLoading(message);
  },
  hideLoading() {
    sessionStore.hideLoading();
  },
  updateMenuRegion(region, menus, dispatcher) {
    menuDescriptorStore.setRegion(region, menus, dispatcher);
  },
  closeDropdownMenus() {
    menuDescriptorStore.closeAll();
  },
  upsertSheetTab(tab) {
    sheetStore.upsertTab(tab);
  },
  activateSheetTab(sheetId) {
    sheetStore.activateTab(sheetId);
  },
  removeSheetTab(sheetId) {
    sheetStore.removeTab(sheetId);
  },
  resetSheetTabs() {
    sheetStore.resetTabs();
  },
  getSheetTabLabel(sheetId) {
    return sheetStore.labelFor(sheetId);
  },
  setActiveGraphSheet(sheetId) {
    graphStore.setActiveSheet(sheetId);
  },
  updateGraphSnapshot(sheetId, kind, snapshot) {
    graphStore.applyGraphSnapshot(sheetId, kind, snapshot);
  },
  selectArgNode(sheetId, nodeId) {
    graphStore.selectArgNode(nodeId, sheetId);
  },
  showDialog(config) {
    return dialogStore.open(config);
  },
  updateRecentFiles(items, loader) {
    recentFilesStore.setItems(items, loader);
  },
};

app.mount('#ivy-vue-root');
