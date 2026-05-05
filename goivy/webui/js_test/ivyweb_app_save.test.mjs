import { beforeEach, describe, expect, it, vi } from 'vitest';
import { loadIvyApp, loadIvyPersist } from './helpers/load_browser_scripts.mjs';
import {
  FakeAPI,
  FakeControls,
  FakeGraph,
  installSaveDom,
  makeApp,
  makePersist,
  makeWritableHandle,
} from './helpers/fakes.mjs';

function loadAppWithPersist(IvyPersist = makePersist()) {
  return {
    IvyPersist,
    IvyApp: loadIvyApp({
      IvyAPI: FakeAPI,
      IvyControls: FakeControls,
      IvyGraph: FakeGraph,
      IvyPersist,
    }),
  };
}

function abortError() {
  const err = new Error('The user aborted a request.');
  err.name = 'AbortError';
  return err;
}

beforeEach(() => {
  installSaveDom();
  window.location.hash = '';
  window.showSaveFilePicker = undefined;
  window.localStorage.clear();
});

describe('IvyApp Save and Save As', () => {
  it('saves through an existing writable file handle without opening the picker', async () => {
    const { IvyApp } = loadAppWithPersist();
    const handle = makeWritableHandle({ name: 'client.ivy', diskContent: 'old model' });
    const picker = vi.fn(async () => makeWritableHandle());
    window.showSaveFilePicker = picker;

    const app = makeApp(IvyApp, {
      fileName: 'client.ivy',
      savedContent: 'old model',
      editorContent: 'new model',
      fileHandle: handle,
    });

    const saved = await app.save();

    expect(saved).toBe(true);
    expect(picker).not.toHaveBeenCalled();
    expect(handle.writes).toEqual(['new model']);
    expect(app._persistedFileContent).toBe('new model');
    expect(app._savedFileContent).toBe('new model');
    expect(app.controls.lastStatus).toEqual({ message: 'Saved: client.ivy', kind: 'success' });
  });

  it('restores a persisted handle before falling back to Save As', async () => {
    const handle = makeWritableHandle({ name: 'helloworld.ivy', diskContent: 'old content' });
    const loadFileHandle = vi.fn(async (state) => {
      expect(state).toEqual({
        sessionId: 's1',
        fileName: 'helloworld.ivy',
        filePath: 'helloworld.ivy',
      });
      return handle;
    });
    const saveFileHandle = vi.fn(async () => undefined);
    const { IvyApp } = loadAppWithPersist(makePersist({ loadFileHandle, saveFileHandle }));

    const app = makeApp(IvyApp, {
      fileName: 'helloworld.ivy',
      savedContent: 'old content',
      editorContent: 'old content ',
    });
    app.saveAs = vi.fn(async () => false);

    const saved = await app.save();

    expect(saved).toBe(true);
    expect(loadFileHandle).toHaveBeenCalledTimes(1);
    expect(saveFileHandle).toHaveBeenCalledTimes(1);
    expect(app.saveAs).not.toHaveBeenCalled();
    expect(app._fileHandle).toBe(handle);
    expect(handle.writes).toEqual(['old content ']);
  });

  it('shows the missing-handle notice while Save As picker is open and hides it afterward', async () => {
    const handle = makeWritableHandle({ name: 'helloworld.ivy' });
    let pickerSawNotice = false;
    window.showSaveFilePicker = vi.fn(async () => {
      pickerSawNotice = document.getElementById('save-as-explain-notice').style.display === 'block';
      return handle;
    });
    const saveFileHandle = vi.fn(async () => undefined);
    const setFileName = vi.fn();
    const { IvyApp } = loadAppWithPersist(makePersist({ saveFileHandle, setFileName }));

    const app = makeApp(IvyApp, {
      fileName: 'helloworld.ivy',
      savedContent: 'old content',
      editorContent: 'old content ',
    });

    const saved = await app.saveAs({ explainMissingHandle: true });

    expect(saved).toBe(true);
    expect(pickerSawNotice).toBe(true);
    expect(document.getElementById('save-as-explain-notice').style.display).toBe('none');
    expect(handle.writes).toEqual(['old content ']);
    expect(saveFileHandle).toHaveBeenCalledTimes(1);
    expect(setFileName).toHaveBeenCalledWith('helloworld.ivy', 'helloworld.ivy');
  });

  it('leaves dirty state intact when the Save As picker is cancelled', async () => {
    window.showSaveFilePicker = vi.fn(async () => {
      throw abortError();
    });
    const { IvyApp } = loadAppWithPersist();
    const app = makeApp(IvyApp, {
      fileName: 'draft.ivy',
      savedContent: 'old content',
      editorContent: 'edited content',
    });

    const saved = await app.saveAs();

    expect(saved).toBe(false);
    expect(app._editorDirty()).toBe(true);
    expect(app._savedFileContent).toBe('old content');
    expect(app._persistedFileContent).toBe('old content');
    expect(app.controls.lastStatus).toEqual({ message: 'Save cancelled', kind: undefined });
  });

  it('reports write failures and does not mark failed Save content as saved', async () => {
    const handle = makeWritableHandle({
      name: 'client.ivy',
      diskContent: 'old model',
      failWrite: new Error('disk full'),
    });
    const { IvyApp } = loadAppWithPersist();
    const app = makeApp(IvyApp, {
      fileName: 'client.ivy',
      savedContent: 'old model',
      editorContent: 'new model',
      fileHandle: handle,
    });

    const saved = await app.save();

    expect(saved).toBe(false);
    expect(handle.writes).toEqual([]);
    expect(app._savedFileContent).toBe('old model');
    expect(app._persistedFileContent).toBe('old model');
    expect(app.controls.lastStatus).toEqual({ message: 'Save failed: disk full', kind: 'error' });
  });

  it('uses the explicit unsupported fallback when File System Access is unavailable', async () => {
    const { IvyApp } = loadAppWithPersist();
    const app = makeApp(IvyApp, {
      fileName: 'client.ivy',
      savedContent: 'old model',
      editorContent: 'new model',
    });

    const saved = await app.saveAs();

    expect(saved).toBe(false);
    expect(app._editorDirty()).toBe(true);
    expect(app.controls.lastStatus).toEqual({
      message: 'Save as... not supported in this browser — use Download instead',
      kind: 'error',
    });
  });
});

describe('IvyPersist recent file state', () => {
  it('round-trips recent session metadata through localStorage', () => {
    const IvyPersist = loadIvyPersist();
    window.location.hash = '#stable-session';
    const app = {
      api: { sessionId: 'server-session' },
      _persistedFileName: 'client.ivy',
      _persistedFilePath: '/tmp/ivy/client.ivy',
      _persistedFileContent: 'ivy content',
      selectedArgNode: 'node0',
      _edgeVisibility: { link: { all_to_all: true } },
      _labelVisibility: { semaphore: { node_maybe: true } },
      argGraph: null,
      conceptGraph: null,
    };

    IvyPersist.save(app);

    const state = IvyPersist.load();
    expect(state).toMatchObject({
      sessionId: 'stable-session',
      fileName: 'client.ivy',
      filePath: '/tmp/ivy/client.ivy',
      fileContent: 'ivy content',
      selectedArgNode: 'node0',
    });
    expect(window.localStorage.getItem('ivy_last_session')).toBe('stable-session');
    expect(JSON.parse(window.localStorage.getItem('ivy_sessions'))).toEqual(['stable-session']);
    expect(IvyPersist.listSessions()).toEqual([
      {
        id: 'stable-session',
        fileName: 'client.ivy',
        filePath: '/tmp/ivy/client.ivy',
        timestamp: state.timestamp,
      },
    ]);
  });
});
