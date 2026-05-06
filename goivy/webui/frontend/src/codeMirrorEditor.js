export function initializeLegacyCodeMirror({
  legacyApp,
  editorStore,
  doc = globalThis.document,
  codeMirror = globalThis.CodeMirror,
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
    keyMap: editorStore && editorStore.keymap ? editorStore.keymap : 'sublime',
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

  if (editor && typeof editor.on === 'function' && legacyApp) {
    editor.on('change', () => {
      if (!legacyApp.cmEditor || typeof legacyApp.cmEditor.getValue !== 'function') return;
      legacyApp._persistedFileContent = legacyApp.cmEditor.getValue();
      if (typeof legacyApp._updateEditorLabel === 'function') {
        legacyApp._updateEditorLabel();
      }
    });
  }

  return editor;
}
