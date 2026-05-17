export function initializeCodeMirrorEditor({
  runtime = null,
  keymap = 'emacs',
  doc = globalThis.document,
  codeMirror = globalThis.CodeMirror,
}: {
  runtime?: any;
  keymap?: string;
  doc?: Document;
  codeMirror?: any;
} = {}) {
  const modelEditor = doc && doc.getElementById('model-editor');
  if (!modelEditor) return null;
  if (modelEditor.__ivyCodeMirrorEditor) {
    return modelEditor.__ivyCodeMirrorEditor;
  }
  if (!codeMirror || typeof codeMirror.fromTextArea !== 'function') {
    throw new Error('CodeMirror is not available');
  }

  installBufferCursorCommands(codeMirror);

  const editor = codeMirror.fromTextArea(modelEditor, {
    lineNumbers: true,
    keyMap: keymap || 'emacs',
    tabSize: 4,
    indentUnit: 4,
    lineWrapping: false,
    matchBrackets: true,
    styleSelectedText: true,
    extraKeys: {
      'Ctrl-F': 'find',
      'Cmd-F': 'find',
      'Shift-Ctrl-F': 'replace',
      'Cmd-Alt-F': 'replace',
      'Shift-Ctrl-R': 'replaceAll',
      'Shift-Cmd-Alt-F': 'replaceAll',
      'Ctrl-Z': 'undo',
      'Ctrl-Shift-Z': 'redo',

      // alternative for jump to end / beginning of buffer.
      'Alt-Shift-,': 'cursorStart',
      'Alt-Shift-.': 'cursorEnd',
    },
  });
  modelEditor.__ivyCodeMirrorEditor = editor;

  installEscapeBufferChord({
    editor,
    codeMirror,
    doc,
  });
  installEmacsYankSelectionCollapse({
    editor,
    doc,
  });

  if (editor && typeof editor.on === 'function' && runtime) {
    editor.on('change', () => {
      if (!runtime.cmEditor || typeof runtime.cmEditor.getValue !== 'function') return;
      runtime._persistedFileContent = runtime.cmEditor.getValue();
      if (typeof runtime._updateEditorLabel === 'function') {
        runtime._updateEditorLabel();
      }
    });
  }

  return editor;
}

function installBufferCursorCommands(codeMirror: any) {
  if (!codeMirror.commands) codeMirror.commands = {};
  codeMirror.commands.cursorStart = (editor: any) => {
    const firstLine = typeof editor.firstLine === 'function' ? editor.firstLine() : 0;
    setEditorCursor(editor, firstLine, 0);
    return true;
  };
  codeMirror.commands.cursorEnd = (editor: any) => {
    const lastLine = typeof editor.lastLine === 'function'
      ? editor.lastLine()
      : Math.max(0, (typeof editor.lineCount === 'function' ? editor.lineCount() : 1) - 1);
    const lineText = typeof editor.getLine === 'function' ? editor.getLine(lastLine) || '' : '';
    setEditorCursor(editor, lastLine, lineText.length);
    return true;
  };
  codeMirror.commands.cursorPageUp = (editor: any) => {
    const cursor = typeof editor.getCursor === 'function' ? editor.getCursor() || {} : {};
    const currentLine = Number.isFinite(cursor.line) ? cursor.line : 0;
    const currentCh = Number.isFinite(cursor.ch) ? cursor.ch : 0;
    const firstLine = typeof editor.firstLine === 'function' ? editor.firstLine() : 0;
    const pageLines = editorPageLineCount(editor);
    const targetLine = Math.max(firstLine, currentLine - pageLines);
    const lineText = typeof editor.getLine === 'function' ? editor.getLine(targetLine) || '' : '';
    setEditorCursor(editor, targetLine, Math.min(currentCh, lineText.length));
    return true;
  };
}

function setEditorCursor(editor: any, line: number, ch: number) {
  if (typeof editor.setCursor === 'function') {
    editor.setCursor(line, ch);
  } else if (typeof editor.setSelection === 'function') {
    editor.setSelection({ line, ch }, { line, ch });
  }
  if (typeof editor.scrollIntoView === 'function') editor.scrollIntoView({ line, ch }, 50);
  if (typeof editor.focus === 'function') editor.focus();
}

function editorPageLineCount(editor: any) {
  const textHeight = typeof editor.defaultTextHeight === 'function' ? Number(editor.defaultTextHeight()) : 0;
  const scrollInfo = typeof editor.getScrollInfo === 'function' ? editor.getScrollInfo() || {} : {};
  const clientHeight = Number(scrollInfo.clientHeight) || 0;
  if (textHeight > 0 && clientHeight > 0) {
    return Math.max(1, Math.floor(clientHeight / textHeight) - 1);
  }
  return 20;
}

function installEscapeBufferChord({
  editor,
  codeMirror,
  doc,
}: {
  editor: any;
  codeMirror: any;
  doc: Document;
}) {
  if (!doc || !editor || !codeMirror) return;
  let lastEscapeAt = 0;
  const chordWindowMs = 2000;
  const isEditorFocused = () => {
    if (typeof editor.hasFocus === 'function') return editor.hasFocus();
    const wrapper = typeof editor.getWrapperElement === 'function' ? editor.getWrapperElement() : null;
    return !!(wrapper && doc.activeElement && wrapper.contains(doc.activeElement));
  };
  doc.addEventListener('keydown', (event) => {
    if (!isEditorFocused()) {
      lastEscapeAt = 0;
      return;
    }
    if (event.key === 'Escape') {
      lastEscapeAt = Date.now();
      return;
    }
    const inChordWindow = lastEscapeAt > 0 && Date.now() - lastEscapeAt <= chordWindowMs;
    const isLessThan = event.key === '<' || (event.key === ',' && event.shiftKey);
    const isGreaterThan = event.key === '>' || (event.key === '.' && event.shiftKey);
    const isPageUp = event.key === 'v';
    const isCopySelection = event.key === 'w'
      && !event.altKey
      && !event.ctrlKey
      && !event.metaKey
      && !event.shiftKey
      && isEmacsKeymap(editor);
    if ((isLessThan || isGreaterThan || isPageUp || isCopySelection) && inChordWindow) {
      lastEscapeAt = 0;
      event.preventDefault();
      event.stopPropagation();
      if (isCopySelection) {
        copySelectionToEmacsYankBuffer(editor);
        return;
      }
      const command = isLessThan
        ? codeMirror.commands.cursorStart
        : isPageUp
          ? codeMirror.commands.cursorPageUp
          : codeMirror.commands.cursorEnd;
      command(editor);
      return;
    }
    if (!event.altKey && !event.ctrlKey && !event.metaKey && event.key !== 'Shift') {
      lastEscapeAt = 0;
    }
  }, true);
}

function installEmacsYankSelectionCollapse({
  editor,
  doc,
}: {
  editor: any;
  doc: Document;
}) {
  if (!doc || !editor) return;
  const isEditorFocused = () => {
    if (typeof editor.hasFocus === 'function') return editor.hasFocus();
    const wrapper = typeof editor.getWrapperElement === 'function' ? editor.getWrapperElement() : null;
    return !!(wrapper && doc.activeElement && wrapper.contains(doc.activeElement));
  };
  doc.addEventListener('keydown', (event) => {
    if (!isEditorFocused() || !isEmacsKeymap(editor)) return;
    if (isNativeEmacsKillCommand(event)) {
      clearEmacsYankBuffer(editor);
      return;
    }
    if (event.key !== 'y' || !event.ctrlKey || event.altKey || event.metaKey || event.shiftKey) return;
    if (yankFromEmacsYankBuffer(editor)) {
      event.preventDefault();
      event.stopPropagation();
      collapseSelectionAtPostYankCursor(editor);
      return;
    }
    setTimeout(() => {
      if (typeof editor.somethingSelected === 'function' && !editor.somethingSelected()) return;
      collapseSelectionAtPostYankCursor(editor);
    }, 0);
  }, true);
}

function isEmacsKeymap(editor: any) {
  if (typeof editor.getOption !== 'function') return true;
  return editor.getOption('keyMap') === 'emacs';
}

function copySelectionToEmacsYankBuffer(editor: any) {
  if (typeof editor.somethingSelected === 'function' && !editor.somethingSelected()) {
    return false;
  }
  const selectedText = typeof editor.getSelection === 'function' ? editor.getSelection() : '';
  if (typeof selectedText !== 'string' || selectedText.length === 0) {
    return false;
  }
  editor.__ivyEmacsYankBuffer = selectedText;
  return true;
}

function yankFromEmacsYankBuffer(editor: any) {
  const yankText = editor.__ivyEmacsYankBuffer;
  if (typeof yankText !== 'string' || yankText.length === 0) {
    return false;
  }
  if (typeof editor.replaceSelection === 'function') {
    editor.replaceSelection(yankText, 'end');
    return true;
  }
  return false;
}

function clearEmacsYankBuffer(editor: any) {
  editor.__ivyEmacsYankBuffer = '';
}

function isNativeEmacsKillCommand(event: KeyboardEvent) {
  if (event.metaKey || event.shiftKey) return false;
  const isCtrlKillLine = event.key === 'k' && event.ctrlKey && !event.altKey;
  const isCtrlKillRegion = event.key === 'w' && event.ctrlKey && !event.altKey;
  const isAltKillWord = event.key === 'd' && event.altKey && !event.ctrlKey;
  const isAltBackwardKillWord = event.key === 'Backspace' && event.altKey && !event.ctrlKey;
  return isCtrlKillLine || isCtrlKillRegion || isAltKillWord || isAltBackwardKillWord;
}

function collapseSelectionAtPostYankCursor(editor: any) {
  setTimeout(() => {
    const cursor = typeof editor.getCursor === 'function' ? editor.getCursor() || {} : {};
    const line = Number.isFinite(cursor.line) ? cursor.line : 0;
    const ch = Number.isFinite(cursor.ch) ? cursor.ch : 0;
    if (typeof editor.setCursor === 'function') editor.setCursor(line, ch);
  }, 0);
}
