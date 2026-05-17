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

  it('maps Escape then > to cursorEnd while the editor has focus', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
    };
    const cursorEnd = vi.fn();
    const codeMirror = {
      commands: { cursorEnd },
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    const greaterThan = new KeyboardEvent('keydown', { key: '>', bubbles: true, cancelable: true });
    doc.dispatchEvent(greaterThan);

    expect(cursorEnd).toHaveBeenCalledWith(editor);
    expect(greaterThan.defaultPrevented).toBe(true);
  });

  it('ignores Escape then > when the editor is not focused', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      hasFocus: vi.fn(() => false),
      on: vi.fn(),
    };
    const cursorEnd = vi.fn();
    const codeMirror = {
      commands: { cursorEnd },
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    doc.dispatchEvent(new KeyboardEvent('keydown', { key: '>', bubbles: true, cancelable: true }));

    expect(cursorEnd).not.toHaveBeenCalled();
  });

  it('installs cursorEnd as a safe CodeMirror command alias', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      focus: vi.fn(),
      getLine: vi.fn(() => 'last line'),
      lastLine: vi.fn(() => 7),
      on: vi.fn(),
      scrollIntoView: vi.fn(),
      setCursor: vi.fn(),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ codeMirror });
    codeMirror.commands.cursorEnd(editor);

    expect(editor.setCursor).toHaveBeenCalledWith(7, 'last line'.length);
    expect(editor.scrollIntoView).toHaveBeenCalledWith({ line: 7, ch: 'last line'.length }, 50);
  });
});
