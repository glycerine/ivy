import { vi } from 'vitest';

export class FakeAPI {
  constructor() {
    this.sessionId = 's1';
  }
}

export class FakeControls {
  constructor() {
    this.statuses = [];
    this.infos = [];
  }

  setStatus(message, kind) {
    this.lastStatus = { message, kind };
    this.statuses.push(this.lastStatus);
  }

  showInfo(shortInfo, longInfo) {
    this.lastInfo = { shortInfo, longInfo };
    this.infos.push(this.lastInfo);
  }
}

export class FakeGraph {}

export function makePersist(overrides = {}) {
  return {
    getSessionIdFromURL: vi.fn(() => 's1'),
    loadFileHandle: vi.fn(async () => null),
    saveFileHandle: vi.fn(async () => undefined),
    save: vi.fn(),
    setFileName: vi.fn(),
    ...overrides,
  };
}

export function makeWritableHandle({
  name = 'model.ivy',
  diskContent = '',
  failCreate = null,
  failWrite = null,
  failClose = null,
} = {}) {
  const writes = [];
  const handle = {
    name,
    writes,
    async getFile() {
      return new File([diskContent], name, { type: 'text/plain' });
    },
    async createWritable() {
      if (failCreate) throw failCreate;
      return {
        async write(content) {
          if (failWrite) throw failWrite;
          writes.push(content);
        },
        async close() {
          if (failClose) throw failClose;
        },
      };
    },
  };
  return handle;
}

export function installSaveDom() {
  document.body.innerHTML = [
    '<div id="save-as-explain-notice" style="display: none"></div>',
    '<div id="model-editor-label"></div>',
    '<button id="file-reopen-last"></button>',
    '<span id="loaded-file"></span>',
    '<select id="mode-select"><option value="concrete" selected>concrete</option></select>',
    '<tbody id="state-checkbox-body"></tbody>',
    '<div id="file-recent-list"></div>',
  ].join('');
}

export function makeEditor(content) {
  let value = content;
  return {
    getValue: vi.fn(() => value),
    setValue: vi.fn((next) => {
      value = next;
    }),
  };
}

export function makeApp(IvyApp, {
  fileName = 'model.ivy',
  filePath = fileName,
  savedContent = '',
  editorContent = savedContent,
  fileHandle = null,
} = {}) {
  const app = new IvyApp();
  app._persistedFileName = fileName;
  app._persistedFilePath = filePath;
  app._persistedFileContent = savedContent;
  app._savedFileContent = savedContent;
  app._fileHandle = fileHandle;
  app.cmEditor = makeEditor(editorContent);
  return app;
}
