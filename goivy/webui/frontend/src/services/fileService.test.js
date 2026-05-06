import { describe, expect, it, vi } from 'vitest';
import {
  downloadModelForUnsupportedSave,
  ensureFileHandleWritable,
  mergeDiskVersionIntoEditBuffer,
  rememberLastOpenFile,
  updateReopenLastFileButton,
} from './fileService.js';

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

    const bridge = {
      updateReopenLastFileButton: vi.fn(),
    };
    app._fileHandle = null;
    app._persistedFileName = '';
    updateReopenLastFileButton(app, { bridge });
    expect(bridge.updateReopenLastFileButton).toHaveBeenCalledWith(true, 'Re-open last file client.ivy');
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
});
