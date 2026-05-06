import { createApp, h, nextTick, render } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import AnalysisSheetShell from './components/panes/AnalysisSheetShell.vue';
import { useContextMenuStore } from './stores/contextMenuStore.js';
import { useDetailsStore } from './stores/detailsStore.js';
import { useDialogStore } from './stores/dialogStore.js';
import { useDropdownStore } from './stores/dropdownStore.js';
import { useEditorStore } from './stores/editorStore.js';
import { useEventTraceStore } from './stores/eventTraceStore.js';
import { useLayoutStore } from './stores/layoutStore.js';
import { useMenuDescriptorStore } from './stores/menuDescriptorStore.js';
import { useGraphStore } from './stores/graphStore.js';
import { useRecentFilesStore } from './stores/recentFilesStore.js';
import { useSessionStore } from './stores/sessionStore.js';
import { useSheetStore } from './stores/sheetStore.js';
import { useStateRelationsStore } from './stores/stateRelationsStore.js';
import { useToastStore } from './stores/toastStore.js';

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);

const editorStore = useEditorStore(pinia);
const contextMenuStore = useContextMenuStore(pinia);
const detailsStore = useDetailsStore(pinia);
const dialogStore = useDialogStore(pinia);
const dropdownStore = useDropdownStore(pinia);
const sessionStore = useSessionStore(pinia);
const stateRelationsStore = useStateRelationsStore(pinia);
const menuDescriptorStore = useMenuDescriptorStore(pinia);
const sheetStore = useSheetStore(pinia);
const graphStore = useGraphStore(pinia);
const recentFilesStore = useRecentFilesStore(pinia);
const eventTraceStore = useEventTraceStore(pinia);
const layoutStore = useLayoutStore(pinia);
const toastStore = useToastStore(pinia);

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
  setEditorKeymap(keymap) {
    editorStore.setKeymap(keymap);
  },
  getEditorKeymap() {
    return editorStore.keymap;
  },
  editorKeymapHandled() {
    return true;
  },
  updateReopenLastFileButton(visible, label) {
    editorStore.setReopenLastFileButton(visible, label);
  },
  staticCommandHandlersHandled() {
    return true;
  },
  tabClicksHandled() {
    return true;
  },
  fileInputHandlersHandled() {
    return true;
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
  getStateRelationToggles() {
    return stateRelationsStore.toggleSnapshot;
  },
  setStateRelationToggles(toggles) {
    stateRelationsStore.applyToggleSnapshot(toggles || {});
  },
  buildStateRelationVisibility() {
    return stateRelationsStore.visibilitySnapshot;
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
  setSessionId(sessionId) {
    sessionStore.setSessionId(sessionId);
  },
  setMode(mode) {
    sessionStore.setMode(mode);
  },
  getMode() {
    return sessionStore.mode;
  },
  setLoadedFile(fileName, filePath) {
    sessionStore.setLoadedFile(fileName, filePath);
  },
  showLoading(message) {
    sessionStore.showLoading(message);
  },
  hideLoading() {
    sessionStore.hideLoading();
  },
  showToast(message, level, options) {
    return toastStore.show(message, level, options || {});
  },
  setSaveAsNoticeVisible(visible) {
    sessionStore.setSaveAsNoticeVisible(visible);
  },
  updateMenuRegion(region, menus, dispatcher) {
    menuDescriptorStore.setRegion(region, menus, dispatcher);
  },
  closeDropdownMenus() {
    menuDescriptorStore.closeAll();
    dropdownStore.closeAll();
  },
  upsertSheetTab(tab) {
    sheetStore.upsertTab(tab);
  },
  createAnalysisSheetShell({ id, counter }) {
    if (!id) return null;
    const sheetArea = document.getElementById('sheet-area');
    if (!sheetArea) return null;
    let sheet = document.getElementById(id);
    if (!sheet) {
      sheet = document.createElement('div');
      sheet.id = id;
      sheet.className = 'sheet-content';
      sheet.__ivyVueRenderedSheet = true;
      sheetArea.appendChild(sheet);
    }
    const vnode = h(AnalysisSheetShell, { counter });
    vnode.appContext = app._context;
    render(vnode, sheet);
    return sheet;
  },
  removeRenderedSheet(sheetId) {
    const sheet = document.getElementById(sheetId);
    if (sheet && sheet.__ivyVueRenderedSheet) {
      render(null, sheet);
      return true;
    }
    return false;
  },
  activateSheetTab(sheetId) {
    sheetStore.activateTab(sheetId);
  },
  removeSheetTab(sheetId) {
    sheetStore.removeTab(sheetId);
    eventTraceStore.removeSheet(sheetId);
  },
  resetSheetTabs() {
    sheetStore.resetTabs();
    eventTraceStore.reset();
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
  editorSaveSheenHandled() {
    return true;
  },
  updateRecentFiles(items, loader) {
    recentFilesStore.setItems(items, loader);
  },
  upsertEventTraceSheet(sheet) {
    eventTraceStore.upsertSheet(sheet);
  },
  setEventTraceExpanded(sheetId, address, expanded) {
    eventTraceStore.setExpanded(sheetId, address, expanded);
  },
  isEventTraceExpanded(sheetId, address) {
    return eventTraceStore.isExpanded(sheetId, address);
  },
  selectEventTraceRow(sheetId, address) {
    eventTraceStore.selectEvent(sheetId, address);
  },
  updateEventPatterns(sheetId, patterns) {
    eventTraceStore.setPatterns(sheetId, patterns);
  },
  setSelectedEventPatternIndex(sheetId, index) {
    eventTraceStore.setSelectedPatternIndex(sheetId, index);
  },
  getSelectedEventPattern(sheetId) {
    return eventTraceStore.selectedPattern(sheetId);
  },
  getSelectedEventPatternIndex(sheetId) {
    return eventTraceStore.selectedPatternIndex(sheetId);
  },
  setTutorialVisible(visible) {
    layoutStore.setTutorialVisible(visible);
  },
  isTutorialVisible() {
    return layoutStore.tutorialVisible;
  },
  setArgPanelWidth(width) {
    layoutStore.setArgPanelWidth(width);
  },
  setStatePanelWidth(width) {
    layoutStore.setStatePanelWidth(width);
  },
  setDetailsHeight(height) {
    layoutStore.setDetailsHeight(height);
  },
  setEditorWidth(width) {
    layoutStore.setEditorWidth(width);
  },
  setTutorialHeight(height) {
    layoutStore.setTutorialHeight(height);
  },
  tutorialUrlBarHandled() {
    return true;
  },
  afterLayoutSettled(callback) {
    if (typeof callback !== 'function') return;
    nextTick(() => {
      const raf = typeof window.requestAnimationFrame === 'function'
        ? window.requestAnimationFrame.bind(window)
        : (fn) => window.setTimeout(fn, 0);
      raf(() => raf(callback));
    });
  },
};

app.mount('#ivy-vue-root');
