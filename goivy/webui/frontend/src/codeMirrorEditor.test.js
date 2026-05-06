import { describe, expect, it, vi } from 'vitest';
import { initializeLegacyCodeMirror } from './codeMirrorEditor.js';

describe('codeMirrorEditor', () => {
  it('initializes CodeMirror with Ivy editor options and wires legacy dirty tracking', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    let changeHandler = null;
    let value = 'saved';
    const editor = {
      getValue: vi.fn(() => value),
      on: vi.fn((event, handler) => {
        if (event === 'change') changeHandler = handler;
      }),
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };
    const legacyApp = {
      cmEditor: editor,
      _updateEditorLabel: vi.fn(),
    };

    const result = initializeLegacyCodeMirror({
      legacyApp,
      editorStore: { keymap: 'vim' },
      codeMirror,
    });

    expect(result).toBe(editor);
    expect(codeMirror.fromTextArea).toHaveBeenCalledWith(
      document.getElementById('model-editor'),
      expect.objectContaining({
        lineNumbers: true,
        keyMap: 'vim',
        tabSize: 4,
      }),
    );

    value = 'edited';
    changeHandler();

    expect(legacyApp._persistedFileContent).toBe('edited');
    expect(legacyApp._updateEditorLabel).toHaveBeenCalled();
  });

  it('reuses an existing editor for the textarea', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const existing = {};
    document.getElementById('model-editor').__ivyCodeMirrorEditor = existing;

    expect(initializeLegacyCodeMirror({ codeMirror: { fromTextArea: vi.fn() } })).toBe(existing);
  });
});
