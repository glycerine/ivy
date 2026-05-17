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
} from './editorService.ts';

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
      clearHistory: vi.fn(),
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

  it('updates the editor label DOM from runtime editor state', () => {
    document.body.innerHTML = '<span id="model-editor-label"></span><span id="loaded-file"></span>';
    const app = appWithContent('saved');

    expect(updateEditorLabel(app, {
      updateReopenLastFileButton: app._updateReopenLastFileButton,
    })).toBe('client.ivy [saved]');
    expect(document.getElementById('model-editor-label').textContent).toBe('client.ivy [saved]');
    expect(document.getElementById('loaded-file').textContent).toBe('client.ivy');
    expect(app._updateReopenLastFileButton).toHaveBeenCalledTimes(1);
  });

  it('sets editor content and marks it saved', () => {
    const app = appWithContent('');

    setEditorContent(app, 'new text');

    expect(app._persistedFileContent).toBe('new text');
    expect(app._savedFileContent).toBe('new text');
    expect(app.cmEditor.setValue).toHaveBeenCalledWith('new text');
    expect(app.cmEditor.clearHistory).toHaveBeenCalledTimes(1);
  });

  it('does not require CodeMirror clearHistory when setting content', () => {
    const app = appWithContent('');
    delete app.cmEditor.clearHistory;

    expect(() => setEditorContent(app, 'new text')).not.toThrow();
    expect(app.cmEditor.setValue).toHaveBeenCalledWith('new text');
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

  it('gets and sets keymaps through the DOM radios', () => {
    document.body.innerHTML = [
      '<label><input type="radio" name="keymap" value="sublime" checked></label>',
      '<label><input type="radio" name="keymap" value="emacs"></label>',
      '<label><input type="radio" name="keymap" value="vim"></label>',
    ].join('');
    const app = {
      cmEditor: {
        setOption: vi.fn(),
      },
    };

    expect(getEditorKeymap({ doc: document })).toBe('sublime');
    expect(setEditorKeymap(app, 'vim', { doc: document })).toBe('vim');

    expect(app.cmEditor.setOption).toHaveBeenCalledWith('keyMap', 'vim');
    expect(document.querySelector('input[name="keymap"][value="sublime"]').checked).toBe(false);
    expect(document.querySelector('input[name="keymap"][value="vim"]').checked).toBe(true);
    expect(app._editorKeymap).toBe('vim');
  });

  it('defaults invalid or missing keymaps to emacs', () => {
    document.body.innerHTML = [
      '<label><input type="radio" name="keymap" value="sublime"></label>',
      '<label><input type="radio" name="keymap" value="emacs"></label>',
    ].join('');
    const app = {
      cmEditor: {
        setOption: vi.fn(),
      },
    };

    expect(getEditorKeymap({ doc: document })).toBe('emacs');
    expect(setEditorKeymap(app, 'made-up', { doc: document })).toBe('emacs');

    expect(app.cmEditor.setOption).toHaveBeenCalledWith('keyMap', 'emacs');
    expect(document.querySelector('input[name="keymap"][value="emacs"]').checked).toBe(true);
  });
});
