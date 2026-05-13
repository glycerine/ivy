import { h, nextTick, render } from 'vue';
import AnalysisSheetShell from './components/panes/AnalysisSheetShell.vue';
import { initializeCodeMirrorEditor } from './codeMirrorEditor.js';
import {
  useContextMenuStore,
  useDetailsStore,
  useDialogStore,
  useDropdownStore,
  useEngineStore,
  useLayoutStore,
  useMenuDescriptorStore,
  useRecentFilesStore,
  useSessionStore,
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

  const engineStore = useEngineStore(pinia);
  const contextMenuStore = useContextMenuStore(pinia);
  const detailsStore = useDetailsStore(pinia);
  const dialogStore = useDialogStore(pinia);
  const dropdownStore = useDropdownStore(pinia);
  const sessionStore = useSessionStore(pinia);
  const menuDescriptorStore = useMenuDescriptorStore(pinia);
  const recentFilesStore = useRecentFilesStore(pinia);
  const layoutStore = useLayoutStore(pinia);
  const toastStore = useToastStore(pinia);

  return {
    createIvyApi() {
      return engineStore.engine;
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
    initializeEditor(runtime) {
      return initializeCodeMirrorEditor({
        runtime,
        keymap: runtime && typeof runtime.getEditorKeymap === 'function'
          ? runtime.getEditorKeymap()
          : 'sublime',
        doc,
        codeMirror: win.CodeMirror,
      });
    },
    staticCommandHandlersHandled() {
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
    showDialog(config) {
      return dialogStore.open(config);
    },
    updateRecentFiles(items, loader) {
      recentFilesStore.setItems(items, loader);
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
  return engineStore.engine;
}

export function installIvyVueBridge(options = {}) {
  const win = options.win || globalThis.window;
  configureEngineFromRuntime({ pinia: options.pinia, win });
  const bridge = createIvyVueBridge(options);
  win.__ivyVueBridge = bridge;
  return bridge;
}
