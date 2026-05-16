import { HostedGoIvyApiAdapter } from '../engines/index.ts';

export class IvyAPIShim extends HostedGoIvyApiAdapter {
  baseURL: string;

  constructor(baseURL = '') {
    super({ baseURL });
    this.baseURL = baseURL || '';
  }
}

export class IvyControlsShim {
  [key: string]: any;
  api: any;
  edgeToggles: Record<string, any>;
  labelToggles: Record<string, any>;
  _contextMenuVisible: boolean;

  constructor(api: any) {
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

  showContextMenu(x: number, y: number, actions: any[] = []) {
    const doc = globalThis.document;
    const menu = doc && doc.getElementById('context-menu');
    if (!menu) return;
    menu.innerHTML = '';
    for (const action of actions || []) {
      if (action.separator) {
        const sep = doc.createElement('div');
        sep.className = 'context-menu-separator';
        menu.appendChild(sep);
        continue;
      }
      if (action.header) {
        const header = doc.createElement('div');
        header.className = 'context-menu-header';
        header.textContent = action.header;
        menu.appendChild(header);
        continue;
      }
      const item = doc.createElement('div');
      item.className = 'context-menu-item';
      item.textContent = action.name || '';
      item.setAttribute('data-action-id', action.id || action.name || '');
      item.addEventListener('click', (event) => {
        event.stopPropagation();
        this.hideContextMenu();
        if (typeof action.callback === 'function') action.callback();
      });
      menu.appendChild(item);
    }
    menu.style.display = 'block';
    menu.style.left = `${Number(x) || 0}px`;
    menu.style.top = `${Number(y) || 0}px`;
    this._contextMenuVisible = true;
  }

  hideContextMenu() {
    const menu = globalThis.document && globalThis.document.getElementById('context-menu');
    if (menu) {
      menu.style.display = 'none';
      menu.innerHTML = '';
    }
    this._contextMenuVisible = false;
  }

  isContextMenuVisible() {
    return this._contextMenuVisible;
  }

  activeInfoElement() {
    const doc = globalThis.document;
    if (!doc) return null;
    const activeInfo = doc.querySelector('.sheet-content.active .info-panel [id^="info-content"]');
    return activeInfo || doc.getElementById('info-content');
  }

  showInfo(shortInfo: any, longInfo: any) {
    const info = this.activeInfoElement();
    if (!info) return;
    const lines: string[] = [];
    if (shortInfo) lines.push(shortInfo);
    if (longInfo) {
      if (Array.isArray(longInfo)) lines.push(...longInfo);
      else lines.push(longInfo);
    }
    const text = lines.join('\n');
    info.textContent = text || 'Select a node or edge to see details';
    info.setAttribute('data-ivy-details-kind', text ? 'selection' : 'placeholder');
  }

  clearInfo() {
    const info = this.activeInfoElement();
    if (info) {
      info.textContent = 'Select a node or edge to see details';
      info.setAttribute('data-ivy-details-kind', 'placeholder');
    }
  }

  setStatus(message: any, level = '') {
    const statusbar = globalThis.document && globalThis.document.getElementById('statusbar');
    const messageEl = statusbar && statusbar.querySelector('.status-message');
    if (statusbar) {
      statusbar.className = level || '';
    }
    if (messageEl) messageEl.textContent = message || '';
  }

  showLoading(message = 'Loading...') {
    const doc = globalThis.document;
    const overlay = doc && doc.getElementById('loading-overlay');
    const msg = doc && doc.getElementById('loading-message');
    if (msg) msg.textContent = message || 'Loading...';
    if (overlay) overlay.style.display = '';
  }

  hideLoading() {
    const overlay = globalThis.document && globalThis.document.getElementById('loading-overlay');
    if (overlay) overlay.style.display = 'none';
  }
}
