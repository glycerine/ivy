import { beforeEach, describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph } from './helpers/fakes.mjs';

function setOffsetWidth(el, width) {
  Object.defineProperty(el, 'offsetWidth', {
    configurable: true,
    value: width,
  });
}

function mouse(type, clientX) {
  return new MouseEvent(type, {
    clientX,
    bubbles: true,
    cancelable: true,
  });
}

function loadApp() {
  return loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
}

beforeEach(() => {
  document.body.innerHTML = [
    '<div id="top-row">',
    '  <div id="sheet-area"></div>',
    '  <div id="divider3"></div>',
    '  <div id="editor-panel"></div>',
    '</div>',
  ].join('');
});

describe('IvyApp layout resizers', () => {
  it('widens the editor pane when divider3 is dragged left', () => {
    const IvyApp = loadApp();
    const app = new IvyApp();
    app.argGraph = { resize: vi.fn() };
    app.conceptGraph = { resize: vi.fn() };
    app._refreshEditorLayout = vi.fn();

    const topRow = document.getElementById('top-row');
    const sheetArea = document.getElementById('sheet-area');
    const divider3 = document.getElementById('divider3');
    const editorPanel = document.getElementById('editor-panel');
    setOffsetWidth(topRow, 1200);
    setOffsetWidth(divider3, 4);
    setOffsetWidth(sheetArea, 756);
    setOffsetWidth(editorPanel, 440);

    app.setupResizer3();
    divider3.dispatchEvent(mouse('mousedown', 760));
    document.dispatchEvent(mouse('mousemove', 620));

    expect(sheetArea.style.flex).toBe('1 1 auto');
    expect(editorPanel.style.flex).toBe('0 0 580px');
    expect(app.argGraph.resize).toHaveBeenCalled();
    expect(app.conceptGraph.resize).toHaveBeenCalled();
    expect(app._refreshEditorLayout).toHaveBeenCalled();

    document.dispatchEvent(mouse('mouseup', 620));

    expect(document.body.style.cursor).toBe('');
    expect(document.body.style.userSelect).toBe('');
  });
});
