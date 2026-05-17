import { afterEach, describe, expect, it, vi } from 'vitest';
import { initializeCodeMirrorEditor } from './codeMirrorEditor.ts';

describe('codeMirrorEditor', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

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
      getLine: vi.fn(() => 'last line'),
      hasFocus: vi.fn(() => true),
      lastLine: vi.fn(() => 4),
      on: vi.fn(),
      setCursor: vi.fn(),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    const greaterThan = new KeyboardEvent('keydown', { key: '>', bubbles: true, cancelable: true });
    doc.dispatchEvent(greaterThan);

    expect(editor.setCursor).toHaveBeenCalledWith(4, 'last line'.length);
    expect(greaterThan.defaultPrevented).toBe(true);
  });

  it('maps Escape then < to cursorStart while the editor has focus', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      firstLine: vi.fn(() => 2),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
      setCursor: vi.fn(),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    const lessThan = new KeyboardEvent('keydown', { key: '<', bubbles: true, cancelable: true });
    doc.dispatchEvent(lessThan);

    expect(editor.setCursor).toHaveBeenCalledWith(2, 0);
    expect(lessThan.defaultPrevented).toBe(true);
  });

  it('maps Escape then v to cursorPageUp while the editor has focus', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      defaultTextHeight: vi.fn(() => 20),
      firstLine: vi.fn(() => 0),
      getCursor: vi.fn(() => ({ line: 30, ch: 6 })),
      getLine: vi.fn(() => 'short'),
      getScrollInfo: vi.fn(() => ({ clientHeight: 100 })),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
      setCursor: vi.fn(),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    const pageUp = new KeyboardEvent('keydown', { key: 'v', bubbles: true, cancelable: true });
    doc.dispatchEvent(pageUp);

    expect(editor.setCursor).toHaveBeenCalledWith(26, 'short'.length);
    expect(pageUp.defaultPrevented).toBe(true);
  });

  it('uses emacs as the default CodeMirror keymap', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const codeMirror = {
      fromTextArea: vi.fn(() => ({ on: vi.fn() })),
    };

    initializeCodeMirrorEditor({ codeMirror });

    expect(codeMirror.fromTextArea).toHaveBeenCalledWith(
      document.getElementById('model-editor'),
      expect.objectContaining({ keyMap: 'emacs' }),
    );
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

  it('cursorEnd collapses selection instead of delegating to CodeMirror goDocEnd', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getLine: vi.fn(() => 'last line'),
      lastLine: vi.fn(() => 7),
      on: vi.fn(),
      setCursor: vi.fn(),
    };
    const codeMirror = {
      commands: { goDocEnd: vi.fn() },
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ codeMirror });
    codeMirror.commands.cursorEnd(editor);

    expect(codeMirror.commands.goDocEnd).not.toHaveBeenCalled();
    expect(editor.setCursor).toHaveBeenCalledWith(7, 'last line'.length);
  });

  it('installs cursorPageUp as a safe CodeMirror command alias', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      defaultTextHeight: vi.fn(() => 10),
      firstLine: vi.fn(() => 0),
      focus: vi.fn(),
      getCursor: vi.fn(() => ({ line: 12, ch: 3 })),
      getLine: vi.fn(() => 'abc'),
      getScrollInfo: vi.fn(() => ({ clientHeight: 50 })),
      on: vi.fn(),
      scrollIntoView: vi.fn(),
      setCursor: vi.fn(),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ codeMirror });
    codeMirror.commands.cursorPageUp(editor);

    expect(editor.setCursor).toHaveBeenCalledWith(8, 3);
    expect(editor.scrollIntoView).toHaveBeenCalledWith({ line: 8, ch: 3 }, 50);
  });

  it('collapses Ctrl-y yank selection at the post-yank cursor in Emacs keymap', () => {
    vi.useFakeTimers();
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getCursor: vi.fn(() => ({ line: 10, ch: 12 })),
      getOption: vi.fn(() => 'emacs'),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
      setCursor: vi.fn(),
      somethingSelected: vi.fn(() => true),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });
    doc.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'y',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    }));
    vi.runAllTimers();

    expect(editor.setCursor).toHaveBeenCalledWith(10, 12);
  });

  it('does not collapse Ctrl-y selection outside the Emacs keymap', () => {
    vi.useFakeTimers();
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getCursor: vi.fn(() => ({ line: 10, ch: 12 })),
      getOption: vi.fn(() => 'sublime'),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
      setCursor: vi.fn(),
      somethingSelected: vi.fn(() => true),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });
    doc.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'y',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    }));
    vi.runAllTimers();

    expect(editor.setCursor).not.toHaveBeenCalled();
  });
});
