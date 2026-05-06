import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, installSaveDom } from './helpers/fakes.mjs';

function loadApp() {
  return loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
}

beforeEach(() => {
  installSaveDom();
  window.__ivyVueBridge = undefined;
});

afterEach(() => {
  window.__ivyVueBridge = undefined;
});

describe('IvyApp API wiring', () => {
  it('uses the Vue engine bridge when Vue owns boot', () => {
    const api = { sessionId: 'from-engine' };
    window.__ivyVueBridge = {
      createLegacyApi: vi.fn(() => api),
    };
    const IvyApp = loadApp();

    const app = new IvyApp();

    expect(window.__ivyVueBridge.createLegacyApi).toHaveBeenCalled();
    expect(app.api).toBe(api);
  });

  it('falls back to IvyAPI when the bridge is absent', () => {
    const IvyApp = loadApp();

    const app = new IvyApp();

    expect(app.api).toBeInstanceOf(FakeAPI);
  });

  it('routes keymap state through Vue without mutating Vue-owned radios', () => {
    document.body.insertAdjacentHTML('beforeend', [
      '<label><input type="radio" name="keymap" value="sublime" checked></label>',
      '<label><input type="radio" name="keymap" value="vim"></label>',
    ].join(''));
    window.__ivyVueBridge = {
      setEditorKeymap: vi.fn(),
    };
    const IvyApp = loadApp();
    const app = new IvyApp();
    app.cmEditor = { setOption: vi.fn() };

    app.setEditorKeymap('vim');

    expect(app.cmEditor.setOption).toHaveBeenCalledWith('keyMap', 'vim');
    expect(window.__ivyVueBridge.setEditorKeymap).toHaveBeenCalledWith('vim');
    expect(document.querySelector('input[name="keymap"][value="sublime"]').checked).toBe(true);
    expect(document.querySelector('input[name="keymap"][value="vim"]').checked).toBe(false);
  });

  it('keeps keymap radio mutation as a non-Vue fallback', () => {
    document.body.insertAdjacentHTML('beforeend', [
      '<label><input type="radio" name="keymap" value="sublime" checked></label>',
      '<label><input type="radio" name="keymap" value="vim"></label>',
    ].join(''));
    const IvyApp = loadApp();
    const app = new IvyApp();
    app.cmEditor = { setOption: vi.fn() };

    app.setEditorKeymap('vim');

    expect(document.querySelector('input[name="keymap"][value="sublime"]').checked).toBe(false);
    expect(document.querySelector('input[name="keymap"][value="vim"]').checked).toBe(true);
  });
});
