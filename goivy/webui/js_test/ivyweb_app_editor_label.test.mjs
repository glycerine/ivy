import { beforeEach, describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeControls, FakeGraph } from './helpers/fakes.mjs';

class InitAPI {
  constructor() {
    this.sessionId = 's1';
    this.connectEvents = vi.fn();
  }

  async createSession() {
    this.sessionId = 's1';
  }
}

function makePersist() {
  return {
    load: vi.fn(() => null),
    setSessionIdInURL: vi.fn(),
    getSessionIdFromURL: vi.fn(() => 's1'),
    save: vi.fn(),
  };
}

function makeCodeMirror(initialValue) {
  let value = initialValue;
  const handlers = {};
  const editor = {
    getValue: vi.fn(() => value),
    setValue: vi.fn((next) => {
      value = next;
    }),
    setOption: vi.fn(),
    on: vi.fn((event, handler) => {
      handlers[event] = handler;
    }),
    triggerChange(next) {
      value = next;
      handlers.change();
    },
  };
  return {
    editor,
    CodeMirror: {
      fromTextArea: vi.fn(() => editor),
    },
  };
}

beforeEach(() => {
  document.body.innerHTML = [
    '<div id="session-id"></div>',
    '<textarea id="model-editor"></textarea>',
    '<div id="model-editor-label"></div>',
    '<button id="file-reopen-last"></button>',
  ].join('');
});

describe('IvyApp editor dirty label', () => {
  it('adds the dirty marker and removes the saved suffix when CodeMirror changes', async () => {
    const { editor, CodeMirror } = makeCodeMirror('saved source');
    const IvyApp = loadIvyApp({
      IvyAPI: InitAPI,
      IvyControls: FakeControls,
      IvyGraph: FakeGraph,
      IvyPersist: makePersist(),
      CodeMirror,
    });
    const app = new IvyApp();
    app.loadMenuDescriptors = vi.fn(async () => undefined);
    app.setupEventHandlers = vi.fn();
    app.setupTabs = vi.fn();
    app.setupResizer = vi.fn();
    app.setupResizer2 = vi.fn();
    app.setupResizer3 = vi.fn();
    app.setupResizerH = vi.fn();
    app.setupDetailsResizer = vi.fn();
    app.setupTutorialUrlBar = vi.fn();
    app.setupKeyboardShortcuts = vi.fn();

    await app.init();
    app._persistedFileName = 'client.ivy';
    app._persistedFilePath = 'client.ivy';
    app._persistedFileContent = 'saved source';
    app._savedFileContent = 'saved source';
    app._updateEditorLabel();
    expect(document.getElementById('model-editor-label').textContent).toBe('client.ivy [saved]');

    editor.triggerChange('edited source');

    expect(document.getElementById('model-editor-label').textContent).toBe('** client.ivy');
    expect(document.getElementById('model-editor-label').textContent).not.toContain('[saved]');
  });
});
