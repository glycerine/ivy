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
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  doc = globalThis.document,
  updateReopenLastFileButton,
} = {}) {
  const name = editorName(app);
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

  if (bridge && typeof bridge.updateEditor === 'function') {
    bridge.updateEditor({
      path: name,
      content: current,
      savedContent: saved,
      saveInProgress: !!(app && app._saveInProgress),
    });
    if (typeof updateReopenLastFileButton === 'function') updateReopenLastFileButton();
    return labelText;
  }

  const editorLabel = doc && doc.getElementById('model-editor-label');
  if (editorLabel) {
    editorLabel.textContent = labelText;
    editorLabel.title = `Editing: ${labelText}`;
  }
  if (typeof updateReopenLastFileButton === 'function') updateReopenLastFileButton();
  return labelText;
}

export function setEditorContent(app, content) {
  app._persistedFileContent = content;
  app._savedFileContent = content;
  if (app.cmEditor && typeof app.cmEditor.setValue === 'function') {
    app.cmEditor.setValue(content);
  }
  updateEditorLabel(app, {
    updateReopenLastFileButton: () => app._updateReopenLastFileButton(),
  });
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
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  doc = globalThis.document,
} = {}) {
  if (bridge && typeof bridge.getEditorKeymap === 'function') {
    return bridge.getEditorKeymap() || 'sublime';
  }
  const checked = doc && doc.querySelector('input[name="keymap"]:checked');
  return checked ? checked.value : 'sublime';
}

export function setEditorKeymap(app, keymap, {
  bridge = globalThis.window && globalThis.window.__ivyVueBridge,
  doc = globalThis.document,
} = {}) {
  const next = EDITOR_KEYMAPS.has(keymap) ? keymap : 'sublime';
  if (app.cmEditor && typeof app.cmEditor.setOption === 'function') {
    app.cmEditor.setOption('keyMap', next);
  }
  if (bridge && typeof bridge.setEditorKeymap === 'function') {
    bridge.setEditorKeymap(next);
    return next;
  }
  const radio = doc && doc.querySelector(`input[name="keymap"][value="${next}"]`);
  if (radio) radio.checked = true;
  return next;
}
