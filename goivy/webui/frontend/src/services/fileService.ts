import { connectSessionEvents, createSession } from './sessionService.ts';
import { applyArgSnapshot, applyConceptSnapshot } from './uiDataRenderService.ts';
import { resetEditorUndoHistory } from './editorService.ts';
import { logStateRelationTableClear } from './conceptVisibilityService.ts';
import { saveMimeType, savePickerOptions } from './saveDialogService.ts';

export const NEW_MODEL_STARTER_CONTENT = '#lang ivy1.8\n\n';
export const PROJECT_ROOT = '/project';

function normalizeProjectRelativePath(path) {
  const parts = String(path || '')
    .replace(/\\/g, '/')
    .split('/')
    .filter((part) => part && part !== '.');
  const normalized = [];
  for (const part of parts) {
    if (part === '..') {
      normalized.pop();
      continue;
    }
    normalized.push(part);
  }
  return normalized.join('/');
}

function projectVirtualPath(relativePath) {
  const rel = normalizeProjectRelativePath(relativePath);
  return rel ? `${PROJECT_ROOT}/${rel}` : '';
}

function basename(path) {
  const rel = normalizeProjectRelativePath(path);
  const parts = rel.split('/').filter(Boolean);
  return parts.length > 0 ? parts[parts.length - 1] : '';
}

function projectFilesForApp(app) {
  return Array.isArray(app && app._projectFiles) ? app._projectFiles : [];
}

export function modelLoadOptionsForApp(app, options = {}) {
  const projectFiles = projectFilesForApp(app);
  if (projectFiles.length === 0) return { ...options };
  return { ...options, projectFiles };
}

function findProjectFile(projectFiles, file, content) {
  if (!Array.isArray(projectFiles) || !file) return null;
  const webkitRelativePath = normalizeProjectRelativePath(file.webkitRelativePath || '');
  if (webkitRelativePath) {
    const exact = projectFiles.find((candidate) => candidate.path === webkitRelativePath);
    if (exact) return exact;
  }
  const fileName = file.name || basename(file.path || '');
  if (!fileName) return null;
  const sameName = projectFiles.filter((candidate) => basename(candidate.path) === fileName);
  const sameContent = sameName.filter((candidate) => candidate.data === content);
  if (sameContent.length === 1) return sameContent[0];
  if (sameName.length === 1) return sameName[0];
  return null;
}

function projectFilenameForBrowserFile(app, file, content) {
  const match = findProjectFile(projectFilesForApp(app), file, content);
  return match ? projectVirtualPath(match.path) : '';
}

async function readProjectDirectory(dirHandle, prefix = '') {
  const files = [];
  if (!dirHandle || typeof dirHandle.entries !== 'function') return files;
  for await (const [name, handle] of dirHandle.entries()) {
    const relativePath = normalizeProjectRelativePath(prefix ? `${prefix}/${name}` : name);
    if (!relativePath) continue;
    if (handle.kind === 'directory') {
      files.push(...await readProjectDirectory(handle, relativePath));
    } else if (handle.kind === 'file' && relativePath.endsWith('.ivy') && typeof handle.getFile === 'function') {
      const file = await handle.getFile();
      files.push({ path: relativePath, data: await file.text() });
    }
  }
  files.sort((a, b) => (a.path < b.path ? -1 : (a.path > b.path ? 1 : 0)));
  return files;
}

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

export async function chooseAndLoadProjectFolder(app, persist, {
  win = globalThis.window,
} = {}) {
  if (!win || typeof win.showDirectoryPicker !== 'function') {
    app.controls.setStatus('Open project folder requires a browser with directory picker support', 'error');
    return false;
  }
  try {
    if (app.controls.showLoading) app.controls.showLoading('Opening project folder...');
    const handle = await win.showDirectoryPicker({ mode: 'read' });
    const projectFiles = await readProjectDirectory(handle);
    app._projectDirectoryHandle = handle;
    app._projectRootName = handle.name || 'project';
    app._projectFiles = projectFiles;

    const currentFile = {
      name: app._persistedFileName || basename(app._persistedFilePath || ''),
      path: app._persistedFilePath || '',
    };
    const currentMatch = findProjectFile(projectFiles, currentFile, app._persistedFileContent || '');
    if (currentMatch) {
      app._persistedFilePath = projectVirtualPath(currentMatch.path);
      if (app.api && typeof app.api.reloadContent === 'function' && app._persistedFileContent) {
        const loadResult = await app.api.reloadContent(
          app._persistedFileContent,
          app._persistedFilePath,
          modelLoadOptionsForApp(app, { isolate: app.activeIsolate || '' }),
        );
        if (app.setIsolates) {
          app.setIsolates(loadResult && loadResult.isolates, loadResult && loadResult.isolate);
        }
        await refreshLoadedModelSnapshots(app);
      }
    }

    if (persist && typeof persist.setFileName === 'function' && app._persistedFileName) {
      persist.setFileName(app._persistedFileName, app._persistedFilePath || app._persistedFileName);
    }
    if (persist && typeof persist.save === 'function') persist.save(app);
    app.controls.setStatus(`Opened project folder: ${app._projectRootName} (${projectFiles.length} Ivy files)`, 'success');
    return true;
  } catch (err) {
    if (err.name !== 'AbortError') {
      app.controls.setStatus(`Open project folder failed: ${err.message}`, 'error');
    }
    return false;
  } finally {
    if (app.controls.hideLoading) app.controls.hideLoading();
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

export function preparePrimarySheetForModelLoad(app, {
  doc = globalThis.document,
} = {}) {
  const sheetId = 'sheet-1';
  const runtimeSheet = app && app.sheets && app.sheets[sheetId];
  if (runtimeSheet) {
    runtimeSheet.reachabilityOnly = false;
    runtimeSheet.visualOnly = false;
  }
  const modelSheet = app && app.uiDataModel && app.uiDataModel.sheets && app.uiDataModel.sheets[sheetId];
  if (modelSheet) {
    modelSheet.reachabilityOnly = false;
    modelSheet.visualOnly = false;
  }
  if (app && typeof app.removeAnalysisStateExtraSheets === 'function') {
    app.removeAnalysisStateExtraSheets();
  }
  if (app && typeof app.switchSheet === 'function' && (!app.sheetExists || app.sheetExists(sheetId))) {
    app.switchSheet(sheetId);
  } else if (app) {
    app.activeSheetId = sheetId;
  }
  const sheetEl = doc && doc.getElementById(sheetId);
  if (sheetEl) {
    sheetEl.classList.remove('reachability-only-sheet');
    sheetEl.removeAttribute('data-sheet-layout');
  }
  return sheetId;
}

export async function refreshLoadedModelSnapshots(app, {
  doc = globalThis.document,
  modelLoad = null,
} = {}) {
  if (modelLoad && app && typeof app._isCurrentModelLoad === 'function' && !app._isCurrentModelLoad(modelLoad)) {
    return { stale: true, sheetId: 'sheet-1', argData: null, conceptData: null };
  }
  if (modelLoad && app && typeof app._commitModelLoad === 'function') {
    if (!app.api || typeof app.api.getARG !== 'function' || typeof app.api.getConceptGraph !== 'function') {
      const loadResult = modelLoad.loadResult || null;
      if (loadResult && app.setIsolates) {
        app.setIsolates(loadResult.isolates || app.availableIsolates || [], loadResult.isolate || modelLoad.isolate || app.activeIsolate || '');
      }
      if (typeof app._markModelStateFresh === 'function') app._markModelStateFresh(modelLoad.content);
      return { snapshotUnavailable: true, sheetId: 'sheet-1', argData: null, conceptData: null };
    }
    const argData = await app.api.getARG();
    if (typeof app._isCurrentModelLoad === 'function' && !app._isCurrentModelLoad(modelLoad)) {
      return { stale: true, sheetId: 'sheet-1', argData, conceptData: null };
    }
    const conceptData = await app.api.getConceptGraph();
    if (typeof app._isCurrentModelLoad === 'function' && !app._isCurrentModelLoad(modelLoad)) {
      return { stale: true, sheetId: 'sheet-1', argData, conceptData };
    }
    app._commitModelLoad(modelLoad, { argData, conceptData, doc });
    return { sheetId: 'sheet-1', argData, conceptData };
  }
  const sheetId = preparePrimarySheetForModelLoad(app, { doc });
  const argData = await app.api.getARG();
  if (argData) {
    applyArgSnapshot(app, sheetId, argData);
  }
  const conceptData = await app.api.getConceptGraph();
  if (conceptData) {
    applyConceptSnapshot(app, sheetId, conceptData);
  }
  app._persistedConceptRelations = conceptData;
  return { sheetId, argData, conceptData };
}

export async function loadModelFile(app, file, persist, { win = globalThis.window } = {}) {
  const modelLoad = app && typeof app._beginModelLoad === 'function'
    ? app._beginModelLoad({ reason: 'file-load', filename: file && file.name })
    : null;
  app.controls.showLoading(`Loading ${file.name}...`);
  app.controls.setStatus(`Loading file: ${file.name}...`);
  try {
    const fileContent = await readBrowserFileText(file, win);
    if (modelLoad && typeof app._isCurrentModelLoad === 'function' && !app._isCurrentModelLoad(modelLoad)) return;
    if (modelLoad) modelLoad.content = fileContent;
    app._persistedFileName = file.name;
    app._persistedFilePath = projectFilenameForBrowserFile(app, file, fileContent) || file.path || file.webkitRelativePath || file.name;
    app._persistedFileContent = fileContent;
    if (app._fileHandle) {
      await persist.saveFileHandle(app);
    }

    if (app.setIsolates) app.setIsolates([], '');
    app.setEditorContent(fileContent);

    const loadResult = await app.api.loadFile(file, {
      ...modelLoadOptionsForApp(app, { isolate: '' }),
      filename: app._persistedFilePath || file.name,
    });
    if (modelLoad && typeof app._isCurrentModelLoad === 'function' && !app._isCurrentModelLoad(modelLoad)) return;
    if (modelLoad) modelLoad.loadResult = loadResult;
    if (!modelLoad && app.setIsolates) app.setIsolates(loadResult && loadResult.isolates, loadResult && loadResult.isolate);
    await refreshLoadedModelSnapshots(app, { modelLoad });
    if (modelLoad && typeof app._isCurrentModelLoad === 'function' && !app._isCurrentModelLoad(modelLoad)) return;
    persist.setFileName(file.name, app._persistedFilePath);
    app.controls.setStatus(`Loaded: ${file.name}`, 'success');
    persist.save(app);
  } catch (err) {
    if (modelLoad && typeof app._abortModelLoad === 'function') app._abortModelLoad(modelLoad);
    app.controls.setStatus(`Load failed: ${err.message}`, 'error');
    console.error('File load error:', err);
  } finally {
    if (modelLoad && typeof app._finishModelLoad === 'function') app._finishModelLoad(modelLoad);
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
    app.downloadTextFile(filename, content, saveMimeType('model'));
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
    app.downloadTextFile(filename, content, saveMimeType('model'));
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
      handle = await win.showSaveFilePicker(savePickerOptions('model', app._persistedFileName || 'model.ivy'));
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
  if (app.setIsolates) app.setIsolates([], '');

  app.setEditorContent(NEW_MODEL_STARTER_CONTENT);
  persist.setFileName('');
  app._updateReopenLastFileButton();
  const tbody = doc && doc.getElementById('state-checkbox-body');
  if (tbody) {
    logStateRelationTableClear('newModel: explicit reset after clearing current model', app, {
      rowCount: 0,
      hadRows: tbody.children.length,
    });
    tbody.innerHTML = '';
  }

  app.controls.setStatus('New model — load an .ivy file to begin', 'success');
}
