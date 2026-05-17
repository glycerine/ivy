import { connectSessionEvents, createSession } from './sessionService.ts';
import { applyArgSnapshot, applyConceptSnapshot } from './uiDataRenderService.ts';
import { resetEditorUndoHistory } from './editorService.ts';

export function rememberLastOpenFile(app, persist) {
  const name = app._persistedFileName || (app._fileHandle && app._fileHandle.name) || '';
  if (!name) return;
  app._lastClosedFileHandle = app._fileHandle;
  app._lastClosedSessionId = persist.getSessionIdFromURL() || (app.api && app.api.sessionId) || '';
  app._lastClosedFileName = name || 'file';
}

export function updateReopenLastFileButton(app, {
  doc = globalThis.document,
} = {}) {
  const btn = doc && doc.getElementById('file-reopen-last');
  const noCurrentFile = !app._fileHandle && !app._persistedFileName;
  const visible = !!(noCurrentFile && (app._lastClosedFileHandle || app._lastClosedSessionId));
  const label = `Re-open last file ${app._lastClosedFileName || 'file'}`;
  if (!btn) return;
  if (visible) {
    btn.textContent = label;
    btn.style.display = '';
  } else {
    btn.style.display = 'none';
  }
}

export async function reopenLastFile(app, persist) {
  if (!app._lastClosedFileHandle && !app._lastClosedSessionId) return;
  try {
    if (app._lastClosedFileHandle) {
      const file = await app._lastClosedFileHandle.getFile();
      app._fileHandle = app._lastClosedFileHandle;
      await app.loadFile(file);
    } else {
      await app.loadRecentSession(app._lastClosedSessionId);
    }
    persist.setFileName(
      app._persistedFileName || app._lastClosedFileName,
      app._persistedFilePath || app._persistedFileName || app._lastClosedFileName,
    );
  } catch (err) {
    app.controls.setStatus(`Re-open failed: ${err.message}`, 'error');
  }
}

export async function ensureFileHandleWritable(app) {
  if (!app._fileHandle) return false;
  if (!app._fileHandle.queryPermission || !app._fileHandle.requestPermission) {
    return true;
  }
  const opts = { mode: 'readwrite' };
  let perm = await app._fileHandle.queryPermission(opts);
  if (perm === 'granted') return true;
  perm = await app._fileHandle.requestPermission(opts);
  return perm === 'granted';
}

export async function readFileHandleContent(app) {
  if (!app._fileHandle) return null;
  const file = await app._fileHandle.getFile();
  return file.text();
}

export async function restoreFileHandleForCurrentFile(app, persist) {
  if (app._fileHandle) return true;
  if (!app._persistedFileName) return false;
  const state = {
    sessionId: persist.getSessionIdFromURL() || (app.api && app.api.sessionId) || '',
    fileName: app._persistedFileName,
    filePath: app._persistedFilePath || app._persistedFileName,
  };
  const handle = await persist.loadFileHandle(state);
  if (!handle) return false;
  app._fileHandle = handle;
  await persist.saveFileHandle(app);
  return true;
}

export function mergeDiskVersionIntoEditBuffer(baseContent, editorContent, diskContent) {
  if (editorContent === baseContent) return diskContent;
  if (diskContent === baseContent) return editorContent;
  return [
    '<<<<<<< EDIT BUFFER',
    editorContent.replace(/\s*$/, ''),
    '||||||| LAST SAVED',
    baseContent.replace(/\s*$/, ''),
    '=======',
    diskContent.replace(/\s*$/, ''),
    '>>>>>>> ON DISK',
    '',
  ].join('\n');
}

export async function confirmNoExternalChangeBeforeSave(app, content, persist) {
  if (!app._fileHandle) return 'ok';
  const diskContent = await app._readFileHandleContent();
  const lastSaved = app._savedFileContent || '';
  if (diskContent === lastSaved || diskContent === content) {
    return 'ok';
  }
  const choice = await app.showExternalChangeDialog();
  if (choice === 'overwrite') {
    return 'overwrite';
  }
  if (choice === 'reload') {
    app.setEditorContent(diskContent);
    app._persistedFileContent = diskContent;
    app._savedFileContent = diskContent;
    persist.save(app);
    app.controls.setStatus(`Reverted to on-disk version: ${app._persistedFileName || 'model'}`, 'success');
  } else if (choice === 'merge') {
    const merged = mergeDiskVersionIntoEditBuffer(lastSaved, content, diskContent);
    app._persistedFileContent = merged;
    app._savedFileContent = diskContent;
    if (app.cmEditor) {
      app.cmEditor.setValue(merged);
      resetEditorUndoHistory(app.cmEditor);
    }
    app._updateEditorLabel();
    persist.save(app);
    app.controls.setStatus('Merged disk changes into editor buffer; resolve conflict markers before saving', 'warning');
  } else {
    app.controls.setStatus('Save cancelled: file changed on disk', 'warning');
  }
  return 'skip';
}

export async function chooseAndLoadModelFile(app, {
  win = globalThis.window,
  doc = globalThis.document,
} = {}) {
  const fileInput = doc && doc.getElementById('file-input');
  if (win && win.showOpenFilePicker) {
    try {
      const handles = await win.showOpenFilePicker({
        types: [{ description: 'Ivy files', accept: { 'text/plain': ['.ivy'] } }],
        multiple: false,
      });
      const handle = handles[0];
      const file = await handle.getFile();
      app._fileHandle = handle;
      await app.loadFile(file);
    } catch (err) {
      if (err.name !== 'AbortError') {
        app.controls.setStatus(`Load failed: ${err.message}`, 'error');
      }
    }
  } else if (fileInput) {
    fileInput.click();
  }
}

export function readBrowserFileText(file, win = globalThis.window) {
  return new Promise((resolve, reject) => {
    const reader = new win.FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = () => reject(reader.error || new Error('Could not read file'));
    reader.readAsText(file);
  });
}

export async function loadModelFile(app, file, persist, { win = globalThis.window } = {}) {
  app.controls.showLoading(`Loading ${file.name}...`);
  app.controls.setStatus(`Loading file: ${file.name}...`);
  try {
    const fileContent = await readBrowserFileText(file, win);
    app._persistedFileName = file.name;
    app._persistedFilePath = file.path || file.webkitRelativePath || file.name;
    app._persistedFileContent = fileContent;
    if (app._fileHandle) {
      await persist.saveFileHandle(app);
    }

    app.setEditorContent(fileContent);

    const loadResult = await app.api.loadFile(file);
    if (app.setIsolates) app.setIsolates(loadResult && loadResult.isolates, loadResult && loadResult.isolate);
    const argData = await app.api.getARG();
    if (argData && argData.elements) {
      applyArgSnapshot(app, app.activeSheetId || 'sheet-1', argData);
    }
    const conceptData = await app.api.getConceptGraph();
    if (conceptData && conceptData.elements) {
      applyConceptSnapshot(app, app.activeSheetId || 'sheet-1', conceptData);
    }
    app._persistedConceptRelations = conceptData;
    persist.setFileName(file.name, app._persistedFilePath);
    app.controls.setStatus(`Loaded: ${file.name}`, 'success');
    persist.save(app);
  } catch (err) {
    app.controls.setStatus(`Load failed: ${err.message}`, 'error');
    console.error('File load error:', err);
  } finally {
    app.controls.hideLoading();
  }
}

export function downloadTextFile(filename, content, mimeType, doc = globalThis.document) {
  const blob = new Blob([content || ''], { type: mimeType || 'text/plain' });
  const url = URL.createObjectURL(blob);
  const a = doc.createElement('a');
  a.href = url;
  a.download = filename || 'download.txt';
  doc.body.appendChild(a);
  a.click();
  doc.body.removeChild(a);
  URL.revokeObjectURL(url);
}

export async function downloadModel(app) {
  app.controls.setStatus('Downloading...');
  try {
    const content = app._editorContent();
    if (!content) {
      app.controls.setStatus('No model loaded to download', 'error');
      return undefined;
    }
    const filename = app._persistedFileName || 'model.ivy';
    app.downloadTextFile(filename, content, 'text/plain');
    app.controls.setStatus(`Downloaded: ${filename}`, 'success');
    return true;
  } catch (err) {
    app.controls.setStatus(`Download failed: ${err.message}`, 'error');
    return false;
  }
}

export function downloadModelForUnsupportedSave(app, content, persist) {
  const filename = app._persistedFileName || 'model.ivy';
  try {
    app.downloadTextFile(filename, content, 'text/plain');
    app._persistedFileContent = content;
    app._savedFileContent = content;
    app._updateEditorLabel();
    persist.save(app);
    app.controls.setStatus(`Downloaded edited copy: ${filename}. In Firefox, choose the original file to overwrite.`, 'success');
    return true;
  } catch (err) {
    app.controls.setStatus(`Download failed: ${err.message}`, 'error');
    return false;
  }
}

export async function saveModel(app, persist, { win = globalThis.window } = {}) {
  const content = app._editorContent();
  const dirty = app._editorDirty();
  let saveProgress = dirty ? app._showSaveProgress('Saving...') : null;
  try {
    await app._restoreFileHandleForCurrentFile();
    if (app._fileHandle) {
      const writableAllowed = await app._ensureFileHandleWritable();
      if (!writableAllowed) {
        app.controls.setStatus('Save permission denied', 'error');
        return false;
      }
      const saveDecision = await app._confirmNoExternalChangeBeforeSave(content);
      if (saveDecision === 'skip') {
        return false;
      }
      if (!dirty && saveDecision !== 'overwrite') {
        app._updateEditorLabel();
        return true;
      }
      const writable = await app._fileHandle.createWritable();
      await writable.write(content);
      await writable.close();
      app._persistedFileContent = content;
      app._savedFileContent = content;
      app._updateEditorLabel();
      app.controls.setStatus(`Saved: ${app._persistedFileName}`, 'success');
      return true;
    }
    if (saveProgress && win.showSaveFilePicker) {
      app._hideSaveProgress();
      saveProgress = null;
    }
    if (!dirty) {
      app._updateEditorLabel();
      return true;
    }
    if (!win.showSaveFilePicker) {
      return app.downloadModelForUnsupportedSave(content);
    }
    return app.saveAs({ explainMissingHandle: true });
  } catch (err) {
    app.controls.setStatus(`Save failed: ${err.message}`, 'error');
    return false;
  } finally {
    if (saveProgress) {
      app._hideSaveProgress();
    }
  }
}

export async function saveModelAs(app, persist, {
  options = {},
  win = globalThis.window,
}: { options?: any; win?: any } = {}) {
  const content = app._editorContent();
  if (!content) {
    app.controls.setStatus('No model loaded to save', 'error');
    return false;
  }
  let saveProgress = null;
  try {
    if (!win.showSaveFilePicker) {
      return app.downloadModelForUnsupportedSave(content);
    }
    if (options.explainMissingHandle) {
      app.showSaveAsExplanationNotice();
    }
    let handle;
    try {
      handle = await win.showSaveFilePicker({
        suggestedName: app._persistedFileName || 'model.ivy',
        types: [{
          description: 'Ivy files',
          accept: { 'text/plain': ['.ivy'] },
        }],
      });
    } finally {
      if (options.explainMissingHandle) {
        app.hideSaveAsExplanationNotice();
      }
    }
    saveProgress = app._showSaveProgress('Saving...');
    const writable = await handle.createWritable();
    await writable.write(content);
    await writable.close();

    app._fileHandle = handle;
    app._persistedFileName = handle.name;
    app._persistedFilePath = handle.name;
    app._persistedFileContent = content;
    app._savedFileContent = content;
    await persist.saveFileHandle(app);
    persist.setFileName(handle.name, app._persistedFilePath);
    app._updateEditorLabel();
    app.controls.setStatus(`Saved: ${handle.name}`, 'success');
    return true;
  } catch (err) {
    if (err.name === 'AbortError') {
      app.controls.setStatus('Save cancelled');
    } else {
      app.controls.setStatus(`Save as failed: ${err.message}`, 'error');
    }
    return false;
  } finally {
    if (saveProgress) {
      app._hideSaveProgress();
    }
  }
}

export async function closeCurrentFile(app) {
  if (app._editorDirty()) {
    const choice = await app.showDirtyCloseDialog();
    if (choice === 'cancel') {
      app.controls.setStatus('Close cancelled');
      return;
    }
    if (choice === 'save') {
      const saved = await app.save();
      if (!saved) {
        return;
      }
    }
  }
  await app.newModel({ skipSaveCurrent: true });
}

export async function newModel(app, persist, {
  options = {},
  doc = globalThis.document,
}: { options?: any; doc?: Document } = {}) {
  if (!options.skipSaveCurrent && app._persistedFileContent) {
    persist.save(app);
  }
  app._rememberLastOpenFile();

  try {
    await createSession(app.api, { controls: app.controls });
    app.updateSessionDisplay(persist.getSessionIdFromURL() || app.api.sessionId);
    persist.setSessionIdInURL(app.api.sessionId);
    connectSessionEvents(app.api, app.handleEvent.bind(app), app.handleConnectionLost && app.handleConnectionLost.bind(app));
  } catch (_err) {
    return;
  }

  app.argGraph.cy.elements().remove();
  app.conceptGraph.cy.elements().remove();

  app._persistedFileName = '';
  app._persistedFilePath = '';
  app._persistedFileContent = '';
  app._persistedConceptRelations = null;
  app._fileHandle = null;
  app._savedFileContent = null;
  app.selectedArgNode = null;

  app.setEditorContent('');
  persist.setFileName('');
  app._updateReopenLastFileButton();
  const tbody = doc && doc.getElementById('state-checkbox-body');
  if (tbody) tbody.innerHTML = '';

  app.controls.setStatus('New model — load an .ivy file to begin', 'success');
}
