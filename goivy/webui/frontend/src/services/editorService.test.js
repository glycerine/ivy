import { describe, expect, it, vi } from 'vitest';
import {
  editorContent,
  editorDirty,
  getEditorKeymap,
  refreshEditorLayout,
  scrollEditorToLine,
  setEditorContent,
  setEditorKeymap,
  updateEditorLabel,
} from './editorService.js';

function appWithContent(content = 'saved') {
  return {
    _persistedFilePath: 'client.ivy',
    _persistedFileContent: content,
    _savedFileContent: content,
    _saveInProgress: false,
    _updateReopenLastFileButton: vi.fn(),
    cmEditor: {
      getValue: vi.fn(() => content),
      setValue: vi.fn(),
    },
  };
}

describe('editorService', () => {
  it('reads content and dirty state from CodeMirror when present', () => {
    const app = appWithContent('changed');
    app._savedFileContent = 'saved';

    expect(editorContent(app)).toBe('changed');
    expect(editorDirty(app)).toBe(true);
  });

  it('syncs legacy editor state into the Vue bridge label model', () => {
    const app = appWithContent('saved');
    const bridge = {
      updateEditor: vi.fn(),
    };

    expect(updateEditorLabel(app, {
      bridge,
      updateReopenLastFileButton: app._updateReopenLastFileButton,
    })).toBe('client.ivy [saved]');
    expect(bridge.updateEditor).toHaveBeenCalledWith({
      path: 'client.ivy',
      content: 'saved',
      savedContent: 'saved',
      saveInProgress: false,
    });
    expect(app._updateReopenLastFileButton).toHaveBeenCalledTimes(1);
  });

  it('sets editor content and marks it saved', () => {
    const app = appWithContent('');
    const bridge = {
      updateEditor: vi.fn(),
    };
    window.__ivyVueBridge = bridge;

    setEditorContent(app, 'new text');

    expect(app._persistedFileContent).toBe('new text');
    expect(app._savedFileContent).toBe('new text');
    expect(app.cmEditor.setValue).toHaveBeenCalledWith('new text');
    expect(bridge.updateEditor).toHaveBeenCalled();
    delete window.__ivyVueBridge;
  });

  it('refreshes CodeMirror immediately and on queued layout turns', () => {
    const app = {
      cmEditor: {
        refresh: vi.fn(),
      },
    };
    const win = {
      requestAnimationFrame: vi.fn((fn) => fn()),
      setTimeout: vi.fn((fn) => fn()),
    };

    refreshEditorLayout(app, win);

    expect(app.cmEditor.refresh).toHaveBeenCalledTimes(3);
  });

  it('scrolls and highlights a source line', () => {
    const app = {
      cmEditor: {
        removeLineClass: vi.fn(),
        setCursor: vi.fn(),
        setSelection: vi.fn(),
        getLine: vi.fn(() => 'abcdef'),
        addLineClass: vi.fn(() => 'handle'),
        scrollIntoView: vi.fn(),
        focus: vi.fn(),
      },
    };

    scrollEditorToLine(app, 3);

    expect(app.cmEditor.setCursor).toHaveBeenCalledWith(2, 0);
    expect(app.cmEditor.addLineClass).toHaveBeenCalledWith(2, 'background', 'ivy-source-highlight');
    expect(app._highlightedEditorLine).toBe(3);
  });

  it('gets and sets keymaps through the bridge', () => {
    const bridge = {
      getEditorKeymap: vi.fn(() => 'vim'),
      setEditorKeymap: vi.fn(),
    };
    const app = {
      cmEditor: {
        setOption: vi.fn(),
      },
    };

    expect(getEditorKeymap({ bridge })).toBe('vim');
    expect(setEditorKeymap(app, 'emacs', { bridge })).toBe('emacs');
    expect(app.cmEditor.setOption).toHaveBeenCalledWith('keyMap', 'emacs');
    expect(bridge.setEditorKeymap).toHaveBeenCalledWith('emacs');
  });
});
