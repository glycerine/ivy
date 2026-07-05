import { vi } from 'vitest';

export class FakeAPI {
  constructor(overrides = {}) {
    this.sessionId = 's1';
    Object.assign(this, overrides);
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

  hideContextMenu() {}
}

export class FakeGraph {
  constructor(containerId = 'graph') {
    this.containerId = containerId;
    this.update = vi.fn();
    this.resize = vi.fn();
    this.fit = vi.fn();
    this.healthCheck = vi.fn();
    this.highlightNode = vi.fn();
    this.onNodeClick = vi.fn();
    this.onNodeRightClick = vi.fn();
    this.onEdgeClick = vi.fn();
    this.onEdgeRightClick = vi.fn();
    this.onBackgroundClick = vi.fn();
  }
}

export function makeEditor(content = '') {
  let value = content;
  return {
    getValue: vi.fn(() => value),
    setValue: vi.fn((next) => {
      value = next;
    }),
    refresh: vi.fn(),
    setOption: vi.fn(),
    removeLineClass: vi.fn(),
    setCursor: vi.fn(),
    setSelection: vi.fn(),
    getLine: vi.fn(() => ''),
    addLineClass: vi.fn(() => 'line-handle'),
    scrollIntoView: vi.fn(),
    focus: vi.fn(),
  };
}

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
  return {
    name,
    writes,
    async getFile() {
      return {
        name,
        type: 'text/plain',
        async text() {
          return diskContent;
        },
      };
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
}

export function makeApp({
  fileName = 'model.ivy',
  filePath = fileName,
  savedContent = '',
  editorContent = savedContent,
  fileHandle = null,
  api = new FakeAPI(),
} = {}) {
  return {
    _persistedFileName: fileName,
    _persistedFilePath: filePath,
    _persistedFileContent: savedContent,
    _savedFileContent: savedContent,
    _fileHandle: fileHandle,
    api,
    cmEditor: makeEditor(editorContent),
  };
}
