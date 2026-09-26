import { applyArgSnapshot, applyConceptSnapshot } from './uiDataRenderService.ts';
import { applyModelToggleUpdates } from './conceptVisibilityService.ts';
import {
  selectConceptSelections,
  selectSheet,
  selectStateToggles,
} from '../models/uiDataSelectors.ts';
import { normalizeArgNodeId } from '../models/uiDataModel.ts';

const STORAGE_SESSIONS = 'ivy_sessions';
const STORAGE_LAST_SESSION = 'ivy_last_session';
const SESSION_PREFIX = 'ivy_sess_';
const EDITOR_KEYMAPS = new Set(['sublime', 'emacs', 'vim']);
const RECENT_SPEC_PREFIX = 'recent-spec:';

function getWindow(win) {
  return win || globalThis.window;
}

function getLocalStorage(win) {
  return win && win.localStorage;
}

function makeHandleKey(persist, appOrState) {
  if (!appOrState) return '';
  let sid = '';
  if (appOrState.api && appOrState.api.sessionId) {
    sid = persist.getSessionIdFromURL() || appOrState.api.sessionId;
  } else {
    sid = appOrState.sessionId || '';
  }
  const fileName = appOrState._persistedFileName || appOrState.fileName || '';
  const filePath = appOrState._persistedFilePath || appOrState.filePath || fileName;
  return `${sid}|${filePath}|${fileName}`;
}

function readJson(storage, key, fallback = null) {
  const raw = storage && storage.getItem(key);
  if (!raw) return fallback;
  return JSON.parse(raw);
}

function recentSpecSessionId(state) {
  if (!state || !state.fileName || state.fileName === '(unnamed)') return '';
  const identity = state.filePath || state.fileName;
  return identity ? `${RECENT_SPEC_PREFIX}${identity}` : '';
}

function argSnapshotNodeIds(argData) {
  const ids = new Set();
  const elements = Array.isArray(argData && argData.elements) ? argData.elements : [];
  for (const element of elements) {
    if (element && element.group && element.group !== 'nodes') continue;
    const data = (element && element.data) || {};
    for (const key of ['id', 'obj']) {
      const normalized = normalizeArgNodeId(data[key]);
      if (normalized) ids.add(normalized);
    }
  }
  return ids;
}

function restorableSelectedArgNode(argData, nodeId) {
  const selected = normalizeArgNodeId(nodeId);
  if (!selected) return null;
  return argSnapshotNodeIds(argData).has(selected) ? selected : null;
}

export function createIvyPersist(winArg = globalThis.window) {
  const win = getWindow(winArg);
  const doc = win && win.document;

  const persist = {
    MAX_SESSIONS: 1000,
    HANDLE_DB: 'ivy_file_handles',
    HANDLE_STORE: 'handles',

    save(app) {
      if (!app || !app.api || !app.api.sessionId) return;
      const storage = getLocalStorage(win);
      if (!storage) return;
      const sid = persist.getSessionIdFromURL() || app.api.sessionId;
      try {
        const state = {
          sessionId: sid,
          timestamp: Date.now(),
          fileName: app._persistedFileName || '',
          filePath: app._persistedFilePath || app._persistedFileName || '',
          fileContent: app._persistedFileContent || '',
          activeIsolate: app.activeIsolate || '',
          availableIsolates: Array.isArray(app.availableIsolates) ? app.availableIsolates.slice() : [],
          editorKeymap: typeof app.getEditorKeymap === 'function' ? app.getEditorKeymap() : persist._getEditorKeymap(),
          selectedArgNode: persist._selectedArgNode(app),
          uiMode: typeof app.getUIMode === 'function' ? app.getUIMode() : persist._getUIMode(),
          mode: persist._getMode(),
          conceptSelections: persist._conceptSelections(app),
          toggles: persist.getToggles(app),
          analysisState: app.buildAnalysisState ? app.buildAnalysisState() : null,
        };
        storage.setItem(`${SESSION_PREFIX}${sid}`, JSON.stringify(state));
        const recentSid = recentSpecSessionId(state);
        if (recentSid && recentSid !== sid) {
          storage.setItem(`${SESSION_PREFIX}${recentSid}`, JSON.stringify(state));
          persist._updateSessionList(recentSid);
        }
        storage.setItem(STORAGE_LAST_SESSION, sid);
        persist._updateSessionList(sid);
      } catch (err) {
        console.warn('IvyPersist.save failed:', err);
      }
    },

    load() {
      const storage = getLocalStorage(win);
      if (!storage) return null;
      try {
        const sid = persist.getSessionIdFromURL() || storage.getItem(STORAGE_LAST_SESSION);
        if (!sid) return null;
        return readJson(storage, `${SESSION_PREFIX}${sid}`, null);
      } catch (err) {
        console.warn('IvyPersist.load failed:', err);
        return null;
      }
    },

    loadSession(sid) {
      const storage = getLocalStorage(win);
      if (!storage || !sid) return null;
      try {
        return readJson(storage, `${SESSION_PREFIX}${sid}`, null);
      } catch (_err) {
        return null;
      }
    },

    deleteSession(sid) {
      const storage = getLocalStorage(win);
      if (!storage || !sid) return;
      try {
        storage.removeItem(`${SESSION_PREFIX}${sid}`);
        const ids = readJson(storage, STORAGE_SESSIONS, []).filter((id) => id !== sid);
        storage.setItem(STORAGE_SESSIONS, JSON.stringify(ids));
        if (storage.getItem(STORAGE_LAST_SESSION) === sid) {
          storage.removeItem(STORAGE_LAST_SESSION);
        }
      } catch (err) {
        console.warn('IvyPersist.deleteSession failed:', err);
      }
    },

    clearSavedSessions() {
      const storage = getLocalStorage(win);
      if (!storage) return 0;
      try {
        const keys = new Set([STORAGE_SESSIONS, STORAGE_LAST_SESSION]);
        const ids = readJson(storage, STORAGE_SESSIONS, []);
        if (Array.isArray(ids)) {
          ids.forEach((id) => {
            if (id) keys.add(`${SESSION_PREFIX}${id}`);
          });
        }
        if (typeof storage.length === 'number' && typeof storage.key === 'function') {
          for (let i = 0; i < storage.length; i++) {
            const key = storage.key(i);
            if (key && key.startsWith(SESSION_PREFIX)) keys.add(key);
          }
        }
        keys.forEach((key) => storage.removeItem(key));
        return keys.size;
      } catch (err) {
        console.warn('IvyPersist.clearSavedSessions failed:', err);
        return 0;
      }
    },

    listSessions() {
      const storage = getLocalStorage(win);
      if (!storage) return [];
      try {
        return readJson(storage, STORAGE_SESSIONS, []).flatMap((id) => {
          const session = persist.loadSession(id);
          if (!session) return [];
          return [{
            id,
            fileName: session.fileName || '(unnamed)',
            filePath: session.filePath || session.fileName || '',
            timestamp: session.timestamp || 0,
          }];
        });
      } catch (_err) {
        return [];
      }
    },

    _handleKey(appOrState) {
      return makeHandleKey(persist, appOrState);
    },

    _openHandleDB() {
      return new Promise((resolve, reject) => {
        if (!win || !win.indexedDB) {
          reject(new Error('IndexedDB unavailable'));
          return;
        }
        const req = win.indexedDB.open(persist.HANDLE_DB, 1);
        req.onupgradeneeded = () => {
          req.result.createObjectStore(persist.HANDLE_STORE);
        };
        req.onsuccess = () => resolve(req.result);
        req.onerror = () => reject(req.error || new Error('open IndexedDB failed'));
      });
    },

    async saveFileHandle(app) {
      if (!app || !app._fileHandle) return;
      try {
        const key = persist._handleKey(app);
        if (!key) return;
        let legacyKey = '';
        if (app._persistedFileName && app.api && app.api.sessionId) {
          const sid = persist.getSessionIdFromURL() || app.api.sessionId;
          legacyKey = `${sid}|${app._persistedFileName}|${app._persistedFileName}`;
        }
        const db = await persist._openHandleDB() as any;
        await new Promise((resolve, reject) => {
          const tx = db.transaction(persist.HANDLE_STORE, 'readwrite');
          const store = tx.objectStore(persist.HANDLE_STORE);
          store.put(app._fileHandle, key);
          if (legacyKey && legacyKey !== key) {
            store.put(app._fileHandle, legacyKey);
          }
          tx.oncomplete = resolve;
          tx.onerror = () => reject(tx.error || new Error('store file handle failed'));
        });
        db.close();
      } catch (err) {
        console.warn('IvyPersist.saveFileHandle failed:', err);
      }
    },

    async loadFileHandle(state) {
      try {
        const key = persist._handleKey(state);
        if (!key) return null;
        const db = await persist._openHandleDB() as any;
        const keys = [key];
        if (state.fileName) {
          const legacyKey = `${state.sessionId || ''}|${state.fileName}|${state.fileName}`;
          if (legacyKey && legacyKey !== key) keys.push(legacyKey);
        }
        const fileName = state.fileName || '';
        const filePath = state.filePath || fileName;
        const handle = await new Promise((resolve, reject) => {
          const tx = db.transaction(persist.HANDLE_STORE, 'readonly');
          const store = tx.objectStore(persist.HANDLE_STORE);
          let index = 0;
          const finishByScanning = () => {
            if (!fileName) {
              resolve(null);
              return;
            }
            const cursorReq = store.openCursor();
            let nameOnlyMatch = null;
            cursorReq.onsuccess = () => {
              const cursor = cursorReq.result;
              if (!cursor) {
                resolve(nameOnlyMatch);
                return;
              }
              const [storedSession, storedPath = '', storedName = ''] = String(cursor.key || '').split('|');
              void storedSession;
              if (storedName === fileName && (storedPath === filePath || storedPath === fileName)) {
                resolve(cursor.value);
                return;
              }
              if (!nameOnlyMatch && storedName === fileName) {
                nameOnlyMatch = cursor.value;
              }
              cursor.continue();
            };
            cursorReq.onerror = () => reject(cursorReq.error || new Error('scan file handles failed'));
          };
          const tryNext = () => {
            if (index >= keys.length) {
              finishByScanning();
              return;
            }
            const req = store.get(keys[index]);
            index += 1;
            req.onsuccess = () => {
              if (req.result) {
                resolve(req.result);
              } else {
                tryNext();
              }
            };
            req.onerror = () => reject(req.error || new Error('load file handle failed'));
          };
          tryNext();
        });
        db.close();
        return handle || null;
      } catch (err) {
        console.warn('IvyPersist.loadFileHandle failed:', err);
        return null;
      }
    },

    async restore(app, state) {
      if (!state || !state.fileContent) return false;
      let modelLoad = null;
      app.controls.setStatus('Restoring session...');
      try {
        app._persistedFileName = state.fileName;
        app._persistedFilePath = state.filePath || state.fileName || '';
        app._fileHandle = await persist.loadFileHandle(state);
        let restoredContent = state.fileContent || '';
        if (app._fileHandle) {
          try {
            const diskFile = await app._fileHandle.getFile();
            restoredContent = await diskFile.text();
            app._persistedFileName = diskFile.name || app._persistedFileName;
            app._persistedFilePath = diskFile.path || app._persistedFilePath || app._persistedFileName;
          } catch (err) {
            console.warn('IvyPersist.restore: could not read disk file handle, using cached content:', err);
          }
        }
        app._persistedFileContent = restoredContent;
        if (app.setEditorContent) app.setEditorContent(restoredContent);
        if (app._updateEditorLabel) app._updateEditorLabel();
        if (app.setIsolates) app.setIsolates(state.availableIsolates || [], state.activeIsolate || '');

        let parseOk = true;
        let argData = null;
        let conceptData = null;
        modelLoad = app && typeof app._beginModelLoad === 'function'
          ? app._beginModelLoad({
            reason: 'session-restore',
            filename: app._persistedFileName || state.fileName || 'restored.ivy',
            isolate: state.activeIsolate || '',
            content: restoredContent,
          })
          : null;
        try {
          const BlobCtor = win.Blob || globalThis.Blob;
          const FileCtor = win.File || globalThis.File;
          const blob = new BlobCtor([restoredContent], { type: 'text/plain' });
          const file = new FileCtor([blob], app._persistedFileName || state.fileName || 'restored.ivy');
          const loadResult = await app.api.loadFile(file, {
            isolate: state.activeIsolate || '',
            filename: app._persistedFilePath || app._persistedFileName || state.fileName || 'restored.ivy',
          });
          if (modelLoad) modelLoad.loadResult = loadResult;
          if (!modelLoad && app.setIsolates) {
            app.setIsolates(
              (loadResult && loadResult.isolates) || state.availableIsolates || [],
              (loadResult && loadResult.isolate) || state.activeIsolate || '',
            );
          }
        } catch (err) {
          parseOk = false;
          console.warn('IvyPersist.restore: server rejected file (parse error):', err.message || String(err));
        }

        if (parseOk) {
          argData = await app.api.getARG();
          if (modelLoad && typeof app._isCurrentModelLoad === 'function' && !app._isCurrentModelLoad(modelLoad)) return false;
          if (!modelLoad && argData && argData.elements) {
            applyArgSnapshot(app, app.activeSheetId || 'sheet-1', argData);
          }
          conceptData = await app.api.getConceptGraph();
          if (modelLoad && typeof app._isCurrentModelLoad === 'function' && !app._isCurrentModelLoad(modelLoad)) return false;
          if (!modelLoad && conceptData && conceptData.elements) {
            applyConceptSnapshot(app, app.activeSheetId || 'sheet-1', conceptData);
          }
          if (modelLoad && typeof app._commitModelLoad === 'function') {
            app._commitModelLoad(modelLoad, { argData, conceptData, deferRender: true });
          }
          app._persistedConceptRelations = conceptData;
        } else if (modelLoad && typeof app._abortModelLoad === 'function') {
          app._abortModelLoad(modelLoad);
        }

        if (state.uiMode) {
          if (typeof app.setUIMode === 'function') app.setUIMode(state.uiMode);
          else persist._setUIMode(state.uiMode);
        }
        if (state.mode) persist._setMode(state.mode);
        if (state.editorKeymap) {
          if (typeof app.setEditorKeymap === 'function') app.setEditorKeymap(state.editorKeymap, { save: false });
          else persist._setEditorKeymap(state.editorKeymap);
        }
        const selectedArgNode = restorableSelectedArgNode(argData, state.selectedArgNode);
        if (selectedArgNode && app.uiDataStore) {
          app.uiDataStore.setSelectedArgNode(app.activeSheetId || 'sheet-1', selectedArgNode);
        }
        if (state.conceptSelections && app.uiDataStore) {
          app.uiDataStore.setConceptSelections(app.activeSheetId || 'sheet-1', state.conceptSelections);
        }
        if (state.toggles) await persist.applyToggles(app, state.toggles);

        if (state.analysisState && app.loadAnalysisStateObject) {
          state.analysisState.fileContent = restoredContent;
          state.analysisState.fileName = app._persistedFileName || state.fileName || state.analysisState.fileName;
          state.analysisState.filePath = app._persistedFilePath || state.filePath || state.analysisState.filePath;
          await app.loadAnalysisStateObject(state.analysisState, {
            preservePrimarySheetModel: parseOk,
            preserveStatus: parseOk,
            skipReloadContent: parseOk,
          });
        }

        persist.setFileName(state.fileName, state.filePath);
        persist.setSessionIdInURL(state.sessionId);
        if (app._fileHandle) await persist.saveFileHandle(app);

        app.controls.setStatus(
          `${parseOk ? 'Restored' : 'Loaded (parse error)'}: ${state.fileName || 'session'}`,
          parseOk ? 'success' : 'warning',
        );
        return true;
      } catch (err) {
        if (modelLoad && typeof app._abortModelLoad === 'function') app._abortModelLoad(modelLoad);
        console.error('IvyPersist.restore failed:', err);
        app.controls.setStatus(`Restore failed: ${err.message}`, 'error');
        return false;
      } finally {
        if (modelLoad && typeof app._finishModelLoad === 'function') app._finishModelLoad(modelLoad);
      }
    },

    _updateSessionList(sid) {
      const storage = getLocalStorage(win);
      if (!storage) return;
      const ids = readJson(storage, STORAGE_SESSIONS, [])
        .filter((id) => id !== sid);
      ids.unshift(sid);
      while (ids.length > persist.MAX_SESSIONS) {
        const old = ids.pop();
        storage.removeItem(`${SESSION_PREFIX}${old}`);
      }
      storage.setItem(STORAGE_SESSIONS, JSON.stringify(ids));
    },

    _getMode() {
      const select = doc && doc.getElementById('mode-select');
      return select ? select.value : 'pdr';
    },

    _setMode(mode) {
      const select = doc && doc.getElementById('mode-select');
      if (select && mode) select.value = mode;
    },

    _getEditorKeymap() {
      const checked = doc && doc.querySelector('input[name="keymap"]:checked') as HTMLInputElement | null;
      return checked ? checked.value : 'emacs';
    },

    _setEditorKeymap(keymap) {
      const normalized = EDITOR_KEYMAPS.has(keymap) ? keymap : 'emacs';
      const radio = doc && doc.querySelector(`input[name="keymap"][value="${normalized}"]`) as HTMLInputElement | null;
      if (radio) radio.checked = true;
    },

    _getUIMode() {
      const select = doc && doc.getElementById('ui-mode-select');
      return select ? select.value : 'cti';
    },

    _setUIMode(mode) {
      const normalized = mode === 'reachability' ? 'reachability' : 'cti';
      const select = doc && doc.getElementById('ui-mode-select');
      if (select) select.value = normalized;
      if (doc && doc.body) doc.body.setAttribute('data-ui-mode', normalized);
    },

    _selectedArgNode(app) {
      const sheet = selectSheet(app && app.uiDataModel, app && app.activeSheetId);
      return sheet ? sheet.selectedArgNode : null;
    },

    _conceptSelections(app) {
      return selectConceptSelections(selectSheet(app && app.uiDataModel, app && app.activeSheetId));
    },

    getToggles(app) {
      return selectStateToggles(selectSheet(app && app.uiDataModel, app && app.activeSheetId));
    },

    async applyToggles(app, toggles = {}) {
      const entries = Object.entries(toggles || {});
      if (!app || !app.api || typeof app.api.setToggles !== 'function' || entries.length === 0) return;
      const modelUpdates = [];
      for (const [key, value] of entries) {
        const split = String(key).split('|');
        const displayClass = split.pop();
        const edge = split.join('|');
        if (!edge || !displayClass) continue;
        const update = {
          edge,
          display_class: displayClass,
          value: !!value,
        };
        await app.api.setToggles(update);
        modelUpdates.push(update);
      }
      applyModelToggleUpdates(app, modelUpdates);
      if (typeof app.refreshConceptGraph === 'function') await app.refreshConceptGraph();
    },

    getSessionIdFromURL() {
      const hash = win && win.location && win.location.hash;
      return hash && hash.length > 1 ? hash.substring(1) : null;
    },

    setSessionIdInURL(sid) {
      if (sid && win && win.history) {
        win.history.replaceState(null, '', `#${sid}`);
      }
    },

    setFileName(fileName, filePath) {
      const el = doc && doc.getElementById('loaded-file');
      const display = filePath || fileName || '';
      if (el) {
        el.textContent = display;
        el.title = display;
      }
      const editorLabel = doc && doc.getElementById('model-editor-label');
      if (editorLabel) {
        const current = editorLabel.textContent || '';
        const shouldSetEditorLabel =
          !current ||
          current === '(unsaved file)' ||
          current.indexOf(fileName || display) < 0;
        if (shouldSetEditorLabel) {
          const labelText = display || '(unsaved file)';
          editorLabel.textContent = labelText;
          editorLabel.title = `Editing: ${labelText}`;
        }
      }
    },

    truncatePath(path, maxChars) {
      if (!path || path.length <= maxChars) return path || '';
      let sep = path.lastIndexOf('/');
      if (sep < 0) sep = path.lastIndexOf('\\');
      const dir = sep >= 0 ? path.substring(0, sep) : '';
      if (!dir) return '';
      if (dir.length <= maxChars) return dir;
      return `...${dir.substring(dir.length - maxChars + 3)}`;
    },
  };

  return persist;
}
