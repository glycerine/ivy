import { afterEach, describe, expect, it, vi } from 'vitest';
import { createApp } from 'vue';
import { createPinia, setActivePinia } from 'pinia';
import { configureEngineFromRuntime, createIvyVueBridge, installIvyVueBridge, syncContextMenuElement } from './ivyVueBridge.js';
import { useContextMenuStore, useEditorStore, useEngineStore, useSheetStore } from './stores/index.js';

function makeBridge() {
  const app = createApp({ template: '<div />' });
  const pinia = createPinia();
  app.use(pinia);
  setActivePinia(pinia);
  return {
    app,
    pinia,
    bridge: createIvyVueBridge({ app, pinia }),
  };
}

afterEach(() => {
  document.body.innerHTML = '';
  delete window.__ivyVueBridge;
  delete window.__IVY_ENGINE__;
  delete window.__IVY_ENGINE_KIND__;
});

describe('ivyVueBridge', () => {
  it('keeps the context menu store and legacy menu element in sync', () => {
    document.body.innerHTML = '<div id="context-menu" style="display:none"></div>';
    const { bridge, pinia } = makeBridge();
    const contextMenuStore = useContextMenuStore(pinia);
    const callback = () => {};
    const action = { name: 'Inspect', callback };

    bridge.showContextMenu(12, 34, [action]);

    expect(contextMenuStore.visible).toBe(true);
    expect(contextMenuStore.items).toEqual([
      {
        kind: 'item',
        key: 'item-0-Inspect',
        name: 'Inspect',
        id: 'Inspect',
        callback,
      },
    ]);
    expect(document.getElementById('context-menu').style.display).toBe('block');
    expect(document.getElementById('context-menu').style.left).toBe('12px');
    expect(document.getElementById('context-menu').style.top).toBe('34px');

    bridge.hideContextMenu();

    expect(contextMenuStore.visible).toBe(false);
    expect(document.getElementById('context-menu').style.display).toBe('none');
  });

  it('renders and unmounts dynamic analysis sheet shells for the legacy graph controller', () => {
    document.body.innerHTML = '<div id="sheet-area"></div>';
    const { bridge } = makeBridge();

    const sheet = bridge.createAnalysisSheetShell({ id: 'sheet-2', counter: 2 });

    expect(sheet).toBe(document.getElementById('sheet-2'));
    expect(sheet.__ivyVueRenderedSheet).toBe(true);
    expect(sheet.querySelector('#arg-graph-2')).not.toBeNull();
    expect(sheet.querySelector('#concept-graph-2')).not.toBeNull();
    expect(bridge.removeRenderedSheet('sheet-2')).toBe(true);
    expect(sheet.innerHTML).toBe('');
  });

  it('exposes editor and sheet state through the installed global bridge', () => {
    const { app, pinia } = makeBridge();
    const bridge = installIvyVueBridge({ app, pinia });
    const editorStore = useEditorStore(pinia);
    const sheetStore = useSheetStore(pinia);

    bridge.updateEditor({
      path: 'client_server_example.ivy',
      content: 'changed',
      savedContent: 'original',
    });
    bridge.upsertSheetTab({ id: 'events-1', label: 'Trace', type: 'events', closable: true });
    bridge.activateSheetTab('events-1');

    expect(window.__ivyVueBridge).toBe(bridge);
    expect(editorStore.path).toBe('client_server_example.ivy');
    expect(editorStore.dirty).toBe(true);
    expect(sheetStore.activeSheetId).toBe('events-1');
    expect(bridge.getSheetTabLabel('events-1')).toBe('Trace');
  });

  it('can select a runtime-supplied engine before the legacy adapter is created', () => {
    const { pinia } = makeBridge();
    const fakeEngine = { kind: 'in-browser', createSession: vi.fn() };
    window.__IVY_ENGINE__ = fakeEngine;
    window.__IVY_ENGINE_KIND__ = 'test-engine';

    const selected = configureEngineFromRuntime({ pinia, win: window });
    const engineStore = useEngineStore(pinia);

    expect(selected).toBe(fakeEngine);
    expect(engineStore.kind).toBe('test-engine');
    expect(engineStore.engine).toBe(fakeEngine);

    delete window.__IVY_ENGINE__;
    delete window.__IVY_ENGINE_KIND__;
  });

  it('can switch to the Wanix placeholder engine through runtime configuration', () => {
    const { app, pinia } = makeBridge();
    window.__IVY_ENGINE_KIND__ = 'wanix';

    const bridge = installIvyVueBridge({ app, pinia });

    expect(useEngineStore(pinia).kind).toBe('wanix');
    expect(bridge.getEngine().kind).toBe('wanix');

    delete window.__IVY_ENGINE_KIND__;
  });

  it('tolerates a missing legacy context menu element', () => {
    expect(() => syncContextMenuElement(true, 1, 2)).not.toThrow();
  });
});
