import { describe, expect, it, vi } from 'vitest';
import {
  confirmNoExternalChangeBeforeSave,
  downloadModelForUnsupportedSave,
  ensureFileHandleWritable,
  mergeDiskVersionIntoEditBuffer,
  rememberLastOpenFile,
  restoreFileHandleForCurrentFile,
  saveModel,
  saveModelAs,
  updateReopenLastFileButton,
} from './fileService.ts';
import { FakeControls, makePersist, makeWritableHandle } from '../test/fakes.ts';

function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

function abortError() {
  const err = new Error('The user aborted a request.');
  err.name = 'AbortError';
  return err;
}

function makeSaveApp({
  fileName = 'client.ivy',
  savedContent = 'old model',
  editorContent = 'new model',
  fileHandle = null,
  persist = makePersist(),
  win = {},
} = {}) {
  const controls = new FakeControls();
  const app = {
    _persistedFileName: fileName,
    _persistedFilePath: fileName,
    _persistedFileContent: savedContent,
    _savedFileContent: savedContent,
    _fileHandle: fileHandle,
    _saveInProgress: false,
    api: { sessionId: 's1' },
    controls,
    _editorContent: vi.fn(() => editorContent),
    _editorDirty: vi.fn(() => editorContent !== app._savedFileContent),
    _showSaveProgress: vi.fn(() => {
      app._saveInProgress = true;
      app._updateEditorLabel();
      return true;
    }),
    _hideSaveProgress: vi.fn(() => {
      app._saveInProgress = false;
      app._updateEditorLabel();
    }),
    _updateEditorLabel: vi.fn(() => {
      if (app._saveInProgress) {
        app.editorLabel = `${app._persistedFileName} [saving...]`;
      } else if (app._editorDirty()) {
        app.editorLabel = `** ${app._persistedFileName}`;
      } else {
        app.editorLabel = `${app._persistedFileName} [saved]`;
      }
    }),
    _restoreFileHandleForCurrentFile: vi.fn(() => restoreFileHandleForCurrentFile(app, persist)),
    _ensureFileHandleWritable: vi.fn(() => ensureFileHandleWritable(app)),
    _confirmNoExternalChangeBeforeSave: vi.fn(async () => 'ok'),
    downloadTextFile: vi.fn(),
    downloadModelForUnsupportedSave: vi.fn((content) => downloadModelForUnsupportedSave(app, content, persist)),
    saveAs: vi.fn((options) => saveModelAs(app, persist, { options, win })),
    showSaveAsExplanationNotice: vi.fn(() => {
      app.saveAsNoticeVisible = true;
    }),
    hideSaveAsExplanationNotice: vi.fn(() => {
      app.saveAsNoticeVisible = false;
    }),
  };
  return { app, persist, controls };
}

describe('fileService', () => {
  it('checks writable permissions only when the browser handle requires it', async () => {
    const app = {
      _fileHandle: {
        queryPermission: vi.fn(async () => 'prompt'),
        requestPermission: vi.fn(async () => 'granted'),
      },
    };

    await expect(ensureFileHandleWritable(app)).resolves.toBe(true);
    expect(app._fileHandle.queryPermission).toHaveBeenCalledWith({ mode: 'readwrite' });
    expect(app._fileHandle.requestPermission).toHaveBeenCalledWith({ mode: 'readwrite' });
  });

  it('builds conflict markers when both editor and disk changed', () => {
    expect(mergeDiskVersionIntoEditBuffer('base\n', 'editor\n', 'disk\n')).toBe([
      '<<<<<<< EDIT BUFFER',
      'editor',
      '||||||| LAST SAVED',
      'base',
      '=======',
      'disk',
      '>>>>>>> ON DISK',
      '',
    ].join('\n'));
  });

  it('clears editor undo history after merging disk changes into the editor buffer', async () => {
    const app = {
      _fileHandle: {},
      _savedFileContent: 'base\n',
      _persistedFileName: 'client.ivy',
      _persistedFileContent: 'editor\n',
      _readFileHandleContent: vi.fn(async () => 'disk\n'),
      showExternalChangeDialog: vi.fn(async () => 'merge'),
      cmEditor: {
        setValue: vi.fn(),
        clearHistory: vi.fn(),
      },
      _updateEditorLabel: vi.fn(),
      controls: new FakeControls(),
    };
    const persist = makePersist();

    await expect(confirmNoExternalChangeBeforeSave(app, 'editor\n', persist)).resolves.toBe('skip');

    expect(app.cmEditor.setValue).toHaveBeenCalledWith([
      '<<<<<<< EDIT BUFFER',
      'editor',
      '||||||| LAST SAVED',
      'base',
      '=======',
      'disk',
      '>>>>>>> ON DISK',
      '',
    ].join('\n'));
    expect(app.cmEditor.clearHistory).toHaveBeenCalledTimes(1);
  });

  it('remembers and displays the reopen-last-file affordance', () => {
    const app = {
      _persistedFileName: 'client.ivy',
      _fileHandle: { name: 'client.ivy' },
      api: { sessionId: 'server-session' },
    };
    const persist = {
      getSessionIdFromURL: vi.fn(() => 'stable-session'),
    };

    rememberLastOpenFile(app, persist);

    expect(app._lastClosedFileHandle).toBe(app._fileHandle);
    expect(app._lastClosedSessionId).toBe('stable-session');
    expect(app._lastClosedFileName).toBe('client.ivy');

    document.body.innerHTML = '<button id="file-reopen-last" style="display:none"></button>';
    app._fileHandle = null;
    app._persistedFileName = '';
    updateReopenLastFileButton(app, { doc: document });
    expect(document.getElementById('file-reopen-last').textContent).toBe('Re-open last file client.ivy');
    expect(document.getElementById('file-reopen-last').style.display).toBe('');
  });

  it('marks downloaded fallback saves as saved', () => {
    const app = {
      _persistedFileName: 'client.ivy',
      _savedFileContent: 'old',
      downloadTextFile: vi.fn(),
      _updateEditorLabel: vi.fn(),
      controls: {
        setStatus: vi.fn(),
      },
    };
    const persist = {
      save: vi.fn(),
    };

    expect(downloadModelForUnsupportedSave(app, 'new', persist)).toBe(true);
    expect(app.downloadTextFile).toHaveBeenCalledWith('client.ivy', 'new', 'text/plain');
    expect(app._savedFileContent).toBe('new');
    expect(app._persistedFileContent).toBe('new');
    expect(persist.save).toHaveBeenCalledWith(app);
  });

  it('saves through an existing writable file handle without opening the picker', async () => {
    const handle = makeWritableHandle({ name: 'client.ivy', diskContent: 'old model' });
    const picker = vi.fn(async () => makeWritableHandle());
    const { app, persist, controls } = makeSaveApp({ fileHandle: handle, win: { showSaveFilePicker: picker } });

    const saved = await saveModel(app, persist, { win: { showSaveFilePicker: picker } });

    expect(saved).toBe(true);
    expect(picker).not.toHaveBeenCalled();
    expect(handle.writes).toEqual(['new model']);
    expect(app._persistedFileContent).toBe('new model');
    expect(app._savedFileContent).toBe('new model');
    expect(controls.lastStatus).toEqual({ message: 'Saved: client.ivy', kind: 'success' });
  });

  it('shows immediate save feedback while a dirty save is in progress', async () => {
    const closeGate = deferred();
    const writes = [];
    const handle = {
      name: 'client.ivy',
      async getFile() {
        return {
          name: 'client.ivy',
          async text() {
            return 'old model';
          },
        };
      },
      async createWritable() {
        return {
          async write(content) {
            writes.push(content);
          },
          async close() {
            await closeGate.promise;
          },
        };
      },
    };
    const { app, persist } = makeSaveApp({ fileHandle: handle });

    const savePromise = saveModel(app, persist);

    expect(app._showSaveProgress).toHaveBeenCalledWith('Saving...');
    expect(app.editorLabel).toBe('client.ivy [saving...]');
    expect(app.editorLabel).not.toContain('**');
    expect(app._hideSaveProgress).not.toHaveBeenCalled();

    closeGate.resolve();
    const saved = await savePromise;

    expect(saved).toBe(true);
    expect(writes).toEqual(['new model']);
    expect(app._hideSaveProgress).toHaveBeenCalled();
    expect(app.editorLabel).toBe('client.ivy [saved]');
  });

  it('restores a persisted handle before falling back to Save As', async () => {
    const handle = makeWritableHandle({ name: 'helloworld.ivy', diskContent: 'old content' });
    const persist = makePersist({
      loadFileHandle: vi.fn(async (state) => {
        expect(state).toEqual({
          sessionId: 's1',
          fileName: 'helloworld.ivy',
          filePath: 'helloworld.ivy',
        });
        return handle;
      }),
      saveFileHandle: vi.fn(async () => undefined),
    });
    const { app } = makeSaveApp({
      fileName: 'helloworld.ivy',
      savedContent: 'old content',
      editorContent: 'old content ',
      persist,
    });

    const saved = await saveModel(app, persist);

    expect(saved).toBe(true);
    expect(persist.loadFileHandle).toHaveBeenCalledTimes(1);
    expect(persist.saveFileHandle).toHaveBeenCalledTimes(1);
    expect(app.saveAs).not.toHaveBeenCalled();
    expect(app._fileHandle).toBe(handle);
    expect(handle.writes).toEqual(['old content ']);
  });

  it('shows the missing-handle notice while the Save As picker is open and hides it afterward', async () => {
    const handle = makeWritableHandle({ name: 'helloworld.ivy' });
    let pickerSawNotice = false;
    const win = {
      showSaveFilePicker: vi.fn(async () => {
        pickerSawNotice = true;
        return handle;
      }),
    };
    const persist = makePersist({
      saveFileHandle: vi.fn(async () => undefined),
      setFileName: vi.fn(),
    });
    const { app } = makeSaveApp({
      fileName: 'helloworld.ivy',
      savedContent: 'old content',
      editorContent: 'old content ',
      persist,
      win,
    });

    const saved = await saveModelAs(app, persist, { options: { explainMissingHandle: true }, win });

    expect(saved).toBe(true);
    expect(pickerSawNotice).toBe(true);
    expect(app.showSaveAsExplanationNotice.mock.invocationCallOrder[0]).toBeLessThan(
      win.showSaveFilePicker.mock.invocationCallOrder[0],
    );
    expect(app.hideSaveAsExplanationNotice.mock.invocationCallOrder[0]).toBeGreaterThan(
      win.showSaveFilePicker.mock.invocationCallOrder[0],
    );
    expect(handle.writes).toEqual(['old content ']);
    expect(persist.saveFileHandle).toHaveBeenCalledTimes(1);
    expect(persist.setFileName).toHaveBeenCalledWith('helloworld.ivy', 'helloworld.ivy');
  });

  it('keeps the Chrome-capable Ctrl-S path on Save As when no handle is available', async () => {
    const handle = makeWritableHandle({ name: 'client.ivy' });
    const win = { showSaveFilePicker: vi.fn(async () => handle) };
    const { app, persist } = makeSaveApp({ win });

    const saved = await saveModel(app, persist, { win });

    expect(saved).toBe(true);
    expect(win.showSaveFilePicker).toHaveBeenCalledTimes(1);
    expect(app.downloadTextFile).not.toHaveBeenCalled();
    expect(handle.writes).toEqual(['new model']);
    expect(app._fileHandle).toBe(handle);
  });

  it('leaves dirty state intact when the Save As picker is cancelled', async () => {
    const win = {
      showSaveFilePicker: vi.fn(async () => {
        throw abortError();
      }),
    };
    const { app, persist, controls } = makeSaveApp({
      fileName: 'draft.ivy',
      savedContent: 'old content',
      editorContent: 'edited content',
      win,
    });

    const saved = await saveModelAs(app, persist, { win });

    expect(saved).toBe(false);
    expect(app._editorDirty()).toBe(true);
    expect(app._savedFileContent).toBe('old content');
    expect(app._persistedFileContent).toBe('old content');
    expect(controls.lastStatus).toEqual({ message: 'Save cancelled', kind: undefined });
  });

  it('reports write failures and does not mark failed Save content as saved', async () => {
    const handle = makeWritableHandle({
      name: 'client.ivy',
      diskContent: 'old model',
      failWrite: new Error('disk full'),
    });
    const { app, persist, controls } = makeSaveApp({ fileHandle: handle });

    const saved = await saveModel(app, persist);

    expect(saved).toBe(false);
    expect(handle.writes).toEqual([]);
    expect(app._savedFileContent).toBe('old model');
    expect(app._persistedFileContent).toBe('old model');
    expect(controls.lastStatus).toEqual({ message: 'Save failed: disk full', kind: 'error' });
  });

  it('downloads the current buffer when Save As is unavailable', async () => {
    const { app, persist, controls } = makeSaveApp();

    const saved = await saveModelAs(app, persist, { win: {} });

    expect(saved).toBe(true);
    expect(app.downloadTextFile).toHaveBeenCalledWith('client.ivy', 'new model', 'text/plain');
    expect(app._editorDirty()).toBe(false);
    expect(app._savedFileContent).toBe('new model');
    expect(app._persistedFileContent).toBe('new model');
    expect(controls.lastStatus.kind).toBe('success');
    expect(controls.lastStatus.message).toContain('Downloaded edited copy: client.ivy');
  });

  it('uses a same-name download for Ctrl-S in browsers without writable file handles', async () => {
    const { app, persist, controls } = makeSaveApp();

    const saved = await saveModel(app, persist, { win: {} });

    expect(saved).toBe(true);
    expect(app.downloadTextFile).toHaveBeenCalledWith('client.ivy', 'new model', 'text/plain');
    expect(app._editorDirty()).toBe(false);
    expect(controls.lastStatus.message).toContain('choose the original file to overwrite');
  });
});
