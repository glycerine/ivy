import { HostedGoIvyApiAdapter } from '../engines/index.js';

export class IvyAPIShim extends HostedGoIvyApiAdapter {
  constructor(baseURL = '') {
    super({ baseURL });
    this.baseURL = baseURL || '';
  }
}

export class IvyControlsShim {
  constructor(api) {
    this.api = api;
    this.edgeToggles = {};
    this.labelToggles = {};
    this._contextMenuVisible = false;
  }

  buildEdgeToggles() {
    this.edgeToggles = {};
  }

  buildLabelToggles() {
    this.labelToggles = {};
  }

  getEdgeToggleState() {
    return {};
  }

  getLabelToggleState() {
    return {};
  }

  showContextMenu(x, y, actions = []) {
    const bridge = globalThis.window && globalThis.window.__ivyVueBridge;
    if (bridge && typeof bridge.showContextMenu === 'function') {
      bridge.showContextMenu(x, y, actions);
      this._contextMenuVisible = true;
    }
  }

  hideContextMenu() {
    const bridge = globalThis.window && globalThis.window.__ivyVueBridge;
    if (bridge && typeof bridge.hideContextMenu === 'function') {
      bridge.hideContextMenu();
    }
    this._contextMenuVisible = false;
  }

  isContextMenuVisible() {
    return this._contextMenuVisible;
  }

  showInfo(shortInfo, longInfo) {
    const bridge = globalThis.window && globalThis.window.__ivyVueBridge;
    if (bridge && typeof bridge.updateDetails === 'function') {
      bridge.updateDetails({ shortInfo, longInfo });
    }
  }

  clearInfo() {
    const bridge = globalThis.window && globalThis.window.__ivyVueBridge;
    if (bridge && typeof bridge.clearDetails === 'function') {
      bridge.clearDetails();
    }
  }

  setStatus(message, level = '') {
    const bridge = globalThis.window && globalThis.window.__ivyVueBridge;
    if (bridge && typeof bridge.updateStatus === 'function') {
      bridge.updateStatus(message, level || '');
    }
  }

  showLoading(message = 'Loading...') {
    const bridge = globalThis.window && globalThis.window.__ivyVueBridge;
    if (bridge && typeof bridge.showLoading === 'function') {
      bridge.showLoading(message || 'Loading...');
    }
  }

  hideLoading() {
    const bridge = globalThis.window && globalThis.window.__ivyVueBridge;
    if (bridge && typeof bridge.hideLoading === 'function') {
      bridge.hideLoading();
    }
  }
}
