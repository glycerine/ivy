import { applyArgSnapshot, applyConceptSnapshot } from './uiDataRenderService.ts';

const STORAGE_SESSIONS = 'ivy_sessions';
const STORAGE_LAST_SESSION = 'ivy_last_session';
const SESSION_PREFIX = 'ivy_sess_';

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
          selectedArgNode: app.selectedArgNode || null,
          mode: persist._getMode(),
          selectedConceptNodes: persist._getSelectedConceptNodes(app),
          toggles: persist._getToggles(),
          edgeVisibility: {},
          labelVisibility: {},
          argElements: persist._getCyElements(app.argGraph),
          conceptElements: persist._getCyElements(app.conceptGraph),
          conceptRelations: app._persistedConceptRelations || null,
          analysisState: app.buildAnalysisState ? app.buildAnalysisState() : null,
        };
        storage.setItem(`${SESSION_PREFIX}${sid}`, JSON.stringify(state));
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

        let parseOk = true;
        let conceptData = null;
        try {
          const BlobCtor = win.Blob || globalThis.Blob;
          const FileCtor = win.File || globalThis.File;
          const blob = new BlobCtor([restoredContent], { type: 'text/plain' });
          const file = new FileCtor([blob], app._persistedFileName || state.fileName || 'restored.ivy');
          await app.api.loadFile(file);
        } catch (err) {
          parseOk = false;
          console.warn('IvyPersist.restore: server rejected file (parse error):', err.message || String(err));
        }

        if (parseOk) {
          const argData = await app.api.getARG();
          if (argData && argData.elements) {
            applyArgSnapshot(app, app.activeSheetId || 'sheet-1', argData);
          }
          conceptData = await app.api.getConceptGraph();
          if (conceptData && conceptData.elements) {
            applyConceptSnapshot(app, app.activeSheetId || 'sheet-1', conceptData);
          }
          app._persistedConceptRelations = conceptData;
        }

        if (state.mode) persist._setMode(state.mode);
        if (state.selectedConceptNodes && app.conceptGraph && app.conceptGraph.cy) {
          state.selectedConceptNodes.forEach((id) => {
            const node = app.conceptGraph.cy.getElementById(id);
            if (node.length > 0) node.addClass('selected_node');
          });
        }
        if (state.toggles) persist._setToggles(state.toggles);

        persist._buildVisibilityFromCheckboxes();

        if (state.selectedArgNode) app.selectedArgNode = state.selectedArgNode;
        if (state.analysisState && app.loadAnalysisStateObject) {
          state.analysisState.fileContent = restoredContent;
          state.analysisState.fileName = app._persistedFileName || state.fileName || state.analysisState.fileName;
          state.analysisState.filePath = app._persistedFilePath || state.filePath || state.analysisState.filePath;
          await app.loadAnalysisStateObject(state.analysisState);
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
        console.error('IvyPersist.restore failed:', err);
        app.controls.setStatus(`Restore failed: ${err.message}`, 'error');
        return false;
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

    _getSelectedConceptNodes(app) {
      const ids = [];
      if (app.conceptGraph && app.conceptGraph.cy) {
        app.conceptGraph.cy.nodes('.selected_node').forEach((node) => {
          ids.push(node.id());
        });
      }
      return ids;
    },

    _getToggles() {
      const toggles = {};
      const table = doc && doc.getElementById('state-checkbox-body');
      if (!table) return toggles;
      table.querySelectorAll('input[type="radio"], input[type="checkbox"]').forEach((input) => {
        if (input.name) toggles[`${input.name}|${input.value}`] = input.checked;
      });
      return toggles;
    },

    _setToggles(toggles = {}) {
      const table = doc && doc.getElementById('state-checkbox-body');
      if (!table) return;
      table.querySelectorAll('input[type="radio"], input[type="checkbox"]').forEach((input) => {
        const key = `${input.name}|${input.value}`;
        if (Object.prototype.hasOwnProperty.call(toggles, key)) {
          input.checked = toggles[key];
        }
      });
    },

    _getCyElements(graph) {
      if (!graph || !graph.cy) return null;
      return graph.cy.json().elements;
    },

    _buildVisibilityFromCheckboxes() {
      const edges = {};
      const labels = {};
      const tbody = doc && doc.getElementById('state-checkbox-body');
      if (!tbody) return { edges, labels };
      const edgeClasses = ['all_to_all', 'edge_unknown', 'none_to_none', 'transitive'];
      const labelClasses = ['node_necessarily', 'node_maybe', 'node_necessarily_not'];
      tbody.querySelectorAll('tr').forEach((row) => {
        const inputs = Array.from(row.querySelectorAll('input[type="checkbox"]')) as HTMLInputElement[];
        const nameCell = row.querySelector('.name-col a');
        const name = nameCell ? nameCell.textContent.trim() : '';
        if (!name) return;
        edges[name] = {};
        labels[name] = {};
        edgeClasses.forEach((className, index) => {
          if (inputs[index]) edges[name][className] = inputs[index].checked;
        });
        labelClasses.forEach((className, index) => {
          if (inputs[index]) labels[name][className] = inputs[index].checked;
        });
      });
      return { edges, labels };
    },

    _buildEdgeVisibilityFromCheckboxes() {
      return persist._buildVisibilityFromCheckboxes().edges;
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
