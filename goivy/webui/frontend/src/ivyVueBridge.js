import { h, nextTick, render } from 'vue';
import AnalysisSheetShell from './components/panes/AnalysisSheetShell.vue';
import { initializeCodeMirrorEditor } from './codeMirrorEditor.js';
import { IvyApiAdapter } from './engines/index.js';
import {
  useContextMenuStore,
  useDetailsStore,
  useDialogStore,
  useDropdownStore,
  useEditorStore,
  useEngineStore,
  useEventTraceStore,
  useGraphStore,
  useLayoutStore,
  useMenuDescriptorStore,
  useRecentFilesStore,
  useSessionStore,
  useSheetStore,
  useStateRelationsStore,
  useToastStore,
} from './stores/index.js';

export function syncContextMenuElement(visible, x = 0, y = 0, doc = globalThis.document) {
  const menu = doc && doc.getElementById('context-menu');
  if (!menu) return;
  menu.style.display = visible ? 'block' : 'none';
  if (visible) {
    menu.style.left = `${Number(x) || 0}px`;
    menu.style.top = `${Number(y) || 0}px`;
  }
}

export function createIvyVueBridge({
  app,
  pinia,
  doc = globalThis.document,
  win = globalThis.window,
  hFn = h,
  nextTickFn = nextTick,
  renderFn = render,
} = {}) {
  if (!pinia) {
    throw new Error('createIvyVueBridge requires a Pinia instance');
  }

  const editorStore = useEditorStore(pinia);
  const engineStore = useEngineStore(pinia);
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

  return {
    createIvyApi() {
      return new IvyApiAdapter(engineStore.engine);
    },
    getEngine() {
      return engineStore.engine;
    },
    showContextMenu(x, y, actions) {
      contextMenuStore.show(x, y, actions);
      syncContextMenuElement(true, x, y, doc);
    },
    hideContextMenu() {
      contextMenuStore.hide();
      syncContextMenuElement(false, 0, 0, doc);
    },
    updateEditor(snapshot) {
      editorStore.applyRuntimeSnapshot(snapshot);
    },
    initializeEditor(runtime) {
      return initializeCodeMirrorEditor({
        runtime,
        editorStore,
        doc,
        codeMirror: win.CodeMirror,
      });
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
    layoutResizersHandled() {
      return true;
    },
    globalInteractionsHandled() {
      return true;
    },
    graphContextMenuSuppressionHandled() {
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
    flashMenuItem(id, durationMs) {
      dropdownStore.flashItem(id, durationMs);
    },
    upsertSheetTab(tab) {
      sheetStore.upsertTab(tab);
    },
    createAnalysisSheetShell({ id, counter }) {
      if (!id) return null;
      const sheetArea = doc && doc.getElementById('sheet-area');
      if (!sheetArea) return null;
      let sheet = doc.getElementById(id);
      if (!sheet) {
        sheet = doc.createElement('div');
        sheet.id = id;
        sheet.className = 'sheet-content';
        sheet.__ivyVueRenderedSheet = true;
        sheetArea.appendChild(sheet);
      }
      const vnode = hFn(AnalysisSheetShell, { counter });
      vnode.appContext = app && app._context;
      renderFn(vnode, sheet);
      return sheet;
    },
    removeRenderedSheet(sheetId) {
      const sheet = doc && doc.getElementById(sheetId);
      if (sheet && sheet.__ivyVueRenderedSheet) {
        renderFn(null, sheet);
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
    flashTutorialButton(durationMs) {
      layoutStore.flashTutorialButton(durationMs);
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
      nextTickFn(() => {
        const raf = typeof win.requestAnimationFrame === 'function'
          ? win.requestAnimationFrame.bind(win)
          : (fn) => win.setTimeout(fn, 0);
        raf(() => raf(callback));
      });
    },
  };
}

export function configureEngineFromRuntime({ pinia, win = globalThis.window } = {}) {
  if (!pinia || !win) return null;
  const engineStore = useEngineStore(pinia);
  if (win.__IVY_ENGINE__) {
    const kind = win.__IVY_ENGINE_KIND__ || win.__IVY_ENGINE__.kind || 'custom';
    return engineStore.setEngine(kind, win.__IVY_ENGINE__);
  }
  if (win.__IVY_ENGINE_KIND__ === 'wanix') {
    return engineStore.useWanix();
  }
  return engineStore.engine;
}

export function installIvyVueBridge(options = {}) {
  const win = options.win || globalThis.window;
  configureEngineFromRuntime({ pinia: options.pinia, win });
  const bridge = createIvyVueBridge(options);
  win.__ivyVueBridge = bridge;
  return bridge;
}
