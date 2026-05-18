const EDITOR_KEYMAPS = new Set(['sublime', 'emacs', 'vim']);

export function editorName(app) {
  return (app && (app._persistedFilePath || app._persistedFileName)) || '';
}

export function editorContent(app) {
  if (app && app.cmEditor && typeof app.cmEditor.getValue === 'function') {
    return app.cmEditor.getValue();
  }
  return (app && app._persistedFileContent) || '';
}

export function editorDirty(app) {
  return editorContent(app) !== ((app && app._savedFileContent) || '');
}

export function updateEditorLabel(app, {
  doc = globalThis.document,
  updateReopenLastFileButton,
}: { doc?: Document; updateReopenLastFileButton?: () => void } = {}) {
  const name = editorName(app);
  if (app && app.uiDataModel && typeof app.uiDataModel.setFile === 'function') {
    app.uiDataModel.setFile(app._persistedFileName || '', app._persistedFilePath || app._persistedFileName || '');
  }
  const current = editorContent(app);
  const saved = (app && app._savedFileContent) || '';
  const dirty = current !== saved;
  let labelText;
  if (!name) {
    labelText = `(unsaved file)${app && app._saveInProgress ? ' [saving...]' : ''}`;
    if (dirty && !(app && app._saveInProgress)) {
      labelText = `** ${labelText}`;
    }
  } else if (app && app._saveInProgress) {
    labelText = `${name} [saving...]`;
  } else if (dirty) {
    labelText = `** ${name}`;
  } else {
    labelText = `${name} [saved]`;
  }

  const editorLabel = doc && doc.getElementById('model-editor-label');
  if (editorLabel) {
    editorLabel.textContent = labelText;
    editorLabel.title = `Editing: ${labelText}`;
  }
  const loadedFile = doc && doc.getElementById('loaded-file');
  if (loadedFile) {
    loadedFile.textContent = name;
    loadedFile.title = name;
  }
  if (typeof updateReopenLastFileButton === 'function') updateReopenLastFileButton();
  return labelText;
}

export function setEditorContent(app, content) {
  app._persistedFileContent = content;
  app._savedFileContent = content;
  if (app.cmEditor && typeof app.cmEditor.setValue === 'function') {
    const previousSuppress = app._suppressModelStateInvalidation;
    app._suppressModelStateInvalidation = true;
    try {
      app.cmEditor.setValue(content);
      resetEditorUndoHistory(app.cmEditor);
    } finally {
      app._suppressModelStateInvalidation = previousSuppress;
    }
  }
  updateEditorLabel(app, {
    updateReopenLastFileButton: () => app._updateReopenLastFileButton(),
  });
}

export function resetEditorUndoHistory(editor) {
  if (editor && typeof editor.clearHistory === 'function') {
    editor.clearHistory();
  }
}

export function refreshEditorLayout(app, win = globalThis.window) {
  const refresh = () => {
    if (app.cmEditor && typeof app.cmEditor.refresh === 'function') {
      app.cmEditor.refresh();
    }
  };

  refresh();
  if (win && typeof win.requestAnimationFrame === 'function') {
    win.requestAnimationFrame(refresh);
  }
  const setTimeoutFn = win && typeof win.setTimeout === 'function' ? win.setTimeout.bind(win) : setTimeout;
  setTimeoutFn(refresh, 0);
}

export function editorTextElement(app, doc = globalThis.document) {
  if (app.cmEditor && typeof app.cmEditor.getWrapperElement === 'function') {
    const wrapper = app.cmEditor.getWrapperElement();
    if (wrapper) return wrapper;
  }
  const editorPanel = doc && doc.getElementById('editor-panel');
  if (editorPanel) {
    return editorPanel.querySelector('.CodeMirror') || editorPanel.querySelector('#model-editor');
  }
  return doc && doc.getElementById('model-editor');
}

export function scrollEditorToLine(app, lineno) {
  if (!app.cmEditor) return;
  const line = lineno - 1;
  if (app._highlightedEditorLineHandle != null) {
    app.cmEditor.removeLineClass(app._highlightedEditorLineHandle, 'background', 'ivy-source-highlight');
  }
  app.cmEditor.setCursor(line, 0);
  app.cmEditor.setSelection(
    { line, ch: 0 },
    { line, ch: app.cmEditor.getLine(line).length },
  );
  app._highlightedEditorLineHandle = app.cmEditor.addLineClass(line, 'background', 'ivy-source-highlight');
  app._highlightedEditorLine = lineno;
  app.cmEditor.scrollIntoView({ line, ch: 0 }, 50);
  app.cmEditor.focus();
}

export function getEditorKeymap({
  doc = globalThis.document,
} = {}) {
  const checked = doc && doc.querySelector<HTMLInputElement>('input[name="keymap"]:checked');
  return checked ? checked.value : 'emacs';
}

export function setEditorKeymap(app, keymap, {
  doc = globalThis.document,
} = {}) {
  const next = EDITOR_KEYMAPS.has(keymap) ? keymap : 'emacs';
  if (app) app._editorKeymap = next;
  if (app.cmEditor && typeof app.cmEditor.setOption === 'function') {
    app.cmEditor.setOption('keyMap', next);
  }
  const radio = doc && doc.querySelector<HTMLInputElement>(`input[name="keymap"][value="${next}"]`);
  if (radio) radio.checked = true;
  return next;
}
