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

  const editor = codeMirror.fromTextArea(modelEditor, {
    lineNumbers: true,
    keyMap: keymap || 'sublime',
    tabSize: 4,
    indentUnit: 4,
    lineWrapping: false,
    matchBrackets: true,
    extraKeys: {
      'Ctrl-Z': 'undo',
      'Ctrl-Y': 'redo',
      'Ctrl-Shift-Z': 'redo',
    },
  });
  modelEditor.__ivyCodeMirrorEditor = editor;

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
