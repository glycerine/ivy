import { afterEach, describe, expect, it, vi } from 'vitest';
import { initializeCodeMirrorEditor } from './codeMirrorEditor.ts';

describe('codeMirrorEditor', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  function makeSearchCursor(match: { from: any; to: any }) {
    return {
      findNext: vi.fn(() => true),
      findPrevious: vi.fn(() => true),
      from: vi.fn(() => match.from),
      to: vi.fn(() => match.to),
    };
  }

  function installReplaceDialog(wrapper: HTMLElement, clicks: string[]) {
    wrapper.innerHTML = [
      '<div class="CodeMirror-dialog">',
      '  Replace?',
      '  <button type="button">Yes</button>',
      '  <button type="button">No</button>',
      '  <button type="button">All</button>',
      '  <button type="button">Stop</button>',
      '</div>',
    ].join('');
    Array.from(wrapper.querySelectorAll('button')).forEach((button) => {
      button.addEventListener('click', () => clicks.push((button.textContent || '').trim()));
    });
  }

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
        styleSelectedText: true,
        tabSize: 4,
        extraKeys: expect.objectContaining({
          'Ctrl-F': 'find',
          'Cmd-F': 'find',
          'Ctrl-X': expect.any(Function),
          'Ctrl-S': expect.any(Function),
          'Ctrl-R': expect.any(Function),
          'Ctrl-W': expect.any(Function),
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

  it('opens Ivy Emacs I-search on Ctrl-S and finds again on Ctrl-S inside the prompt', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = document.createElement('div');
    document.body.appendChild(wrapper);
    const closeDialog = vi.fn();
    const firstCursor = makeSearchCursor({ from: { line: 2, ch: 0 }, to: { line: 2, ch: 4 } });
    const secondCursor = makeSearchCursor({ from: { line: 4, ch: 0 }, to: { line: 4, ch: 4 } });
    const firstMark = { clear: vi.fn() };
    const secondMark = { clear: vi.fn() };
    const editor = {
      focus: vi.fn(),
      getCursor: vi.fn(() => ({ line: 1, ch: 2 })),
      getOption: vi.fn(() => 'emacs'),
      getSearchCursor: vi.fn()
        .mockReturnValueOnce(firstCursor)
        .mockReturnValueOnce(secondCursor),
      getWrapperElement: vi.fn(() => wrapper),
      markText: vi.fn()
        .mockReturnValueOnce(firstMark)
        .mockReturnValueOnce(secondMark),
      on: vi.fn(),
      openDialog: vi.fn((html) => {
        wrapper.innerHTML = `<div class="CodeMirror-dialog">${html}</div>`;
        return closeDialog;
      }),
      scrollIntoView: vi.fn(),
      setCursor: vi.fn(),
      setSelection: vi.fn(),
    };
    const codeMirror = {
      Pass: Symbol('CodeMirror.Pass'),
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ codeMirror });
    const options = codeMirror.fromTextArea.mock.calls[0][1];

    expect(options.extraKeys['Ctrl-S'](editor)).toBe(true);
    expect(editor.openDialog.mock.calls[0][0]).toContain('I-search:');
    expect(editor.openDialog.mock.calls[0][0]).not.toContain('Search:');

    const input = wrapper.querySelector('input') as HTMLInputElement;
    input.value = 'link';
    input.dispatchEvent(new Event('input', { bubbles: true }));

    expect(editor.getSearchCursor).toHaveBeenNthCalledWith(1, 'link', { line: 1, ch: 2 }, true);
    expect(firstCursor.findNext).toHaveBeenCalled();
    expect(editor.markText).toHaveBeenNthCalledWith(1, { line: 2, ch: 0 }, { line: 2, ch: 4 }, { className: 'ivy-emacs-isearch-match' });
    expect(editor.setSelection).toHaveBeenNthCalledWith(1, { line: 2, ch: 0 }, { line: 2, ch: 4 });

    const repeat = new KeyboardEvent('keydown', {
      key: 's',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    input.dispatchEvent(repeat);

    expect(repeat.defaultPrevented).toBe(true);
    expect(editor.getSearchCursor).toHaveBeenNthCalledWith(2, 'link', { line: 2, ch: 4 }, true);
    expect(secondCursor.findNext).toHaveBeenCalled();
    expect(firstMark.clear).toHaveBeenCalledTimes(1);
    expect(editor.markText).toHaveBeenNthCalledWith(2, { line: 4, ch: 0 }, { line: 4, ch: 4 }, { className: 'ivy-emacs-isearch-match' });
    expect(editor.setSelection).toHaveBeenNthCalledWith(2, { line: 4, ch: 0 }, { line: 4, ch: 4 });

    input.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'Enter',
      bubbles: true,
      cancelable: true,
    }));

    expect(closeDialog).toHaveBeenCalledTimes(1);
    expect(secondMark.clear).toHaveBeenCalledTimes(1);
    expect(editor.setCursor).toHaveBeenCalledWith(4, 4);
    expect(wrapper.querySelector('.CodeMirror-dialog')).toBeNull();
  });

  it('switches Ivy Emacs I-search to case-sensitive once the query has uppercase', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = document.createElement('div');
    document.body.appendChild(wrapper);
    const lowerCursor = makeSearchCursor({ from: { line: 2, ch: 0 }, to: { line: 2, ch: 4 } });
    const upperCursor = makeSearchCursor({ from: { line: 3, ch: 0 }, to: { line: 3, ch: 4 } });
    const editor = {
      getCursor: vi.fn(() => ({ line: 1, ch: 2 })),
      getOption: vi.fn(() => 'emacs'),
      getSearchCursor: vi.fn()
        .mockReturnValueOnce(lowerCursor)
        .mockReturnValueOnce(upperCursor),
      getWrapperElement: vi.fn(() => wrapper),
      on: vi.fn(),
      openDialog: vi.fn((html) => {
        wrapper.innerHTML = `<div class="CodeMirror-dialog">${html}</div>`;
        return vi.fn();
      }),
      scrollIntoView: vi.fn(),
      setCursor: vi.fn(),
      setSelection: vi.fn(),
    };
    const codeMirror = {
      Pass: Symbol('CodeMirror.Pass'),
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ codeMirror });
    const options = codeMirror.fromTextArea.mock.calls[0][1];

    options.extraKeys['Ctrl-S'](editor);
    const input = wrapper.querySelector('input') as HTMLInputElement;
    input.value = 'link';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    input.value = 'Link';
    input.dispatchEvent(new Event('input', { bubbles: true }));

    expect(editor.getSearchCursor).toHaveBeenNthCalledWith(1, 'link', { line: 1, ch: 2 }, true);
    expect(editor.getSearchCursor).toHaveBeenNthCalledWith(2, 'Link', { line: 1, ch: 2 }, false);
  });

  it('exits Ivy Emacs I-search on arrow keys and moves from the current match', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = document.createElement('div');
    document.body.appendChild(wrapper);
    const closeDialog = vi.fn();
    const cursor = makeSearchCursor({ from: { line: 2, ch: 0 }, to: { line: 2, ch: 4 } });
    const mark = { clear: vi.fn() };
    const editor = {
      execCommand: vi.fn(),
      focus: vi.fn(),
      getCursor: vi.fn(() => ({ line: 1, ch: 2 })),
      getOption: vi.fn(() => 'emacs'),
      getSearchCursor: vi.fn(() => cursor),
      getWrapperElement: vi.fn(() => wrapper),
      markText: vi.fn(() => mark),
      on: vi.fn(),
      openDialog: vi.fn((html) => {
        wrapper.innerHTML = `<div class="CodeMirror-dialog">${html}</div>`;
        return closeDialog;
      }),
      scrollIntoView: vi.fn(),
      setCursor: vi.fn(),
      setSelection: vi.fn(),
    };
    const codeMirror = {
      Pass: Symbol('CodeMirror.Pass'),
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ codeMirror });
    const options = codeMirror.fromTextArea.mock.calls[0][1];

    options.extraKeys['Ctrl-S'](editor);
    const input = wrapper.querySelector('input') as HTMLInputElement;
    input.value = 'link';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    const arrow = new KeyboardEvent('keydown', {
      key: 'ArrowRight',
      bubbles: true,
      cancelable: true,
    });
    input.dispatchEvent(arrow);

    expect(arrow.defaultPrevented).toBe(true);
    expect(closeDialog).toHaveBeenCalledTimes(1);
    expect(mark.clear).toHaveBeenCalledTimes(1);
    expect(editor.setCursor).toHaveBeenCalledWith(2, 4);
    expect(editor.focus).toHaveBeenCalled();
    expect(editor.execCommand).toHaveBeenCalledWith('goCharRight');
  });

  it('opens Ivy Emacs reverse I-search on Ctrl-R', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = document.createElement('div');
    document.body.appendChild(wrapper);
    const cursor = makeSearchCursor({ from: { line: 1, ch: 0 }, to: { line: 1, ch: 4 } });
    const editor = {
      getCursor: vi.fn(() => ({ line: 3, ch: 1 })),
      getLine: vi.fn(() => 'type server'),
      getOption: vi.fn(() => 'emacs'),
      getSearchCursor: vi.fn(() => cursor),
      getWrapperElement: vi.fn(() => wrapper),
      lastLine: vi.fn(() => 10),
      on: vi.fn(),
      openDialog: vi.fn((html) => {
        wrapper.innerHTML = `<div class="CodeMirror-dialog">${html}</div>`;
        return vi.fn();
      }),
      scrollIntoView: vi.fn(),
      setCursor: vi.fn(),
      setSelection: vi.fn(),
    };
    const codeMirror = {
      Pass: Symbol('CodeMirror.Pass'),
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ codeMirror });
    const options = codeMirror.fromTextArea.mock.calls[0][1];

    expect(options.extraKeys['Ctrl-R'](editor)).toBe(true);
    expect(editor.openDialog.mock.calls[0][0]).toContain('I-search backward:');
    const input = wrapper.querySelector('input') as HTMLInputElement;
    input.value = 'type';
    input.dispatchEvent(new Event('input', { bubbles: true }));

    expect(cursor.findPrevious).toHaveBeenCalled();
    expect(editor.getSearchCursor).toHaveBeenCalledWith('type', { line: 3, ch: 1 }, true);
    expect(editor.setSelection).toHaveBeenCalledWith({ line: 1, ch: 0 }, { line: 1, ch: 4 });

    input.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'Enter',
      bubbles: true,
      cancelable: true,
    }));

    expect(editor.setCursor).toHaveBeenCalledWith(1, 0);
  });

  it('leaves Ctrl-S and Ctrl-R available to other keymaps', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getOption: vi.fn(() => 'vim'),
      on: vi.fn(),
    };
    const codeMirror = {
      Pass: Symbol('CodeMirror.Pass'),
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ codeMirror });
    const options = codeMirror.fromTextArea.mock.calls[0][1];

    expect(options.extraKeys['Ctrl-S'](editor)).toBe(codeMirror.Pass);
    expect(options.extraKeys['Ctrl-R'](editor)).toBe(codeMirror.Pass);
  });

  it('maps Ctrl-x Ctrl-s to save through CodeMirror while the Emacs editor has focus', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getOption: vi.fn(() => 'emacs'),
      on: vi.fn(),
    };
    const runtime = {
      save: vi.fn(),
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ runtime, codeMirror });
    const options = codeMirror.fromTextArea.mock.calls[0][1];

    expect(options.extraKeys['Ctrl-X'](editor)).toBe(true);
    expect(options.extraKeys['Ctrl-S'](editor)).toBe(true);
    expect(runtime.save).toHaveBeenCalledTimes(1);
  });

  it('maps Ctrl-x Ctrl-w to Save As through CodeMirror while the Emacs editor has focus', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getOption: vi.fn(() => 'emacs'),
      on: vi.fn(),
    };
    const runtime = {
      save: vi.fn(),
      saveAs: vi.fn(),
    };
    const codeMirror = {
      Pass: Symbol('CodeMirror.Pass'),
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ runtime, codeMirror });
    const options = codeMirror.fromTextArea.mock.calls[0][1];

    expect(options.extraKeys['Ctrl-W'](editor)).toBe(codeMirror.Pass);
    expect(runtime.saveAs).not.toHaveBeenCalled();

    expect(options.extraKeys['Ctrl-X'](editor)).toBe(true);
    expect(options.extraKeys['Ctrl-W'](editor)).toBe(true);
    expect(runtime.saveAs).toHaveBeenCalledTimes(1);
    expect(runtime.save).not.toHaveBeenCalled();
  });

  it('does not map Ctrl-x Ctrl-s to save outside the Emacs keymap', () => {
    document.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getOption: vi.fn(() => 'vim'),
      on: vi.fn(),
    };
    const runtime = {
      save: vi.fn(),
      saveAs: vi.fn(),
    };
    const codeMirror = {
      Pass: Symbol('CodeMirror.Pass'),
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ runtime, codeMirror });
    const options = codeMirror.fromTextArea.mock.calls[0][1];

    expect(options.extraKeys['Ctrl-X'](editor)).toBe(codeMirror.Pass);
    expect(options.extraKeys['Ctrl-S'](editor)).toBe(codeMirror.Pass);
    expect(options.extraKeys['Ctrl-W'](editor)).toBe(codeMirror.Pass);
    expect(runtime.save).not.toHaveBeenCalled();
    expect(runtime.saveAs).not.toHaveBeenCalled();
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

  it('maps Escape then > to extend an active Emacs mark to the end of the buffer', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const anchor = { line: 1, ch: 3 };
    const editor = {
      getCursor: vi.fn((which) => (which === 'anchor' ? anchor : { line: 2, ch: 0 })),
      getExtending: vi.fn(() => true),
      getLine: vi.fn(() => 'last line'),
      hasFocus: vi.fn(() => true),
      lastLine: vi.fn(() => 4),
      on: vi.fn(),
      setCursor: vi.fn(),
      setExtending: vi.fn(),
      setSelection: vi.fn(),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    doc.dispatchEvent(new KeyboardEvent('keydown', { key: '>', bubbles: true, cancelable: true }));

    expect(editor.setCursor).not.toHaveBeenCalled();
    expect(editor.setSelection).toHaveBeenCalledWith(anchor, { line: 4, ch: 'last line'.length });
    expect(editor.setExtending).toHaveBeenCalledWith(true);
  });

  it('maps Escape then % to CodeMirror replace in the Emacs keymap', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getOption: vi.fn(() => 'emacs'),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
    };
    const replace = vi.fn();
    const codeMirror = {
      commands: { replace },
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    const replaceChord = new KeyboardEvent('keydown', { key: '%', bubbles: true, cancelable: true });
    doc.dispatchEvent(replaceChord);

    expect(replaceChord.defaultPrevented).toBe(true);
    expect(replace).toHaveBeenCalledWith(editor);
  });

  it('does not map Escape then % outside the Emacs keymap', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getOption: vi.fn(() => 'vim'),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
    };
    const replace = vi.fn();
    const codeMirror = {
      commands: { replace },
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    const replaceChord = new KeyboardEvent('keydown', { key: '%', bubbles: true, cancelable: true });
    doc.dispatchEvent(replaceChord);

    expect(replaceChord.defaultPrevented).toBe(false);
    expect(replace).not.toHaveBeenCalled();
  });

  it('maps Emacs replace prompt response keys to CodeMirror replace buttons', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    const clicks: string[] = [];
    installReplaceDialog(wrapper, clicks);
    doc.body.appendChild(wrapper);
    const editor = {
      getOption: vi.fn(() => 'emacs'),
      getWrapperElement: vi.fn(() => wrapper),
      on: vi.fn(),
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    for (const key of ['y', ' ', 'n', 'Backspace', 'Delete', '!', 'q', 'Enter']) {
      const event = new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true });
      doc.dispatchEvent(event);
      expect(event.defaultPrevented).toBe(true);
    }

    expect(clicks).toEqual(['Yes', 'Yes', 'No', 'No', 'No', 'All', 'Stop', 'Stop']);
  });

  it('maps Ctrl-g to Stop in the Emacs replace question prompt', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    const clicks: string[] = [];
    installReplaceDialog(wrapper, clicks);
    doc.body.appendChild(wrapper);
    const editor = {
      getOption: vi.fn(() => 'emacs'),
      getWrapperElement: vi.fn(() => wrapper),
      on: vi.fn(),
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    const event = new KeyboardEvent('keydown', {
      key: 'g',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    doc.dispatchEvent(event);

    expect(event.defaultPrevented).toBe(true);
    expect(clicks).toEqual(['Stop']);
  });

  it('aborts the Emacs replace input prompt directly on Ctrl-g', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    wrapper.innerHTML = '<div class="CodeMirror-dialog">Replace: <input value="link"></div>';
    doc.body.appendChild(wrapper);
    const input = wrapper.querySelector('input') as HTMLInputElement;
    const escapeHandler = vi.fn((event) => {
      if (event.key === 'Escape') event.preventDefault();
    });
    input.addEventListener('keydown', escapeHandler);
    const editor = {
      focus: vi.fn(),
      getOption: vi.fn(() => 'emacs'),
      getWrapperElement: vi.fn(() => wrapper),
      on: vi.fn(),
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    const event = new KeyboardEvent('keydown', {
      key: 'g',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    input.dispatchEvent(event);

    expect(event.defaultPrevented).toBe(true);
    expect(escapeHandler).not.toHaveBeenCalled();
    expect(wrapper.querySelector('.CodeMirror-dialog')).toBeNull();
    expect(editor.focus).toHaveBeenCalled();
  });

  it('does not arm Escape-percent while Escape aborts an active replace input prompt', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    wrapper.innerHTML = '<div class="CodeMirror-dialog">Replace: <input value="link"></div>';
    doc.body.appendChild(wrapper);
    const editor = {
      getOption: vi.fn(() => 'emacs'),
      getWrapperElement: vi.fn(() => wrapper),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
    };
    const replace = vi.fn();
    const codeMirror = {
      commands: { replace },
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    wrapper.innerHTML = '';
    const percent = new KeyboardEvent('keydown', { key: '%', bubbles: true, cancelable: true });
    doc.dispatchEvent(percent);

    expect(percent.defaultPrevented).toBe(false);
    expect(replace).not.toHaveBeenCalled();
  });

  it('does not arm Escape-percent after Ctrl-g aborts an active replace input prompt', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    wrapper.innerHTML = '<div class="CodeMirror-dialog">Replace: <input value="link"></div>';
    doc.body.appendChild(wrapper);
    const input = wrapper.querySelector('input') as HTMLInputElement;
    const editor = {
      focus: vi.fn(),
      getOption: vi.fn(() => 'emacs'),
      getWrapperElement: vi.fn(() => wrapper),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
    };
    const replace = vi.fn();
    const codeMirror = {
      commands: { replace },
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    input.dispatchEvent(new KeyboardEvent('keydown', {
      key: 'g',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    }));
    const percent = new KeyboardEvent('keydown', { key: '%', bubbles: true, cancelable: true });
    doc.dispatchEvent(percent);

    expect(percent.defaultPrevented).toBe(false);
    expect(replace).not.toHaveBeenCalled();
  });

  it('maps period in the Emacs replace prompt to replace once and stop', () => {
    vi.useFakeTimers();
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    const clicks: string[] = [];
    installReplaceDialog(wrapper, clicks);
    doc.body.appendChild(wrapper);
    const editor = {
      getOption: vi.fn(() => 'emacs'),
      getWrapperElement: vi.fn(() => wrapper),
      on: vi.fn(),
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    const event = new KeyboardEvent('keydown', { key: '.', bubbles: true, cancelable: true });
    doc.dispatchEvent(event);
    vi.runAllTimers();

    expect(event.defaultPrevented).toBe(true);
    expect(clicks).toEqual(['Yes', 'Stop']);
  });

  it('stops Emacs replace prompt responses at the last match instead of wrapping', () => {
    vi.useFakeTimers();
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    const clicks: string[] = [];
    installReplaceDialog(wrapper, clicks);
    doc.body.appendChild(wrapper);
    const cursor = { findNext: vi.fn(() => false) };
    const editor = {
      getCursor: vi.fn((which) => (which === 'from' ? { line: 4, ch: 2 } : { line: 4, ch: 6 })),
      getOption: vi.fn(() => 'emacs'),
      getSearchCursor: vi.fn(() => cursor),
      getWrapperElement: vi.fn(() => wrapper),
      on: vi.fn(),
      scrollIntoView: vi.fn(),
      state: { search: { query: 'link' } },
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    const noEvent = new KeyboardEvent('keydown', { key: 'n', bubbles: true, cancelable: true });
    doc.dispatchEvent(noEvent);
    vi.runAllTimers();

    expect(noEvent.defaultPrevented).toBe(true);
    expect(clicks).toEqual(['Stop']);
    expect(editor.getSearchCursor).toHaveBeenCalledWith('link', { line: 4, ch: 6 }, { caseFold: true });

    clicks.length = 0;
    cursor.findNext.mockClear();
    installReplaceDialog(wrapper, clicks);
    const yesEvent = new KeyboardEvent('keydown', { key: 'y', bubbles: true, cancelable: true });
    doc.dispatchEvent(yesEvent);
    vi.runAllTimers();

    expect(yesEvent.defaultPrevented).toBe(true);
    expect(clicks).toEqual(['Yes', 'Stop']);
  });

  it('centers the active Emacs replace prompt match before asking about the next match', () => {
    vi.useFakeTimers();
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    const clicks: string[] = [];
    installReplaceDialog(wrapper, clicks);
    doc.body.appendChild(wrapper);
    const editor = {
      charCoords: vi.fn(() => ({ top: 400, bottom: 420 })),
      getCursor: vi.fn((which) => (which === 'from' ? { line: 12, ch: 4 } : { line: 12, ch: 10 })),
      getOption: vi.fn(() => 'emacs'),
      getScrollerElement: vi.fn(() => ({ clientHeight: 100 })),
      getSearchCursor: vi.fn(() => ({ findNext: vi.fn(() => true) })),
      getWrapperElement: vi.fn(() => wrapper),
      on: vi.fn(),
      scrollTo: vi.fn(),
      state: { search: { query: 'client' } },
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'n', bubbles: true, cancelable: true }));
    vi.runAllTimers();

    expect(clicks).toEqual(['No']);
    expect(editor.scrollTo).toHaveBeenCalledWith(null, 360);
    expect(editor.scrollTo).toHaveBeenCalledTimes(2);
  });

  it('replaces rest of buffer for ! without wrapping to earlier matches', () => {
    vi.useFakeTimers();
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    const clicks: string[] = [];
    installReplaceDialog(wrapper, clicks);
    doc.body.appendChild(wrapper);
    const ranges = [
      { from: { line: 4, ch: 1 }, to: { line: 4, ch: 5 } },
      { from: { line: 8, ch: 2 }, to: { line: 8, ch: 6 } },
    ];
    let currentMatch = -1;
    const cursor = {
      findNext: vi.fn(() => {
        currentMatch += 1;
        return currentMatch < ranges.length;
      }),
      replace: vi.fn(),
      to: vi.fn(() => ranges[Math.max(0, currentMatch)].to),
    };
    const editor = {
      getCursor: vi.fn((which) => {
        const range = ranges[0];
        return which === 'from' ? range.from : range.to;
      }),
      getLine: vi.fn(() => 'last line'),
      getOption: vi.fn(() => 'emacs'),
      getSearchCursor: vi.fn(() => cursor),
      getWrapperElement: vi.fn(() => wrapper),
      lastLine: vi.fn(() => 20),
      on: vi.fn(),
      operation: vi.fn((fn) => fn()),
      scrollIntoView: vi.fn(),
      setCursor: vi.fn(),
      state: { search: { query: 'link' } },
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });
    wrapper.innerHTML = '<div class="CodeMirror-dialog">With: <input value="node\\n"></div>';
    const input = wrapper.querySelector('input') as HTMLInputElement;
    input.dispatchEvent(new Event('input', { bubbles: true }));
    installReplaceDialog(wrapper, clicks);

    const event = new KeyboardEvent('keydown', { key: '!', bubbles: true, cancelable: true });
    doc.dispatchEvent(event);
    vi.runAllTimers();

    expect(event.defaultPrevented).toBe(true);
    expect(clicks).toEqual(['Stop']);
    expect(clicks).not.toContain('All');
    expect(editor.operation).toHaveBeenCalledTimes(1);
    expect(editor.getSearchCursor).toHaveBeenCalledWith('link', { line: 4, ch: 1 }, { caseFold: true });
    expect(cursor.replace).toHaveBeenCalledTimes(2);
    expect(cursor.replace).toHaveBeenNthCalledWith(1, 'node\n');
    expect(cursor.replace).toHaveBeenNthCalledWith(2, 'node\n');
    expect(editor.setCursor).toHaveBeenCalledWith(8, 6);
  });

  it('does not map replace prompt response keys outside the Emacs keymap', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const wrapper = doc.createElement('div');
    const clicks: string[] = [];
    installReplaceDialog(wrapper, clicks);
    doc.body.appendChild(wrapper);
    const editor = {
      getOption: vi.fn(() => 'vim'),
      getWrapperElement: vi.fn(() => wrapper),
      on: vi.fn(),
    };
    const codeMirror = {
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    const event = new KeyboardEvent('keydown', { key: 'y', bubbles: true, cancelable: true });
    doc.dispatchEvent(event);

    expect(event.defaultPrevented).toBe(false);
    expect(clicks).toEqual([]);
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

  it('maps Escape then < to extend an active Emacs mark to the beginning of the buffer', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const anchor = { line: 5, ch: 2 };
    const editor = {
      firstLine: vi.fn(() => 0),
      getCursor: vi.fn((which) => (which === 'anchor' ? anchor : { line: 4, ch: 0 })),
      getExtending: vi.fn(() => true),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
      setCursor: vi.fn(),
      setExtending: vi.fn(),
      setSelection: vi.fn(),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    doc.dispatchEvent(new KeyboardEvent('keydown', { key: '<', bubbles: true, cancelable: true }));

    expect(editor.setCursor).not.toHaveBeenCalled();
    expect(editor.setSelection).toHaveBeenCalledWith(anchor, { line: 0, ch: 0 });
    expect(editor.setExtending).toHaveBeenCalledWith(true);
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

  it('maps Escape then w to copy the selection into the Emacs yank buffer', () => {
    vi.useFakeTimers();
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getCursor: vi.fn(() => ({ line: 8, ch: 15 })),
      getOption: vi.fn(() => 'emacs'),
      getSelection: vi.fn(() => 'copied text\nsecond line'),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
      replaceSelection: vi.fn(),
      setCursor: vi.fn(),
      setExtending: vi.fn(),
      somethingSelected: vi.fn(() => true),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    const copy = new KeyboardEvent('keydown', { key: 'w', bubbles: true, cancelable: true });
    doc.dispatchEvent(copy);

    expect(copy.defaultPrevented).toBe(true);
    expect(editor.getSelection).toHaveBeenCalled();
    expect(editor.setExtending).toHaveBeenCalledWith(false);
    expect(editor.setCursor).toHaveBeenCalledTimes(1);
    expect(editor.setCursor).toHaveBeenLastCalledWith(8, 15);

    const yank = new KeyboardEvent('keydown', {
      key: 'y',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    doc.dispatchEvent(yank);
    vi.runAllTimers();

    expect(yank.defaultPrevented).toBe(true);
    expect(editor.replaceSelection).toHaveBeenCalledWith('copied text\nsecond line', 'end');
    expect(editor.setCursor).toHaveBeenCalledTimes(2);
    expect(editor.setCursor).toHaveBeenLastCalledWith(8, 15);
  });

  it('does not install Escape then w as a yank-buffer copy outside the Emacs keymap', () => {
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getOption: vi.fn(() => 'sublime'),
      getSelection: vi.fn(() => 'copied text'),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
      replaceSelection: vi.fn(),
      somethingSelected: vi.fn(() => true),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    const copy = new KeyboardEvent('keydown', { key: 'w', bubbles: true, cancelable: true });
    doc.dispatchEvent(copy);

    expect(copy.defaultPrevented).toBe(false);
    expect(editor.getSelection).not.toHaveBeenCalled();
  });

  it('lets a native Emacs kill command replace the Escape-w yank buffer', () => {
    vi.useFakeTimers();
    const doc = document.implementation.createHTMLDocument('');
    doc.body.innerHTML = '<textarea id="model-editor"></textarea>';
    const editor = {
      getCursor: vi.fn(() => ({ line: 4, ch: 2 })),
      getOption: vi.fn(() => 'emacs'),
      getSelection: vi.fn(() => 'old copied text'),
      hasFocus: vi.fn(() => true),
      on: vi.fn(),
      replaceSelection: vi.fn(),
      setCursor: vi.fn(),
      somethingSelected: vi.fn()
        .mockReturnValueOnce(true)
        .mockReturnValue(false),
    };
    const codeMirror = {
      commands: {},
      fromTextArea: vi.fn(() => editor),
    };

    initializeCodeMirrorEditor({ doc, codeMirror });

    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    doc.dispatchEvent(new KeyboardEvent('keydown', { key: 'w', bubbles: true, cancelable: true }));
    const killLine = new KeyboardEvent('keydown', {
      key: 'k',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    doc.dispatchEvent(killLine);
    const yank = new KeyboardEvent('keydown', {
      key: 'y',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    doc.dispatchEvent(yank);
    vi.runAllTimers();

    expect(killLine.defaultPrevented).toBe(false);
    expect(yank.defaultPrevented).toBe(false);
    expect(editor.replaceSelection).not.toHaveBeenCalled();
    expect(editor.setCursor).toHaveBeenCalledTimes(1);
    expect(editor.setCursor).toHaveBeenCalledWith(4, 2);
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
