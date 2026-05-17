export function initializeCodeMirrorEditor({
  runtime = null,
  keymap = 'sublime',
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
    keyMap: keymap || 'sublime',
    tabSize: 4,
    indentUnit: 4,
    lineWrapping: false,
    matchBrackets: true,
    extraKeys: {
      'Ctrl-F': 'find',
      'Cmd-F': 'find',
      'Shift-Ctrl-G': 'findPrev',
      'Shift-Cmd-G': 'findPrev',
      'Shift-Ctrl-F': 'replace',
      'Cmd-Alt-F': 'replace',
      'Shift-Ctrl-R': 'replaceAll',
      'Shift-Cmd-Alt-F': 'replaceAll',
      'Ctrl-Z': 'undo',
      'Ctrl-Y': 'redo',
      'Ctrl-Shift-Z': 'redo',

      // trying for jump to end / beginning of buffer.
      'Ctrl-<': 'cursorStart',
      'Ctrl->': 'cursorEnd',
      'Ctrl-Shift-,': 'cursorStart', // Explicitly catch the unshifted/shifted variant
      'Ctrl-Shift-.': 'cursorEnd',
      'Alt-<': 'cursorStart',
      'Alt->': 'cursorEnd',
      'Alt-Shift-,': 'cursorStart',
      'Alt-Shift-.': 'cursorEnd',
      'Esc-<': 'cursorStart',
      'Esc->': 'cursorEnd'
    },
  });
  modelEditor.__ivyCodeMirrorEditor = editor;

  installEscapeEndChord({
    editor,
    codeMirror,
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

function installEscapeEndChord({
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
    const isGreaterThan = event.key === '>' || (event.key === '.' && event.shiftKey);
    if (isGreaterThan && lastEscapeAt > 0 && Date.now() - lastEscapeAt <= chordWindowMs) {
      lastEscapeAt = 0;
      event.preventDefault();
      event.stopPropagation();
      codeMirror.commands.cursorEnd(editor);
      return;
    }
    if (!event.altKey && !event.ctrlKey && !event.metaKey && event.key !== 'Shift') {
      lastEscapeAt = 0;
    }
  }, true);
}
