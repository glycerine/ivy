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
      //'Ctrl-F': 'find',
      //'Cmd-F': 'find',
      'Ctrl-X': (cm: any) => armEmacsSaveKey(cm, codeMirror, runtime),
      'Ctrl-S': (cm: any) => runEmacsSearchKey(cm, codeMirror, 'Ctrl-S', runtime),
      'Ctrl-R': (cm: any) => runEmacsSearchKey(cm, codeMirror, 'Ctrl-R'),
      'Ctrl-W': (cm: any) => runEmacsSaveAsKey(cm, codeMirror, runtime),
      //'Shift-Ctrl-F': 'replace',
      //'Cmd-Alt-F': 'replace',
      //'Shift-Ctrl-R': 'replaceAll',
      //'Shift-Cmd-Alt-F': 'replaceAll',
      'Ctrl-G': (cm: any) => {
        if (cm.__ivyEmacsISearch && typeof cm.__ivyEmacsISearch.abort === 'function') {
          cm.__ivyEmacsISearch.abort();
          return true;
        }
        return codeMirror.Pass;
      },
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
  installEmacsReplacePromptKeys({
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
      if (typeof runtime._invalidateModelState === 'function') {
        runtime._invalidateModelState('editor-change');
      }
      if (typeof runtime._syncAnalysisSpreadsheetFromEditor === 'function') {
        runtime._syncAnalysisSpreadsheetFromEditor();
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
    if (findReplaceDialog(editor, doc)) {
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
    const isReplace = (event.key === '%' || (event.key === '5' && event.shiftKey))
      && !event.altKey
      && !event.ctrlKey
      && !event.metaKey
      && isEmacsKeymap(editor);
    const isCopySelection = event.key === 'w'
      && !event.altKey
      && !event.ctrlKey
      && !event.metaKey
      && !event.shiftKey
      && isEmacsKeymap(editor);
    if ((isLessThan || isGreaterThan || isPageUp || isReplace || isCopySelection) && inChordWindow) {
      lastEscapeAt = 0;
      event.preventDefault();
      event.stopPropagation();
      if (isReplace) {
        runEditorCommand(editor, codeMirror, 'replace');
        return;
      }
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

function installEmacsReplacePromptKeys({
  editor,
  doc,
}: {
  editor: any;
  doc: Document;
}) {
  if (!doc || !editor) return;
  doc.addEventListener('input', (event) => {
    rememberReplaceDialogInput(editor, event.target);
  }, true);
  doc.addEventListener('keydown', (event) => {
    rememberReplaceDialogInput(editor, event.target);
    if (!isEmacsKeymap(editor)) return;
    if (isCtrlG(event)) {
      const replaceDialog = findReplaceDialog(editor, doc);
      if (!replaceDialog) return;
      event.preventDefault();
      event.stopPropagation();
      const questionDialog = findReplaceQuestionDialog(editor, doc);
      if (questionDialog) clickReplacePromptButton(questionDialog, 'Stop');
      else abortReplaceDialog(editor, replaceDialog);
      return;
    }
    const dialog = findReplaceQuestionDialog(editor, doc);
    if (!dialog) return;
    const action = replacePromptAction(event);
    if (!action) return;

    event.preventDefault();
    event.stopPropagation();
    const replaceRange = currentEditorSelectionRange(editor);
    scrollCurrentReplaceMatchIntoView(editor);
    if (action === 'yes-stop') {
      clickReplacePromptButton(dialog, 'Yes');
      setTimeout(() => {
        const nextDialog = findReplaceQuestionDialog(editor, doc);
        if (nextDialog) clickReplacePromptButton(nextDialog, 'Stop');
        if (replaceRange) collapseSelectionAtPosition(editor, replaceRange.to);
      }, 0);
      return;
    }
    if (action === 'rest') {
      replaceRestToEnd(editor, doc, dialog);
      return;
    }
    const stopAtEnd = (action === 'Yes' || action === 'No') && isLastReplaceMatch(editor);
    if (stopAtEnd && action === 'No') {
      clickReplacePromptButton(dialog, 'Stop');
      return;
    }
    clickReplacePromptButton(dialog, action);
    if (action === 'Yes' || action === 'No') {
      setTimeout(() => {
        const nextDialog = findReplaceQuestionDialog(editor, doc);
        if (stopAtEnd) {
          if (nextDialog) clickReplacePromptButton(nextDialog, 'Stop');
          if (action === 'Yes' && replaceRange) collapseSelectionAtPosition(editor, replaceRange.to);
        } else if (nextDialog) {
          scrollCurrentReplaceMatchIntoView(editor);
        }
      }, 0);
    }
  }, true);
}

function isCtrlG(event: KeyboardEvent) {
  return event.key === 'g'
    && event.ctrlKey
    && !event.altKey
    && !event.metaKey
    && !event.shiftKey;
}

function findReplaceDialog(editor: any, doc: Document): HTMLElement | null {
  const questionDialog = findReplaceQuestionDialog(editor, doc);
  if (questionDialog) return questionDialog;
  const wrapper = typeof editor.getWrapperElement === 'function' ? editor.getWrapperElement() : null;
  const roots = [
    wrapper,
    wrapper && wrapper.parentElement,
    doc,
  ].filter(Boolean) as Array<HTMLElement | Document>;
  for (const root of roots) {
    const dialogs = Array.from(root.querySelectorAll('.CodeMirror-dialog')) as HTMLElement[];
    const dialog = dialogs.find((candidate) => {
      const text = (candidate.textContent || '').replace(/\s+/g, ' ').trim();
      return text.startsWith('Replace:')
        || text.startsWith('With:')
        || text.startsWith('Replace with:');
    });
    if (dialog) return dialog;
  }
  return null;
}

function abortReplaceDialog(editor: any, dialog: HTMLElement) {
  if (dialog && dialog.parentNode) dialog.parentNode.removeChild(dialog);
  editor.__ivyEmacsReplaceQueryText = '';
  editor.__ivyEmacsReplaceText = null;
  if (typeof editor.focus === 'function') editor.focus();
}

function findReplaceQuestionDialog(editor: any, doc: Document): HTMLElement | null {
  const wrapper = typeof editor.getWrapperElement === 'function' ? editor.getWrapperElement() : null;
  const roots = [
    wrapper,
    wrapper && wrapper.parentElement,
    doc,
  ].filter(Boolean) as Array<HTMLElement | Document>;
  for (const root of roots) {
    const dialogs = Array.from(root.querySelectorAll('.CodeMirror-dialog')) as HTMLElement[];
    const dialog = dialogs.find((candidate) => {
      const text = (candidate.textContent || '').replace(/\s+/g, ' ');
      return text.includes('Replace?')
        && !!replacePromptButton(candidate, 'Yes')
        && !!replacePromptButton(candidate, 'No')
        && !!replacePromptButton(candidate, 'All')
        && !!replacePromptButton(candidate, 'Stop');
    });
    if (dialog) return dialog;
  }
  return null;
}

function replacePromptAction(event: KeyboardEvent) {
  if (event.ctrlKey || event.altKey || event.metaKey) return '';
  if (event.key === 'y' || event.key === 'Y' || event.key === ' ') return 'Yes';
  if (event.key === 'n' || event.key === 'N' || event.key === 'Backspace' || event.key === 'Delete') return 'No';
  if (event.key === '!') return 'rest';
  if (event.key === 'q' || event.key === 'Q' || event.key === 'Enter') return 'Stop';
  if (event.key === '.') return 'yes-stop';
  return '';
}

function replacePromptButton(dialog: HTMLElement, label: string): HTMLButtonElement | null {
  const buttons = Array.from(dialog.querySelectorAll('button')) as HTMLButtonElement[];
  return buttons.find((button) => (button.textContent || '').trim() === label) || null;
}

function clickReplacePromptButton(dialog: HTMLElement, label: string) {
  const button = replacePromptButton(dialog, label);
  if (button) button.click();
}

function replaceRestToEnd(editor: any, doc: Document, initialDialog: HTMLElement) {
  if (!canBulkReplaceToEOF(editor)) {
    clickReplacePromptButton(initialDialog, 'All');
    return;
  }
  const query = currentReplaceQuery(editor);
  const replacement = currentReplaceText(editor);
  const range = currentEditorSelectionRange(editor);
  const start = range.from;
  let lastReplacementEnd = range.to;
  const replaceAll = () => {
    const cursor = editor.getSearchCursor(query, start, searchCursorOptionsForQuery(query));
    while (cursor && typeof cursor.findNext === 'function') {
      const match = cursor.findNext();
      if (!match) break;
      const replacementText = replacementForSearchMatch(query, replacement, match);
      if (typeof cursor.replace !== 'function') break;
      cursor.replace(replacementText);
      if (typeof cursor.to === 'function') {
        lastReplacementEnd = cursor.to();
      }
    }
  };
  if (typeof editor.operation === 'function') editor.operation(replaceAll);
  else replaceAll();
  const nextDialog = findReplaceQuestionDialog(editor, doc);
  clickReplacePromptButton(nextDialog || initialDialog, 'Stop');
  collapseSelectionAtPosition(editor, lastReplacementEnd);
  scrollPositionIntoView(editor, lastReplacementEnd);
}

function canBulkReplaceToEOF(editor: any) {
  return typeof editor.getSearchCursor === 'function'
    && !!currentReplaceQuery(editor)
    && currentReplaceText(editor) !== null
    && !!currentEditorSelectionRange(editor);
}

function isLastReplaceMatch(editor: any) {
  if (typeof editor.getSearchCursor !== 'function') return false;
  const query = currentReplaceQuery(editor);
  const range = currentEditorSelectionRange(editor);
  if (!query || !range) return false;
  const options = typeof query === 'string' ? { caseFold: isLowerCaseSearchQuery(query) } : undefined;
  const cursor = editor.getSearchCursor(query, range.to, options);
  return !(cursor && typeof cursor.findNext === 'function' && cursor.findNext());
}

function currentReplaceQuery(editor: any) {
  const state = editor && editor.state && editor.state.search;
  const stateQuery = state && (state.query || state.lastQuery);
  if (stateQuery) return stateQuery;
  if (typeof editor.getSelection === 'function') {
    const selection = editor.getSelection();
    if (typeof selection === 'string' && selection.length > 0) return selection;
  }
  return null;
}

function currentReplaceText(editor: any): string | null {
  return typeof editor.__ivyEmacsReplaceText === 'string' ? editor.__ivyEmacsReplaceText : null;
}

function rememberReplaceDialogInput(editor: any, target: EventTarget | null) {
  const input = target as HTMLInputElement | null;
  if (!input || input.tagName !== 'INPUT') return;
  const dialog = input.closest('.CodeMirror-dialog') as HTMLElement | null;
  if (!dialog) return;
  const label = (dialog.textContent || '').replace(/\s+/g, ' ').trim();
  if (label.startsWith('Replace:')) {
    editor.__ivyEmacsReplaceQueryText = input.value;
  } else if (label.startsWith('With:') || label.startsWith('Replace with:')) {
    editor.__ivyEmacsReplaceText = parseCodeMirrorReplaceString(input.value);
  }
}

function parseCodeMirrorReplaceString(value: string) {
  return String(value || '').replace(/\\([nrt\\])/g, (_match, ch) => {
    if (ch === 'n') return '\n';
    if (ch === 'r') return '\r';
    if (ch === 't') return '\t';
    return ch;
  });
}

function searchCursorOptionsForQuery(query: any) {
  return typeof query === 'string' ? { caseFold: isLowerCaseSearchQuery(query) } : undefined;
}

function replacementForSearchMatch(query: any, replacement: string, match: any) {
  if (typeof query === 'string' || !match) return replacement;
  return replacement.replace(/\$(\d+)/g, (_whole, index) => match[Number(index)] || '');
}

function currentEditorSelectionRange(editor: any) {
  if (typeof editor.getCursor !== 'function') return null;
  const from = editor.getCursor('from');
  const to = editor.getCursor('to');
  if (!from || !to) return null;
  if (from.line === to.line && from.ch === to.ch) return null;
  return { from, to };
}

function scrollCurrentReplaceMatchIntoView(editor: any) {
  const range = currentEditorSelectionRange(editor);
  if (!range) return false;
  return scrollRangeIntoView(editor, range);
}

function scrollRangeIntoView(editor: any, range: { from: any; to: any }) {
  if (typeof editor.charCoords === 'function' && typeof editor.scrollTo === 'function') {
    const from = editor.charCoords(range.from, 'local');
    const to = editor.charCoords(range.to, 'local');
    const top = Math.min(Number(from && from.top) || 0, Number(to && to.top) || 0);
    const bottom = Math.max(Number(from && from.bottom) || top, Number(to && to.bottom) || top);
    const scroller = typeof editor.getScrollerElement === 'function' ? editor.getScrollerElement() : null;
    const height = Number(scroller && scroller.clientHeight) || 0;
    editor.scrollTo(null, Math.max(0, Math.round((top + bottom) / 2 - height / 2)));
    return true;
  }
  if (typeof editor.scrollIntoView === 'function') {
    editor.scrollIntoView({ from: range.from, to: range.to }, 120);
    return true;
  }
  return false;
}

function scrollPositionIntoView(editor: any, position: any) {
  if (!position) return false;
  return scrollRangeIntoView(editor, { from: position, to: position });
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

function runEmacsSaveAsKey(editor: any, codeMirror: any, runtime: any = null) {
  if (!isEmacsKeymap(editor)) {
    return codeMirror && codeMirror.Pass;
  }
  if (runArmedEmacsFileCommand(editor, runtime, 'saveAs', 'Save as failed')) {
    return true;
  }
  return codeMirror && codeMirror.Pass;
}

function runArmedEmacsSaveKey(editor: any, runtime: any) {
  return runArmedEmacsFileCommand(editor, runtime, 'save', 'Save failed');
}

function runArmedEmacsFileCommand(editor: any, runtime: any, method: 'save' | 'saveAs', errorPrefix: string) {
  const armedAt = Number(editor.__ivyEmacsSaveChordAt) || 0;
  editor.__ivyEmacsSaveChordAt = 0;
  if (!armedAt || Date.now() - armedAt > 2000 || !runtime || typeof runtime[method] !== 'function') {
    return false;
  }
  const result = runtime[method]();
  if (result && typeof result.catch === 'function') {
    result.catch((err: any) => {
      if (runtime.controls && typeof runtime.controls.setStatus === 'function') {
        runtime.controls.setStatus(`${errorPrefix}: ${err && err.message ? err.message : String(err)}`, 'error');
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
  removeIsearchDialogs(editor);

  const originCursor = getEditorCursor(editor);
  const originSelections = typeof editor.listSelections === 'function'
    ? editor.listSelections()
    : null;
  let currentFrom: any = null;
  let currentTo: any = null;
  let currentSearchMark: any = null;
  let currentDirection = direction;
  let query = '';
  let closed = false;
  let closeDialog: any = null;
  let removeMousedownListener: (() => void) | null = null;

  const closeSearch = ({ restore = false } = {}) => {
    if (closed) return;
    closed = true;
    delete editor.__ivyEmacsISearch;
    clearSearchMark(currentSearchMark);
    currentSearchMark = null;
    if (!restore && currentFrom && currentTo) {
      collapseSelectionAtPosition(editor, currentDirection === 'backward' ? currentFrom : currentTo);
    }
    if (restore) restoreEditorSelections(editor, originSelections, originCursor);
    if (removeMousedownListener) { removeMousedownListener(); removeMousedownListener = null; }
    if (typeof closeDialog === 'function') closeDialog();
    removeIsearchDialogs(editor);
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
    currentDirection = searchDirection;
    currentSearchMark = selectSearchMatch(editor, match, currentSearchMark);
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
      else {
        clearSearchMark(currentSearchMark);
        currentSearchMark = null;
        restoreEditorSelections(editor, originSelections, originCursor);
      }
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

  const ownerDoc = (typeof editor.getWrapperElement === 'function'
    ? editor.getWrapperElement()?.ownerDocument
    : null) || (globalThis as any).document;
  if (ownerDoc) {
    const onDocMousedown = (event: MouseEvent) => {
      const dialogs = findIsearchDialogs(editor);
      const clickedInsideDialog = dialogs.some(d => d.contains(event.target as Node));
      if (!clickedInsideDialog) {
        acceptSearch();
      }
    };
    ownerDoc.addEventListener('mousedown', onDocMousedown);
    removeMousedownListener = () => ownerDoc.removeEventListener('mousedown', onDocMousedown);
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
  removeNonInputIsearchDialogs(editor);
  const wrapper = typeof editor.getWrapperElement === 'function' ? editor.getWrapperElement() : null;
  return (wrapper && wrapper.querySelector('.CodeMirror-dialog input'))
    || (wrapper && wrapper.parentElement && wrapper.parentElement.querySelector('.CodeMirror-dialog input'))
    || null;
}

function removeIsearchDialogs(editor: any) {
  for (const dialog of findIsearchDialogs(editor)) {
    if (dialog.parentNode) dialog.parentNode.removeChild(dialog);
  }
}

function removeNonInputIsearchDialogs(editor: any) {
  for (const dialog of findIsearchDialogs(editor)) {
    if (!dialog.querySelector('input') && dialog.parentNode) dialog.parentNode.removeChild(dialog);
  }
}

function findIsearchDialogs(editor: any): HTMLElement[] {
  const wrapper = typeof editor.getWrapperElement === 'function' ? editor.getWrapperElement() : null;
  const roots = [
    wrapper,
    wrapper && wrapper.parentElement,
  ].filter(Boolean) as HTMLElement[];
  const dialogs: HTMLElement[] = [];
  for (const root of roots) {
    for (const dialog of Array.from(root.querySelectorAll('.CodeMirror-dialog')) as HTMLElement[]) {
      const text = (dialog.textContent || '').replace(/\s+/g, ' ').trim();
      const hasIsearchLabel = text.startsWith('I-search:') || text.startsWith('I-search backward:');
      const hasIsearchInput = !!dialog.querySelector('input.CodeMirror-search-field');
      if ((hasIsearchLabel || hasIsearchInput) && !dialogs.includes(dialog)) {
        dialogs.push(dialog);
      }
    }
  }
  return dialogs;
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

function selectSearchMatch(editor: any, match: { from: any; to: any }, previousMark: any = null) {
  clearSearchMark(previousMark);
  let nextMark = null;
  if (typeof editor.markText === 'function') {
    nextMark = editor.markText(match.from, match.to, { className: 'ivy-emacs-isearch-match' });
  }
  if (typeof editor.setSelection === 'function') {
    editor.setSelection(match.from, match.to);
  } else if (typeof editor.setCursor === 'function') {
    editor.setCursor(match.to.line, match.to.ch);
  }
  if (typeof editor.scrollIntoView === 'function') {
    editor.scrollIntoView({ from: match.from, to: match.to }, 80);
  }
  return nextMark;
}

function clearSearchMark(mark: any) {
  if (mark && typeof mark.clear === 'function') mark.clear();
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

function collapseSelectionAtPosition(editor: any, position: any) {
  const line = Number.isFinite(position && position.line) ? position.line : 0;
  const ch = Number.isFinite(position && position.ch) ? position.ch : 0;
  if (typeof editor.setCursor === 'function') {
    editor.setCursor(line, ch);
  } else if (typeof editor.setSelection === 'function') {
    editor.setSelection({ line, ch }, { line, ch });
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
