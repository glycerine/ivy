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
      'Ctrl-X': (cm: any) => armEmacsSaveKey(cm, codeMirror, runtime),
      'Ctrl-S': (cm: any) => runEmacsSearchKey(cm, codeMirror, 'Ctrl-S', runtime),
      'Ctrl-R': (cm: any) => runEmacsSearchKey(cm, codeMirror, 'Ctrl-R'),
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
  const target = { line, ch };
  const extending = editorSelectionIsExtending(editor);
  if (extending && typeof editor.setSelection === 'function') {
    editor.setSelection(editorSelectionAnchor(editor), target);
    if (typeof editor.setExtending === 'function') editor.setExtending(true);
  } else if (typeof editor.setCursor === 'function') {
    editor.setCursor(line, ch);
  } else if (typeof editor.setSelection === 'function') {
    editor.setSelection(target, target);
  }
  if (typeof editor.scrollIntoView === 'function') editor.scrollIntoView({ line, ch }, 50);
  if (typeof editor.focus === 'function') editor.focus();
}

function editorSelectionIsExtending(editor: any) {
  return typeof editor.getExtending === 'function' && !!editor.getExtending();
}

function editorSelectionAnchor(editor: any) {
  if (typeof editor.getCursor === 'function') {
    const anchor = editor.getCursor('anchor');
    if (anchor && Number.isFinite(anchor.line) && Number.isFinite(anchor.ch)) return anchor;
  }
  if (typeof editor.listSelections === 'function') {
    const selections = editor.listSelections();
    const anchor = selections && selections[0] && selections[0].anchor;
    if (anchor && Number.isFinite(anchor.line) && Number.isFinite(anchor.ch)) return anchor;
  }
  return getEditorCursor(editor);
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
        if (copySelectionToEmacsYankBuffer(editor)) {
          deactivateEmacsMarkAtCursor(editor);
        }
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

function armEmacsSaveKey(editor: any, codeMirror: any, runtime: any) {
  if (!isEmacsKeymap(editor) || !runtime || typeof runtime.save !== 'function') {
    return codeMirror && codeMirror.Pass;
  }
  editor.__ivyEmacsSaveChordAt = Date.now();
  return true;
}

function runEmacsSearchKey(editor: any, codeMirror: any, keyName: 'Ctrl-S' | 'Ctrl-R', runtime: any = null) {
  if (!isEmacsKeymap(editor)) {
    return codeMirror && codeMirror.Pass;
  }
  if (keyName === 'Ctrl-S' && runArmedEmacsSaveKey(editor, runtime)) {
    return true;
  }
  return startIvyEmacsISearch(editor, codeMirror, keyName === 'Ctrl-R' ? 'backward' : 'forward');
}

function runArmedEmacsSaveKey(editor: any, runtime: any) {
  const armedAt = Number(editor.__ivyEmacsSaveChordAt) || 0;
  editor.__ivyEmacsSaveChordAt = 0;
  if (!armedAt || Date.now() - armedAt > 2000 || !runtime || typeof runtime.save !== 'function') {
    return false;
  }
  const result = runtime.save();
  if (result && typeof result.catch === 'function') {
    result.catch((err: any) => {
      if (runtime.controls && typeof runtime.controls.setStatus === 'function') {
        runtime.controls.setStatus(`Save failed: ${err && err.message ? err.message : String(err)}`, 'error');
      }
    });
  }
  return true;
}

function startIvyEmacsISearch(editor: any, codeMirror: any, direction: 'forward' | 'backward') {
  if (editor.__ivyEmacsISearch && typeof editor.__ivyEmacsISearch.repeat === 'function') {
    editor.__ivyEmacsISearch.repeat(direction);
    return true;
  }
  if (typeof editor.openDialog !== 'function' || typeof editor.getSearchCursor !== 'function') {
    return codeMirror && codeMirror.Pass;
  }

  const originCursor = getEditorCursor(editor);
  const originSelections = typeof editor.listSelections === 'function'
    ? editor.listSelections()
    : null;
  let currentFrom: any = null;
  let currentTo: any = null;
  let query = '';
  let closed = false;
  let closeDialog: any = null;

  const closeSearch = ({ restore = false } = {}) => {
    if (closed) return;
    closed = true;
    delete editor.__ivyEmacsISearch;
    if (restore) restoreEditorSelections(editor, originSelections, originCursor);
    if (typeof closeDialog === 'function') closeDialog();
    if (typeof editor.focus === 'function' && !restore) editor.focus();
  };

  const runSearch = (searchDirection: 'forward' | 'backward', repeat = false) => {
    const text = readIsearchInput(editor) || query;
    if (!text) return false;
    query = text;
    editor.__ivyLastISearchQuery = text;
    const start = repeat && currentFrom && currentTo
      ? (searchDirection === 'forward' ? currentTo : currentFrom)
      : originCursor;
    const match = findSearchMatch(editor, text, start, searchDirection);
    if (!match) return false;
    currentFrom = match.from;
    currentTo = match.to;
    selectSearchMatch(editor, match);
    return true;
  };

  const repeatSearch = (searchDirection: 'forward' | 'backward') => {
    if (!query && editor.__ivyLastISearchQuery) {
      query = editor.__ivyLastISearchQuery;
      writeIsearchInput(editor, query);
    }
    runSearch(searchDirection, true);
  };

  const acceptSearch = () => {
    if (query) editor.__ivyLastISearchQuery = query;
    closeSearch();
  };

  const abortSearch = () => {
    closeSearch({ restore: true });
  };

  editor.__ivyEmacsISearch = {
    repeat: repeatSearch,
    accept: acceptSearch,
    abort: abortSearch,
  };

  const prompt = direction === 'forward' ? 'I-search:' : 'I-search backward:';
  closeDialog = editor.openDialog(
    `<span class="CodeMirror-search-label">${prompt}</span> <input type="text" class="CodeMirror-search-field" autocomplete="off" />`,
    acceptSearch,
    {
      bottom: true,
      closeOnBlur: false,
      closeOnEnter: false,
      value: '',
    },
  );

  const input = findIsearchInput(editor);
  if (input) {
    input.addEventListener('input', () => {
      query = input.value;
      if (query) runSearch(direction, false);
      else restoreEditorSelections(editor, originSelections, originCursor);
    });
    input.addEventListener('keydown', (event) => {
      const isForwardRepeat = event.key === 's' && event.ctrlKey && !event.altKey && !event.metaKey;
      const isBackwardRepeat = event.key === 'r' && event.ctrlKey && !event.altKey && !event.metaKey;
      const isAbort = event.key === 'Escape'
        || (event.key === 'g' && event.ctrlKey && !event.altKey && !event.metaKey);
      const arrowCommand = editorArrowCommand(event.key);
      if (isForwardRepeat || isBackwardRepeat) {
        event.preventDefault();
        event.stopPropagation();
        repeatSearch(isBackwardRepeat ? 'backward' : 'forward');
      } else if (event.key === 'Enter') {
        event.preventDefault();
        event.stopPropagation();
        acceptSearch();
      } else if (arrowCommand) {
        event.preventDefault();
        event.stopPropagation();
        acceptSearch();
        runEditorCommand(editor, codeMirror, arrowCommand);
      } else if (isAbort) {
        event.preventDefault();
        event.stopPropagation();
        abortSearch();
      }
    }, true);
    input.focus();
  }

  return true;
}

function editorArrowCommand(key: string) {
  switch (key) {
    case 'ArrowLeft':
      return 'goCharLeft';
    case 'ArrowRight':
      return 'goCharRight';
    case 'ArrowUp':
      return 'goLineUp';
    case 'ArrowDown':
      return 'goLineDown';
    default:
      return '';
  }
}

function runEditorCommand(editor: any, codeMirror: any, command: string) {
  if (typeof editor.execCommand === 'function') {
    editor.execCommand(command);
    return true;
  }
  if (codeMirror && codeMirror.commands && typeof codeMirror.commands[command] === 'function') {
    codeMirror.commands[command](editor);
    return true;
  }
  return false;
}

function readIsearchInput(editor: any) {
  const input = findIsearchInput(editor);
  return input ? input.value : '';
}

function writeIsearchInput(editor: any, value: string) {
  const input = findIsearchInput(editor);
  if (input) input.value = value;
}

function findIsearchInput(editor: any): HTMLInputElement | null {
  const wrapper = typeof editor.getWrapperElement === 'function' ? editor.getWrapperElement() : null;
  return (wrapper && wrapper.querySelector('.CodeMirror-dialog input'))
    || (wrapper && wrapper.parentElement && wrapper.parentElement.querySelector('.CodeMirror-dialog input'))
    || null;
}

function findSearchMatch(editor: any, query: string, start: any, direction: 'forward' | 'backward') {
  const caseFold = isLowerCaseSearchQuery(query);
  const cursor = editor.getSearchCursor(query, start, caseFold);
  const found = direction === 'forward'
    ? cursor.findNext()
    : cursor.findPrevious();
  if (found) return { from: cursor.from(), to: cursor.to() };

  const wrapStart = direction === 'forward'
    ? { line: typeof editor.firstLine === 'function' ? editor.firstLine() : 0, ch: 0 }
    : editorDocumentEnd(editor);
  const wrapCursor = editor.getSearchCursor(query, wrapStart, caseFold);
  const wrapFound = direction === 'forward'
    ? wrapCursor.findNext()
    : wrapCursor.findPrevious();
  if (wrapFound) return { from: wrapCursor.from(), to: wrapCursor.to() };
  return null;
}

function isLowerCaseSearchQuery(query: string) {
  return query === query.toLowerCase();
}

function selectSearchMatch(editor: any, match: { from: any; to: any }) {
  if (typeof editor.setSelection === 'function') {
    editor.setSelection(match.from, match.to);
  } else if (typeof editor.setCursor === 'function') {
    editor.setCursor(match.to.line, match.to.ch);
  }
  if (typeof editor.scrollIntoView === 'function') {
    editor.scrollIntoView({ from: match.from, to: match.to }, 80);
  }
}

function restoreEditorSelections(editor: any, selections: any, cursor: any) {
  if (selections && typeof editor.setSelections === 'function') {
    editor.setSelections(selections);
  } else if (selections && selections[0] && typeof editor.setSelection === 'function') {
    editor.setSelection(selections[0].anchor, selections[0].head);
  } else if (typeof editor.setCursor === 'function') {
    editor.setCursor(cursor.line, cursor.ch);
  }
}

function getEditorCursor(editor: any) {
  const cursor = typeof editor.getCursor === 'function' ? editor.getCursor() || {} : {};
  return {
    line: Number.isFinite(cursor.line) ? cursor.line : 0,
    ch: Number.isFinite(cursor.ch) ? cursor.ch : 0,
  };
}

function editorDocumentEnd(editor: any) {
  const lastLine = typeof editor.lastLine === 'function'
    ? editor.lastLine()
    : Math.max(0, (typeof editor.lineCount === 'function' ? editor.lineCount() : 1) - 1);
  const lineText = typeof editor.getLine === 'function' ? editor.getLine(lastLine) || '' : '';
  return { line: lastLine, ch: lineText.length };
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
    collapseSelectionAtCursor(editor);
  }, 0);
}

function collapseSelectionAtCursor(editor: any) {
  const cursor = typeof editor.getCursor === 'function' ? editor.getCursor() || {} : {};
  const line = Number.isFinite(cursor.line) ? cursor.line : 0;
  const ch = Number.isFinite(cursor.ch) ? cursor.ch : 0;
  if (typeof editor.setCursor === 'function') editor.setCursor(line, ch);
}

function deactivateEmacsMarkAtCursor(editor: any) {
  if (typeof editor.setExtending === 'function') editor.setExtending(false);
  collapseSelectionAtCursor(editor);
}
