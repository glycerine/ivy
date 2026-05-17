import { describe, expect, it, vi } from 'vitest';
import { initializeCodeMirrorEditor } from './codeMirrorEditor.ts';

describe('codeMirrorEditor', () => {
  it('initializes CodeMirror with Ivy editor options and wires dirty tracking', () => {
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
    const runtime = {
      cmEditor: editor,
      _updateEditorLabel: vi.fn(),
    };

    const result = initializeCodeMirrorEditor({
      runtime,
      keymap: 'vim',
      codeMirror,
    });

    expect(result).toBe(editor);
    expect(codeMirror.fromTextArea).toHaveBeenCalledWith(
      document.getElementById('model-editor'),
      expect.objectContaining({
        lineNumbers: true,
        keyMap: 'vim',
        tabSize: 4,
        extraKeys: expect.objectContaining({
          'Ctrl-F': 'find',
          'Cmd-F': 'find',
          'Shift-Ctrl-F': 'replace',
          'Cmd-Alt-F': 'replace',
          'Shift-Ctrl-R': 'replaceAll',
          'Shift-Cmd-Alt-F': 'replaceAll',
        }),
      }),
    );

    value = 'edited';
    changeHandler();

    expect(runtime._persistedFileContent).toBe('edited');
    expect(runtime._updateEditorLabel).toHaveBeenCalled();
  });

  it('reuses an existing editor for the textarea', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const existing = {};
    document.getElementById('model-editor').__ivyCodeMirrorEditor = existing;

    expect(initializeCodeMirrorEditor({ codeMirror: { fromTextArea: vi.fn() } })).toBe(existing);
  });
});
